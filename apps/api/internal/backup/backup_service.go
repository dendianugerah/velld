package backup

import (
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
	settingsRepo     *settings.SettingsRepository
	notificationRepo *notification.NotificationRepository
	cryptoService    *common.EncryptionService
}

func NewBackupService(
	connStorage *connection.ConnectionRepository,
	backupDir string,
	backupRepo *BackupRepository,
	settingsRepo *settings.SettingsRepository,
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
		settingsRepo:     settingsRepo,
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

	if err := s.verifyBackupTools(conn.Type); err != nil {
		return nil, err
	}

	backupID := uuid.New()
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql", conn.DatabaseName, timestamp)
	backupPath := filepath.Join(s.backupDir, filename)

	backup := &Backup{
		ID:           backupID,
		ConnectionID: connectionID,
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
			conn.Type, conn.DatabaseName, conn.Host, conn.Port, errorMsg)
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
