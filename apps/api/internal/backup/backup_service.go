package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/notification"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type BackupService struct {
	connStorage      *connection.ConnectionRepository
	backupDir        string
	backupRepo       *BackupRepository
	cronManager      *cron.Cron
	cronEntries      map[string]cron.EntryID // map[scheduleID]entryID
	settingsService  *settings.SettingsService
	notificationRepo *notification.NotificationRepository
	cryptoService    *common.EncryptionService
}

func NewBackupService(
	connStorage *connection.ConnectionRepository,
	backupDir string,
	backupRepo *BackupRepository,
	settingsService *settings.SettingsService,
	notificationRepo *notification.NotificationRepository,
	cryptoService *common.EncryptionService,
) *BackupService {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		panic(err)
	}

	cronManager := cron.New(cron.WithSeconds())
	service := &BackupService{
		connStorage:      connStorage,
		backupDir:        backupDir,
		backupRepo:       backupRepo,
		settingsService:  settingsService,
		notificationRepo: notificationRepo,
		cryptoService:    cryptoService,
		cronManager:      cronManager,
		cronEntries:      make(map[string]cron.EntryID),
	}

	// Recover existing schedules before starting the cron manager
	if err := service.recoverSchedules(); err != nil {
		fmt.Printf("Error recovering schedules: %v\n", err)
	}

	cronManager.Start()
	return service
}

func (s *BackupService) recoverSchedules() error {
	schedules, err := s.backupRepo.GetAllActiveSchedules()
	if err != nil {
		return fmt.Errorf("failed to get active schedules: %v", err)
	}

	now := time.Now()
	for _, schedule := range schedules {
		scheduleID := schedule.ID.String()

		// Check if we missed any backups
		if schedule.NextRunTime != nil && schedule.NextRunTime.Before(now) {
			// Execute a backup immediately for missed schedule
			go s.executeCronBackup(schedule)
		}

		// Re-register the cron job
		entryID, err := s.cronManager.AddFunc(schedule.CronSchedule, func() {
			s.executeCronBackup(schedule)
		})
		if err != nil {
			fmt.Printf("Error re-registering schedule %s: %v\n", scheduleID, err)
			continue
		}

		s.cronEntries[scheduleID] = entryID
	}

	return nil
}

func (s *BackupService) CreateBackup(connectionID string) (*Backup, error) {
	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %v", err)
	}

	// Check if multi-database backup is needed
	if len(conn.SelectedDatabases) > 0 {
		// Create backups for all selected databases
		return s.createMultiDatabaseBackup(conn)
	}

	// Single database backup
	return s.createSingleDatabaseBackup(conn, conn.DatabaseName)
}

