package backup

import (
	"context"
	"fmt"
	"io"
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

// RunToWriter streams the command's stdout into w and captures stderr for
// error reporting. It is the streaming counterpart of Run.
func (p *preparedCommand) RunToWriter(ctx context.Context, w io.Writer) error {
	defer p.Close()

	if ctx == nil {
		ctx = context.Background()
	}

	bound := exec.CommandContext(ctx, p.cmd.Path, p.cmd.Args[1:]...)
	bound.Env = p.cmd.Env
	bound.Dir = p.cmd.Dir
	bound.Stdout = w

	stderr, err := bound.StderrPipe()
	if err != nil {
		return fmt.Errorf("open stderr pipe: %w", err)
	}

	if err := bound.Start(); err != nil {
		return fmt.Errorf("start %s: %w", p.cmd.Path, err)
	}

	stderrBytes, _ := io.ReadAll(stderr)
	if err := bound.Wait(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("dump cancelled: %w", ctxErr)
		}
		if msg := trimStderr(stderrBytes); msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}

	return nil
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
func (s *BackupService) dumpCommand(conn *connection.StoredConnection, dst dumpDestination) (*preparedCommand, error) {
	if err := s.verifyBackupTools(conn.Type); err != nil {
		return nil, err
	}

	switch conn.Type {
	case "postgresql":
		return s.createPgDumpCmd(conn, dst)
	case "mysql", "mariadb":
		return s.createMySQLDumpCmd(conn, dst)
	case "mongodb":
		return s.createMongoDumpCmd(conn, dst)
	case "redis":
		return s.createRedisDumpCmd(conn, dst)
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

// dumpDestination says where a dump tool should write its output.
type dumpDestination struct {
	path   string
	stdout bool
}

func toFile(path string) dumpDestination   { return dumpDestination{path: path} }
func toStdout() dumpDestination            { return dumpDestination{stdout: true} }
func (d dumpDestination) isStdout() bool   { return d.stdout }
func (d dumpDestination) filePath() string { return d.path }

func (s *BackupService) createPgDumpCmd(conn *connection.StoredConnection, dst dumpDestination) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("postgresql")
	if err != nil {
		return nil, err
	}

	args := []string{
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-d", conn.DatabaseName,
	}
	if !dst.isStdout() {
		args = append(args, "-f", dst.filePath())
	}

	cmd := exec.Command(binPath, args...)

	// PGPASSWORD is passed through the environment rather than argv so the
	// password does not show up in the process list.
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", conn.Password))
	return &preparedCommand{cmd: cmd}, nil
}

func (s *BackupService) createMySQLDumpCmd(conn *connection.StoredConnection, dst dumpDestination) (*preparedCommand, error) {
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

	args = append(args, conn.DatabaseName)
	if !dst.isStdout() {
		args = append(args, "-r", dst.filePath())
	}

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

func (s *BackupService) createMongoDumpCmd(conn *connection.StoredConnection, dst dumpDestination) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("mongodb")
	if err != nil {
		return nil, err
	}

	args := []string{
		"--host", conn.Host,
		"--port", fmt.Sprintf("%d", conn.Port),
		"--db", conn.DatabaseName,
	}

	if dst.isStdout() {
		args = append(args, "--archive")
	} else {
		args = append(args, "--out", filepath.Dir(dst.filePath()))
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

func (s *BackupService) createRedisDumpCmd(conn *connection.StoredConnection, dst dumpDestination) (*preparedCommand, error) {
	binPath, err := s.resolveToolPath("redis")
	if err != nil {
		return nil, err
	}

	if dst.isStdout() {
		return nil, fmt.Errorf("redis dumps must be written to a file")
	}

	args := []string{
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
	}
	if conn.DatabaseName != "" {
		args = append(args, "-n", conn.DatabaseName)
	}
	args = append(args, "--rdb", dst.filePath())

	cmd := exec.Command(binPath, args...)
	if conn.Password != "" {
		// REDISCLI_AUTH keeps the password off the command line.
		cmd.Env = append(os.Environ(), fmt.Sprintf("REDISCLI_AUTH=%s", conn.Password))
	}

	return &preparedCommand{cmd: cmd}, nil
}

// trimStderr trims trailing newlines and caps the length so a tool's error
// output stays readable in a log or API response.
func trimStderr(b []byte) string {
	s := string(b)
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	const limit = 4096
	if len(s) > limit {
		s = s[:limit]
	}
	return s
}
