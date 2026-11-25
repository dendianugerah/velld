package backup

import (
	"encoding/json"
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

func (s *BackupService) setupSSHTunnelIfNeeded(conn *connection.StoredConnection) (*connection.SSHTunnel, string, int, error) {
	if !conn.SSHEnabled {
		return nil, conn.Host, conn.Port, nil
	}

	tunnel, err := connection.NewSSHTunnel(
		conn.SSHHost,
		conn.SSHPort,
		conn.SSHUsername,
		conn.SSHPassword,
		conn.SSHPrivateKey,
		conn.Host,
		conn.Port,
	)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to create SSH tunnel: %w", err)
	}

	if err := tunnel.Start(); err != nil {
		return nil, "", 0, fmt.Errorf("failed to start SSH tunnel: %w", err)
	}

	return tunnel, "127.0.0.1", tunnel.GetLocalPort(), nil
}

func (s *BackupService) createPgDumpCmd(conn *connection.StoredConnection) *exec.Cmd {
	binaryPath := s.findDatabaseBinaryPath("postgresql")
	if binaryPath == "" {
		fmt.Printf("ERROR: pg_dump binary not found. Please install PostgreSQL client tools.\n")
		return nil
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(requiredTools["postgresql"]))

	return postgresCmd(binPath, conn)
}

func postgresCmd(binPath string, conn *connection.StoredConnection, args ...string) *exec.Cmd {
	cmd := exec.Command(binPath, append([]string{
		"-w",
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-d", conn.DatabaseName,
	}, args...)...)

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

	args = append(args, "-r", outputPath, "--", conn.DatabaseName)

	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conn.Password)
	return cmd
}

func (s *BackupService) createMongoDumpCmd(conn *connection.StoredConnection, outputPath string) (*exec.Cmd, func(), error) {
	binaryPath := s.findDatabaseBinaryPath("mongodb")
	if binaryPath == "" {
		return nil, nil, fmt.Errorf("mongodump not found; install MongoDB Database Tools")
	}

	binPath := filepath.Join(binaryPath, common.GetPlatformExecutableName(requiredTools["mongodb"]))
	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--db", conn.DatabaseName,
		"--archive=" + outputPath,
	}

	if conn.Username != "" {
		args = append(args, "--username", conn.Username)
	}

	if conn.MongoAuthSource != "" {
		args = append(args, "--authenticationDatabase", conn.MongoAuthSource)
	}

	return mongoToolCmd(binPath, conn.Password, args)
}

func mongoToolCmd(binPath, password string, args []string) (*exec.Cmd, func(), error) {
	if password == "" {
		return exec.Command(binPath, args...), func() {}, nil
	}

	file, err := os.CreateTemp("", "velld-mongo-*.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("create MongoDB credentials file: %w", err)
	}
	cleanup := func() { _ = os.Remove(file.Name()) }
	encoded, _ := json.Marshal(password)
	_, writeErr := fmt.Fprintf(file, "password: %s\n", encoded)
	closeErr := file.Close()
	if writeErr != nil {
		cleanup()
		return nil, nil, fmt.Errorf("write MongoDB credentials file: %w", writeErr)
	}
	if closeErr != nil {
		cleanup()
		return nil, nil, fmt.Errorf("close MongoDB credentials file: %w", closeErr)
	}
	return exec.Command(binPath, append(args, "--config", file.Name())...), cleanup, nil
}
