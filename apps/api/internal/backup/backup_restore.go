package backup

import (
	"bytes"
	"context"
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

	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}

	// Ensure backup file is available (local or download from S3)
	filePath, isTemp, err := s.ensureBackupFileAvailable(backup, conn.UserID)
	if err != nil {
		return err
	}

	// Clean up temp file after restore if needed
	if isTemp {
		defer func() {
			if err := os.Remove(filePath); err != nil {
				fmt.Printf("Warning: Failed to remove temp file %s: %v\n", filePath, err)
			}
		}()
	}

	_, err = s.restoreFile(context.Background(), conn, filePath)
	return err
}

func (s *BackupService) restoreFile(ctx context.Context, conn *connection.StoredConnection, filePath string) ([]byte, error) {
	return s.executeRestore(ctx, conn, filePath, false)
}

func (s *BackupService) executeRestore(ctx context.Context, conn *connection.StoredConnection, filePath string, collectPostgresDiagnostics bool) ([]byte, error) {
	if err := s.verifyRestoreTools(conn.Type); err != nil {
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

	var cmd *exec.Cmd
	switch conn.Type {
	case "postgresql":
		cmd = s.createPsqlRestoreCmd(conn)
		if cmd != nil {
			if collectPostgresDiagnostics {
				cmd.Args = append(cmd.Args, "-q", "-v", "ON_ERROR_STOP=0")
				cmd.Env = append(cmd.Env, "LC_ALL=C", "PGOPTIONS=-c lc_messages=C")
			}
			input, closeInput, err := openPostgresBackup(filePath)
			if err != nil {
				return nil, fmt.Errorf("open PostgreSQL backup: %w", err)
			}
			defer closeInput()
			cmd.Stdin, err = postgresRestoreInput(input)
			if err != nil {
				return nil, fmt.Errorf("read PostgreSQL backup header: %w", err)
			}
		}
	case "mysql", "mariadb":
		cmd = s.createMySQLRestoreCmd(conn)
		if cmd != nil {
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("open MySQL backup: %w", err)
			}
			defer file.Close()
			cmd.Stdin = file
		}
	case "mongodb":
		var cleanup func()
		cmd, cleanup, err = s.createMongoRestoreCmd(conn, filePath)
		if err != nil {
			return nil, err
		}
		defer cleanup()
	default:
		return nil, fmt.Errorf("unsupported database type for restore: %s", conn.Type)
	}

	if cmd == nil {
		return nil, fmt.Errorf("restore tool not found for %s. Please ensure %s is installed", conn.Type, restoreTools[conn.Type])
	}

	restore := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	restore.Stdin, restore.Env = cmd.Stdin, cmd.Env
	if collectPostgresDiagnostics {
		var diagnostics bytes.Buffer
		restore.Stderr = &diagnostics
		err := restore.Run()
		return diagnostics.Bytes(), err
	}
	output, err := restore.CombinedOutput()
	return output, s.validateRestoreOutput(conn.Type, conn.DatabaseName, output, err)
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
				return fmt.Errorf("restore failed: target database must be empty. See documentation for restore best practices.\n\nError details:\n%s", errLine)
			}
		}
		return fmt.Errorf("restore failed with %d error(s):\n%s", len(criticalErrors), strings.Join(criticalErrors, "\n"))
	}

	if cmdErr != nil {
		return fmt.Errorf("restore failed: %s", cmdErr.Error())
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

func (s *BackupService) createPsqlRestoreCmd(conn *connection.StoredConnection) *exec.Cmd {
	binaryPath := s.findDatabaseRestorePath("postgresql")
	if binaryPath == "" {
		fmt.Printf("ERROR: psql binary not found. Please install PostgreSQL client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools["postgresql"]))

	return postgresCmd(binPath, conn, "-X", "-f", "-", "-v", "ON_ERROR_STOP=1")
}

func (s *BackupService) createMySQLRestoreCmd(conn *connection.StoredConnection) *exec.Cmd {
	binaryPath := s.findDatabaseRestorePath(conn.Type)
	if binaryPath == "" {
		fmt.Printf("ERROR: mysql binary not found. Please install MySQL/MariaDB client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools[conn.Type]))

	args := []string{
		"-h", conn.Host,
		"-P", fmt.Sprintf("%d", conn.Port),
		"-u", conn.Username,
	}

	if !conn.SSL {
		args = append(args, "--skip-ssl")
	} else {
		args = append(args, "--ssl-mode=REQUIRED")
	}

	args = append(args, conn.DatabaseName)

	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conn.Password)
	return cmd
}

func (s *BackupService) createMongoRestoreCmd(conn *connection.StoredConnection, backupPath string) (*exec.Cmd, func(), error) {
	binaryPath := s.findDatabaseRestorePath("mongodb")
	if binaryPath == "" {
		return nil, nil, fmt.Errorf("mongorestore not found; install MongoDB Database Tools")
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(restoreTools["mongodb"]))

	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--nsInclude", conn.DatabaseName + ".*",
		"--archive=" + backupPath,
	}

	if conn.Username != "" {
		args = append(args, "--username", conn.Username)
	}

	if conn.MongoAuthSource != "" {
		args = append(args, "--authenticationDatabase", conn.MongoAuthSource)
	}

	return mongoToolCmd(binPath, conn.Password, args)
}
