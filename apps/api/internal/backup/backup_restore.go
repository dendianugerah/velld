package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

type RestoreRequest struct {
	BackupID     string `json:"backup_id"`
	ConnectionID string `json:"connection_id"`
}

var restoreTools = map[string]string{
	"postgresql": "psql",
	"mysql":      "mysql",
	"mariadb":    "mysql",
	"mongodb":    "mongorestore",
}

// RestoreBackup restores a backup to a target database connection
func (s *BackupService) RestoreBackup(backupID string, connectionID string) error {
	backup, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup: %v", err)
	}

	if _, err := os.Stat(backup.Path); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backup.Path)
	}

	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}

	if err := s.verifyRestoreTools(conn.Type); err != nil {
		return err
	}

	tunnel, effectiveHost, effectivePort, err := s.setupSSHTunnelIfNeeded(conn)
	if err != nil {
		return fmt.Errorf("failed to setup SSH tunnel: %v", err)
	}
	if tunnel != nil {
		defer tunnel.Stop()
		conn.Host = effectiveHost
		conn.Port = effectivePort
	}

	var cmd *exec.Cmd
	switch conn.Type {
	case "postgresql":
		cmd = s.createPsqlRestoreCmd(conn, backup.Path)
	case "mysql", "mariadb":
		cmd = s.createMySQLRestoreCmd(conn, backup.Path)
	case "mongodb":
		cmd = s.createMongoRestoreCmd(conn, backup.Path)
	default:
		return fmt.Errorf("unsupported database type for restore: %s", conn.Type)
	}

	if cmd == nil {
		return fmt.Errorf("restore tool not found for %s. Please ensure %s is installed", conn.Type, restoreTools[conn.Type])
	}

	output, err := cmd.CombinedOutput()
	return s.validateRestoreOutput(conn.Type, conn.DatabaseName, output, err)
}

func (s *BackupService) validateRestoreOutput(dbType, dbName string, output []byte, cmdErr error) error {
	switch dbType {
	case "postgresql":
		return s.validatePostgreSQLRestore(output, cmdErr)
	case "mysql", "mariadb":
		return s.validateMySQLRestore(dbName, output, cmdErr)
	case "mongodb":
		return s.validateMongoDBRestore(dbName, output, cmdErr)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func (s *BackupService) validatePostgreSQLRestore(output []byte, cmdErr error) error {
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	var criticalErrors []string

	for _, line := range lines {
		if !strings.Contains(line, "ERROR:") {
			continue
		}

		if isCriticalPostgreSQLError(line) {
			criticalErrors = append(criticalErrors, line)
		}
	}

	if len(criticalErrors) > 0 {
		for _, errLine := range criticalErrors {
			if strings.Contains(errLine, "already exists") {
				return fmt.Errorf("restore failed: target database must be empty. See documentation for restore best practices")
			}
		}
		return fmt.Errorf("restore failed with %d error(s)", len(criticalErrors))
	}

	return nil
}

func (s *BackupService) validateMySQLRestore(dbName string, output []byte, cmdErr error) error {
	if cmdErr != nil {
		outputStr := string(output)
		if outputStr == "" {
			outputStr = cmdErr.Error()
		}
		return fmt.Errorf("restore failed for database '%s': %s", dbName, outputStr)
	}
	return nil
}

func (s *BackupService) validateMongoDBRestore(dbName string, output []byte, cmdErr error) error {
	if cmdErr != nil {
		outputStr := string(output)
		if outputStr == "" {
			outputStr = cmdErr.Error()
		}
		return fmt.Errorf("restore failed for database '%s': %s", dbName, outputStr)
	}
	return nil
}

func isCriticalPostgreSQLError(line string) bool {
	nonCriticalPatterns := []string{
		"WARNING:",
		"must be member of role",
		"no privileges",
		"NOTICE:",
	}

	for _, pattern := range nonCriticalPatterns {
		if strings.Contains(line, pattern) {
			return false
		}
	}

	return true
}

func (s *BackupService) verifyRestoreTools(dbType string) error {
	if _, exists := restoreTools[dbType]; !exists {
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
	return nil
}

func (s *BackupService) findDatabaseRestorePath(dbType string) string {
	if path := common.FindBinaryPath(dbType, restoreTools[dbType]); path != "" {
		return path
	}
	return ""
}

func (s *BackupService) createPsqlRestoreCmd(conn *connection.StoredConnection, backupPath string) *exec.Cmd {
	binaryPath := s.findDatabaseRestorePath("postgresql")
	if binaryPath == "" {
		fmt.Printf("ERROR: psql binary not found. Please install PostgreSQL client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools["postgresql"]))

	// Use -v ON_ERROR_STOP=1 to exit immediately on first error
	// This ensures errors are properly caught
	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-d", conn.DatabaseName,
		"-f", backupPath,
		"-v", "ON_ERROR_STOP=1", // Exit on first error
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", conn.Password))
	return cmd
}

func (s *BackupService) createMySQLRestoreCmd(conn *connection.StoredConnection, backupPath string) *exec.Cmd {
	binaryPath := s.findDatabaseRestorePath(conn.Type)
	if binaryPath == "" {
		fmt.Printf("ERROR: mysql binary not found. Please install MySQL/MariaDB client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools[conn.Type]))

	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-P", fmt.Sprintf("%d", conn.Port),
		"-u", conn.Username,
		fmt.Sprintf("-p%s", conn.Password),
		conn.DatabaseName,
	)

	file, err := os.Open(backupPath)
	if err != nil {
		fmt.Printf("ERROR: failed to open backup file: %v\n", err)
		return nil
	}

	cmd.Stdin = file
	return cmd
}

func (s *BackupService) createMongoRestoreCmd(conn *connection.StoredConnection, backupPath string) *exec.Cmd {
	binaryPath := s.findDatabaseRestorePath("mongodb")
	if binaryPath == "" {
		fmt.Printf("ERROR: mongorestore binary not found. Please install MongoDB Database Tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools["mongodb"]))

	backupDir := filepath.Dir(backupPath)

	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--db", conn.DatabaseName,
		backupDir,
	}

	if conn.Username != "" {
		args = append(args, "--username", conn.Username)
	}

	if conn.Password != "" {
		args = append(args, "--password", conn.Password)
	}

	return exec.Command(binPath, args...)
}