func (s *BackupService) createMultiDatabaseBackup(conn *connection.StoredConnection) (*Backup, error) {
	if err := s.verifyBackupTools(conn.Type); err != nil {
		return nil, err
	}

	tunnel, effectiveHost, effectivePort, err := s.setupSSHTunnelIfNeeded(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSH tunnel: %v", err)
	}
	if tunnel != nil {
		defer tunnel.Stop()
		conn.Host = effectiveHost
		conn.Port = effectivePort
	}

	connectionFolder := filepath.Join(s.backupDir, common.SanitizeConnectionName(conn.Name))
	if err := os.MkdirAll(connectionFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create connection backup folder: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	startTime := time.Now()

	var failedDatabases []string
	var successfulBackups []*Backup

	for _, dbName := range conn.SelectedDatabases {
		backupID := uuid.New()
		filename := fmt.Sprintf("%s_%s.sql", dbName, timestamp)
		backupPath := filepath.Join(connectionFolder, filename)

		tempConn := *conn
		tempConn.DatabaseName = dbName

		var cmd *exec.Cmd
		switch conn.Type {
		case "postgresql":
			cmd = s.createPgDumpCmd(&tempConn, backupPath)
		case "mysql", "mariadb":
			cmd = s.createMySQLDumpCmd(&tempConn, backupPath)
		case "mongodb":
			cmd = s.createMongoDumpCmd(&tempConn, backupPath)
		case "redis":
			cmd = s.createRedisDumpCmd(&tempConn, backupPath)
		default:
			return nil, fmt.Errorf("unsupported database type for backup: %s", conn.Type)
		}

		if cmd == nil {
			fmt.Printf("Warning: backup tool not found for database '%s'\n", dbName)
			failedDatabases = append(failedDatabases, dbName)
			continue
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Warning: Failed to backup database '%s': %s\n", dbName, string(output))
			failedDatabases = append(failedDatabases, dbName)
			continue
		}

		fileInfo, err := os.Stat(backupPath)
		if err != nil {
			fmt.Printf("Warning: Failed to get file info for database '%s': %v\n", dbName, err)
			failedDatabases = append(failedDatabases, dbName)
			continue
		}

		backup := &Backup{
			ID:           backupID,
			ConnectionID: conn.ID,
			StartedTime:  startTime,
			Status:       "completed",
			Path:         backupPath,
			Size:         fileInfo.Size(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		now := time.Now()
		backup.CompletedTime = &now

		if err := s.uploadToS3IfEnabled(backup, conn.UserID); err != nil {
			fmt.Printf("Warning: Failed to upload backup '%s' to S3: %v\n", dbName, err)
		}

		if err := s.backupRepo.CreateBackup(backup); err != nil {
			fmt.Printf("Warning: Failed to save backup record for '%s': %v\n", dbName, err)
			failedDatabases = append(failedDatabases, dbName)
			continue
		}

		successfulBackups = append(successfulBackups, backup)
	}

	if len(successfulBackups) == 0 {
		if len(failedDatabases) > 0 {
			return nil, fmt.Errorf("all database backups failed: %v", failedDatabases)
		}
		return nil, fmt.Errorf("all database backups failed")
	}

	if len(failedDatabases) > 0 {
		fmt.Printf("Multi-database backup completed with some failures: %d/%d databases backed up successfully, %d failed: %v\n",
			len(successfulBackups), len(conn.SelectedDatabases), len(failedDatabases), failedDatabases)
	} else {
		fmt.Printf("Multi-database backup completed: %d/%d databases backed up successfully\n",
			len(successfulBackups), len(conn.SelectedDatabases))
	}

	return successfulBackups[0], nil
}

func (s *BackupService) createSingleDatabaseBackup(conn *connection.StoredConnection, dbName string) (*Backup, error) {
	if err := s.verifyBackupTools(conn.Type); err != nil {
		return nil, err
	}

	// Setup SSH tunnel if enabled
	tunnel, effectiveHost, effectivePort, err := s.setupSSHTunnelIfNeeded(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSH tunnel: %v", err)
	}
	if tunnel != nil {
		defer tunnel.Stop()
		// Update connection to use tunnel
		conn.Host = effectiveHost
		conn.Port = effectivePort
	}

	backupID := uuid.New()
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql", dbName, timestamp)

	connectionFolder := filepath.Join(s.backupDir, common.SanitizeConnectionName(conn.Name))
	if err := os.MkdirAll(connectionFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create connection backup folder: %v", err)
	}

	backupPath := filepath.Join(connectionFolder, filename)

	backup := &Backup{
		ID:           backupID,
		ConnectionID: conn.ID,
		StartedTime:  time.Now(),
		Status:       "in_progress",
		Path:         backupPath,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	var cmd *exec.Cmd
	switch conn.Type {
	case "postgresql":
		cmd = s.createPgDumpCmd(conn, backupPath)
	case "mysql", "mariadb":
		cmd = s.createMySQLDumpCmd(conn, backupPath)
	case "mongodb":
		cmd = s.createMongoDumpCmd(conn, backupPath)
	case "redis":
		cmd = s.createRedisDumpCmd(conn, backupPath)
	default:
		return nil, fmt.Errorf("unsupported database type for backup: %s", conn.Type)
	}

	if cmd == nil {
		return nil, fmt.Errorf("backup tool not found for %s. Please ensure %s is installed and available in PATH", conn.Type, requiredTools[conn.Type])
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := string(output)
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return nil, fmt.Errorf("backup failed for %s database '%s' on %s:%d - %s",
			conn.Type, dbName, conn.Host, conn.Port, errorMsg)
	}

	// Get file size
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %v", err)
	}

	backup.Size = fileInfo.Size()
	backup.Status = "completed"
	now := time.Now()
	backup.CompletedTime = &now

	if err := s.uploadToS3IfEnabled(backup, conn.UserID); err != nil {
		fmt.Printf("Warning: Failed to upload backup to S3: %v\n", err)
	}

	if err := s.backupRepo.CreateBackup(backup); err != nil {
		return nil, fmt.Errorf("failed to save backup: %v", err)
	}

	return backup, nil
}

func (s *BackupService) GetBackup(id string) (*Backup, error) {
	return s.backupRepo.GetBackup(id)
}

func (s *BackupService) GetAllBackupsWithPagination(opts BackupListOptions) ([]*BackupList, int, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	return s.backupRepo.GetAllBackupsWithPagination(opts)
}

func (s *BackupService) GetBackupStats(userID uuid.UUID) (*BackupStats, error) {
	return s.backupRepo.GetBackupStats(userID)
}

func (s *BackupService) uploadToS3IfEnabled(backup *Backup, userID uuid.UUID) error {
	userSettings, err := s.settingsService.GetUserSettingsInternal(userID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	if !userSettings.S3Enabled {
		return nil
	}

	if userSettings.S3Endpoint == nil || *userSettings.S3Endpoint == "" {
		return fmt.Errorf("S3 endpoint not configured")
	}
	if userSettings.S3Bucket == nil || *userSettings.S3Bucket == "" {
		return fmt.Errorf("S3 bucket not configured")
	}
	if userSettings.S3AccessKey == nil || *userSettings.S3AccessKey == "" {
		return fmt.Errorf("S3 access key not configured")
	}
	if userSettings.S3SecretKey == nil || *userSettings.S3SecretKey == "" {
		return fmt.Errorf("S3 secret key not configured")
	}

	secretKey, err := s.cryptoService.Decrypt(*userSettings.S3SecretKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt S3 secret key: %w", err)
	}

	// (default to us-east-1 if not set)
	region := "us-east-1"
	if userSettings.S3Region != nil && *userSettings.S3Region != "" {
		region = *userSettings.S3Region
	}

	pathPrefix := ""
	if userSettings.S3PathPrefix != nil {
		pathPrefix = *userSettings.S3PathPrefix
	}

	s3Config := S3Config{
		Endpoint:   *userSettings.S3Endpoint,
		Region:     region,
		Bucket:     *userSettings.S3Bucket,
		AccessKey:  *userSettings.S3AccessKey,
		SecretKey:  secretKey,
		UseSSL:     userSettings.S3UseSSL,
		PathPrefix: pathPrefix,
	}

	s3Storage, err := NewS3Storage(s3Config)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage client: %w", err)
	}

	ctx := context.Background()
	objectKey, err := s3Storage.UploadFile(ctx, backup.Path)
	if err != nil {
		return fmt.Errorf("failed to upload backup to S3: %w", err)
	}

	backup.S3ObjectKey = &objectKey

	fmt.Printf("Successfully uploaded backup %s to S3: %s\n", backup.ID, objectKey)

	// Purge local backup file if enabled
	if userSettings.S3PurgeLocal {
		if err := os.Remove(backup.Path); err != nil {
			fmt.Printf("Warning: Failed to purge local backup file %s: %v\n", backup.Path, err)
		} else {
			fmt.Printf("Successfully purged local backup file: %s\n", backup.Path)
		}
	}

	return nil
}
