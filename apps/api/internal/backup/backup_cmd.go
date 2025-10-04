package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

var requiredTools = map[string]string{
	"postgresql": "pg_dump",
	"mysql":      "mysqldump",
	"mariadb":    "mysqldump",
	"mongodb":    "mongodump",
}

func (s *BackupService) verifyBackupTools(dbType string) error {
	if _, exists := requiredTools[dbType]; !exists {
		return fmt.Errorf("unsupported database type: %s", dbType)
	}
	return nil
}

func (s *BackupService) findDatabaseBinaryPath(dbType string) string {
	if path := common.FindBinaryPath(dbType, requiredTools[dbType]); path != "" {
		return path
	}

	return ""
}

func (s *BackupService) createPgDumpCmd(conn *connection.StoredConnection, outputPath string) *exec.Cmd {
	binaryPath := s.findDatabaseBinaryPath("postgresql")
	if binaryPath == "" {
		fmt.Printf("ERROR: pg_dump binary not found. Please install PostgreSQL client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(requiredTools["postgresql"]))

	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-d", conn.DatabaseName,
		"-f", outputPath,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", conn.Password))
	return cmd
}

func (s *BackupService) createMySQLDumpCmd(conn *connection.StoredConnection, outputPath string) *exec.Cmd {
	binaryPath := s.findDatabaseBinaryPath(conn.Type)
	if binaryPath == "" {
		fmt.Printf("ERROR: mysqldump binary not found. Please install MySQL/MariaDB client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(requiredTools[conn.Type]))
	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-P", fmt.Sprintf("%d", conn.Port),
		"-u", conn.Username,
		fmt.Sprintf("-p%s", conn.Password),
		conn.DatabaseName,
		"-r", outputPath,
	)
	return cmd
}

func (s *BackupService) createMongoDumpCmd(conn *connection.StoredConnection, outputPath string) *exec.Cmd {
	binaryPath := s.findDatabaseBinaryPath("mongodb")
	if binaryPath == "" {
		fmt.Printf("ERROR: mongodump binary not found. Please install MongoDB Database Tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(requiredTools["mongodb"]))
	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--db", conn.DatabaseName,
		"--out", filepath.Dir(outputPath),
	}

	if conn.Username != "" {
		args = append(args, "--username", conn.Username)
	}

	if conn.Password != "" {
		args = append(args, "--password", conn.Password)
	}

	return exec.Command(binPath, args...)
}
