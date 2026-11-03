package backup

import (
	"context"
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
	"redis":      "redis-cli",
}

// preparedCommand is a dump command plus the resources (temporary credential
// files) it needs. Cleanup must be called once the command has finished,
// whether it succeeded or not.
type preparedCommand struct {
	cmd     *exec.Cmd
	cleanup func()
}

func (p *preparedCommand) Close() {
	if p != nil && p.cleanup != nil {
		p.cleanup()
	}
}

// Run executes the command with ctx and always releases its resources
// afterwards, so callers cannot forget to clean up a credential file.
func (p *preparedCommand) Run(ctx context.Context) ([]byte, error) {
	defer p.Close()
	return runDumpCmd(ctx, p.cmd)
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

func (s *BackupService) resolveToolPath(dbType string) (string, error) {
	dir := s.findDatabaseBinaryPath(dbType)
	if dir == "" {
		return "", fmt.Errorf("backup tool for %s not found. Please ensure %s is installed and available in PATH",
			dbType, requiredTools[dbType])
	}

	return filepath.Join(dir, common.GetPlatformExecutableName(requiredTools[dbType])), nil
}

// dumpCommand builds the dump command for a database type. It centralizes the
// dispatch so callers do not repeat the same switch for every backup path, and
// resolves the tool up front so a missing binary is reported before any work
// starts.
func (s *BackupService) dumpCommand(conn *connection.StoredConnection, outputPath string) (*preparedCommand, error) {
	if err := s.verifyBackupTools(conn.Type); err != nil {
		return nil, err
	}

	switch conn.Type {
	case "postgresql":
		return s.createPgDumpCmd(conn, outputPath)
	case "mysql", "mariadb":
		return s.createMySQLDumpCmd(conn, outputPath)
	case "mongodb":
		return s.createMongoDumpCmd(conn, outputPath)
	case "redis":
		return s.createRedisDumpCmd(conn, outputPath)
	}

	return nil, fmt.Errorf("unsupported database type: %s", conn.Type)
}

// runDumpCmd executes a dump command bound to ctx and returns its combined
// output. Rebinding the command to ctx lets an interrupted request or a
// shutdown cancel a long-running dump instead of leaking the child process.
func runDumpCmd(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	if ctx != nil {
		bound := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
		bound.Env = cmd.Env
		bound.Dir = cmd.Dir
		cmd = bound
	}

	return cmd.CombinedOutput()
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

func (s *BackupService) createPgDumpCmd(conn *connection.StoredConnection, outputPath string) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("postgresql")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-d", conn.DatabaseName,
		"-f", outputPath,
	)

	// PGPASSWORD is passed through the environment rather than argv so the
	// password does not show up in the process list.
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", conn.Password))
	return &preparedCommand{cmd: cmd}, nil
}

func (s *BackupService) createMySQLDumpCmd(conn *connection.StoredConnection, outputPath string) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath(conn.Type)
	if err != nil {
		return nil, err
	}

	args := []string{
		"-h", conn.Host,
		"-P", fmt.Sprintf("%d", conn.Port),
		"-u", conn.Username,
	}

	if conn.SSL {
		args = append(args, "--ssl-mode=REQUIRED")
	} else {
		args = append(args, "--skip-ssl")
	}

	args = append(args, conn.DatabaseName, "-r", outputPath)

	cmd := exec.Command(binPath, args...)

	if conn.Password != "" {
		optionFile, cleanup, err := writeMySQLOptionFile(conn.Password)
		if err != nil {
			return nil, err
		}

		// --defaults-extra-file is only honored when it is the first argument.
		cmd.Args = append([]string{binPath, "--defaults-extra-file=" + optionFile}, cmd.Args[1:]...)
		return &preparedCommand{cmd: cmd, cleanup: cleanup}, nil
	}

	return &preparedCommand{cmd: cmd}, nil
}

func (s *BackupService) createMongoDumpCmd(conn *connection.StoredConnection, outputPath string) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("mongodb")
	if err != nil {
		return nil, err
	}

	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--db", conn.DatabaseName,
		"--out", filepath.Dir(outputPath),
	}

	if conn.Username != "" {
		args = append(args, "--username", conn.Username)
	}

	tool := &preparedCommand{cmd: exec.Command(binPath, args...)}

	if conn.Password != "" {
		configFile, cleanup, err := writeMongoConfigFile(conn.Username, conn.Password)
		if err != nil {
			return nil, err
		}
		tool.cleanup = cleanup
		tool.cmd.Args = append(tool.cmd.Args, "--config", configFile)
	}

	return tool, nil
}

func (s *BackupService) createRedisDumpCmd(conn *connection.StoredConnection, outputPath string) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("redis")
	if err != nil {
		return nil, err
	}

	args := []string{
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
	}
	if conn.DatabaseName != "" {
		args = append(args, "-n", conn.DatabaseName)
	}
	args = append(args, "--rdb", outputPath)

	cmd := exec.Command(binPath, args...)
	if conn.Password != "" {
		// REDISCLI_AUTH keeps the password off the command line.
		cmd.Env = append(os.Environ(), fmt.Sprintf("REDISCLI_AUTH=%s", conn.Password))
	}

	return &preparedCommand{cmd: cmd}, nil
}
