//go:build integration

package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
)

func TestPgDumpProducesRestorableSQL(t *testing.T) {
	pg := startTestPostgres(t)

	pg.exec(t, "CREATE TABLE users (id serial primary key, email text not null)")
	pg.exec(t, "INSERT INTO users (email) VALUES ('a@example.com'), ('b@example.com')")

	dir := t.TempDir()
	dumpPath := filepath.Join(dir, "dump.sql")

	conn := &connection.StoredConnection{
		Type:         "postgresql",
		Host:         pg.host,
		Port:         pg.port,
		Username:     pg.user,
		Password:     pg.password,
		DatabaseName: pg.database,
	}

	svc := &BackupService{}
	prepared, err := svc.createPgDumpCmd(conn, dumpPath)
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}

	output, err := prepared.Run(context.Background())
	if err != nil {
		t.Fatalf("dump failed: %v\n%s", err, output)
	}

	info, err := os.Stat(dumpPath)
	if err != nil {
		t.Fatalf("dump file missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("dump file is empty")
	}

	content, err := os.ReadFile(dumpPath)
	if err != nil {
		t.Fatalf("read dump: %v", err)
	}

	for _, want := range []string{"CREATE TABLE public.users", "a@example.com", "b@example.com"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("dump is missing %q", want)
		}
	}
}

func TestPgDumpRespectsCancellation(t *testing.T) {
	pg := startTestPostgres(t)

	pg.exec(t, "CREATE TABLE big (id bigserial primary key, payload text)")
	pg.exec(t, "INSERT INTO big (payload) SELECT repeat('x', 1024) FROM generate_series(1, 5000)")

	dir := t.TempDir()
	conn := &connection.StoredConnection{
		Type:         "postgresql",
		Host:         pg.host,
		Port:         pg.port,
		Username:     pg.user,
		Password:     pg.password,
		DatabaseName: pg.database,
	}

	svc := &BackupService{}
	prepared, err := svc.createPgDumpCmd(conn, filepath.Join(dir, "dump.sql"))
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := prepared.Run(ctx); err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}
