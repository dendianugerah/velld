package backup

import (
	"context"
	"os/exec"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
)

func TestCreatePgBaseBackupCmdStreamsToStdout(t *testing.T) {
	if _, err := exec.LookPath("pg_basebackup"); err != nil {
		t.Skip("pg_basebackup not installed")
	}

	s := &BackupService{}
	conn := &connection.StoredConnection{
		Type:     "postgresql",
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "secret",
	}

	prepared, err := s.createPgBaseBackupCmd(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer prepared.Close()

	var hasFormatTar, hasStream, hasDashD bool
	for _, arg := range prepared.cmd.Args {
		switch arg {
		case "--format=tar":
			hasFormatTar = true
		case "--wal-method=stream":
			hasStream = true
		case "-":
			hasDashD = true
		case "secret":
			t.Fatalf("password must not be on argv: %v", prepared.cmd.Args)
		}
	}
	if !hasFormatTar || !hasStream || !hasDashD {
		t.Fatalf("missing required base backup flags: %v", prepared.cmd.Args)
	}
}

func TestCreateBaseBackupRejectsNonPostgres(t *testing.T) {
	svc, repo := newVerificationService(t)
	seedBackupWithType(t, repo, "cccccccc-0000-0000-0000-000000000001", "/tmp/x.sql", "mysql")

	if _, err := svc.CreateBaseBackup(context.Background(), "conn-1"); err == nil {
		t.Fatal("expected base backup to reject a non-PostgreSQL connection")
	}
}
