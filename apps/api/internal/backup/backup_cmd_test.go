package backup

import (
	"context"
	"errors"
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
			if cmd == nil {
				t.Fatal("expected a command")
			}
		})
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
