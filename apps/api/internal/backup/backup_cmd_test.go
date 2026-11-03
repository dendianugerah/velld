package backup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
)

func TestDumpCommandRejectsUnsupportedType(t *testing.T) {
	s := &BackupService{}
	conn := &connection.StoredConnection{Type: "cassandra"}

	if _, err := s.dumpCommand(conn, "out.sql"); err == nil {
		t.Fatal("expected error for unsupported database type")
	}
}

func TestDumpCommandBuildsEverySupportedType(t *testing.T) {
	supported := []string{"postgresql", "mysql", "mariadb", "mongodb", "redis"}

	for _, dbType := range supported {
		t.Run(dbType, func(t *testing.T) {
			s := &BackupService{}
			conn := &connection.StoredConnection{
				Type:         dbType,
				Host:         "localhost",
				Port:         5432,
				Username:     "user",
				DatabaseName: "app",
			}

			cmd, err := s.dumpCommand(conn, "out.sql")
			if err != nil {
				t.Skipf("client tool for %s not installed: %v", dbType, err)
			}
			defer cmd.Close()
			if cmd.cmd == nil {
				t.Fatal("expected a command")
			}
		})
	}
}

func TestDumpCommandsKeepPasswordOffArgv(t *testing.T) {
	s := &BackupService{}
	conn := &connection.StoredConnection{
		Host:         "localhost",
		Port:         5432,
		Username:     "user",
		Password:     "supersecret",
		DatabaseName: "app",
	}

	cases := map[string]func() (*preparedCommand, error){
		"postgresql": func() (*preparedCommand, error) { return s.createPgDumpCmd(conn, "out.sql") },
		"mysql":      func() (*preparedCommand, error) { return s.createMySQLDumpCmd(conn, "out.sql") },
		"redis":      func() (*preparedCommand, error) { return s.createRedisDumpCmd(conn, "out.sql") },
	}

	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			cmd, err := build()
			if err != nil {
				t.Skipf("client tool for %s not installed: %v", name, err)
			}
			defer cmd.Close()

			for _, arg := range cmd.cmd.Args {
				if arg == "supersecret" {
					t.Fatalf("%s passed the password on argv: %v", name, cmd.cmd.Args)
				}
			}
		})
	}
}

func TestMySQLOptionFileIsModeRestrictedAndCleanedUp(t *testing.T) {
	path, cleanup, err := writeMySQLOptionFile("p@ss")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("credential file not created: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %o", info.Mode().Perm())
	}

	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("credential file was not removed: %v", err)
	}
}

func TestRunDumpCmdReportsCombinedOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell command differs on windows")
	}

	cmd := exec.Command("sh", "-c", "echo oops >&2; exit 3")

	out, err := runDumpCmd(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected a non-zero exit error")
	}
	if string(out) != "oops\n" {
		t.Fatalf("expected stderr in output, got %q", string(out))
	}
}

func TestRunDumpCmdCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell command differs on windows")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := exec.Command("sh", "-c", "sleep 5")
	_, err := runDumpCmd(ctx, cmd)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
