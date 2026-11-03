//go:build integration

package backup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
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
	prepared, err := svc.createPgDumpCmd(conn, toFile(dumpPath))
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
	prepared, err := svc.createPgDumpCmd(conn, toFile(filepath.Join(dir, "dump.sql")))
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := prepared.Run(ctx); err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}

func TestPgDumpStreamsToWriter(t *testing.T) {
	pg := startTestPostgres(t)

	pg.exec(t, "CREATE TABLE streamed (id serial primary key, note text)")
	pg.exec(t, "INSERT INTO streamed (note) VALUES ('hello'), ('world')")

	conn := &connection.StoredConnection{
		Type:         "postgresql",
		Host:         pg.host,
		Port:         pg.port,
		Username:     pg.user,
		Password:     pg.password,
		DatabaseName: pg.database,
	}

	svc := &BackupService{}
	prepared, err := svc.createPgDumpCmd(conn, toStdout())
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}

	var sink bytes.Buffer
	cw := NewChecksumWriter(&sink)
	if err := prepared.RunToWriter(context.Background(), cw); err != nil {
		t.Fatalf("stream dump: %v", err)
	}

	if cw.Size() == 0 {
		t.Fatal("streamed zero bytes")
	}
	if !strings.Contains(sink.String(), "CREATE TABLE public.streamed") {
		t.Fatal("streamed output is missing the schema")
	}
	if !strings.Contains(sink.String(), "hello") {
		t.Fatal("streamed output is missing the rows")
	}
	if len(cw.Checksum()) != 64 {
		t.Fatalf("expected 64-char sha256 hex, got %q", cw.Checksum())
	}
}

func TestVerifyRestoreComparesRowCounts(t *testing.T) {
	pg := startTestPostgres(t)

	pg.exec(t, "CREATE TABLE accounts (id serial primary key, name text)")
	pg.exec(t, "INSERT INTO accounts (name) VALUES ('a'), ('b'), ('c')")

	conn := &connection.StoredConnection{
		Type:         "postgresql",
		Host:         pg.host,
		Port:         pg.port,
		Username:     pg.user,
		Password:     pg.password,
		DatabaseName: pg.database,
	}

	dir := t.TempDir()
	dumpPath := filepath.Join(dir, "dump.sql")

	svc := &BackupService{}
	prepared, err := svc.createPgDumpCmd(conn, toFile(dumpPath))
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}
	if out, err := prepared.Run(context.Background()); err != nil {
		t.Fatalf("dump: %v\n%s", err, out)
	}

	backup := &Backup{ID: uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")}

	matches, details, err := svc.verifyRestore(context.Background(), backup, conn, dumpPath)
	if err != nil {
		t.Fatalf("verify restore: %v", err)
	}
	if !matches {
		t.Fatalf("expected restore to match, details: %s", details)
	}

	out, err := svc.psqlQuery(context.Background(), conn, "postgres",
		"SELECT datname FROM pg_database WHERE datname LIKE 'velld_verify_%'")
	if err != nil {
		t.Fatalf("list databases: %v", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("scratch database was not dropped, found: %q", string(out))
	}
}
