package backup

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", t.TempDir()+"/test.db")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	if err := goose.Up(db, "../database/migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return db
}

func TestBackupManifestRoundTrip(t *testing.T) {
	db := newTestDB(t)
	repo := NewBackupRepository(db)

	if _, err := db.Exec(`
		INSERT INTO connections (id, name, type, host, port, username, password, database_name, user_id)
		VALUES ('conn-1', 'primary', 'postgresql', 'localhost', 5432, 'u', 'p', 'app', 'user-1');
		INSERT INTO backups (id, connection_id, status, path)
		VALUES ('backup-1', 'conn-1', 'completed', '/tmp/dump.sql');
	`); err != nil {
		t.Fatalf("seed rows: %v", err)
	}

	want := &BackupManifest{
		BackupID:          "backup-1",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          "abc123",
		UncompressedSize:  1024,
		CompressedSize:    512,
		DurationMs:        42,
	}
	if err := repo.CreateBackupManifest(want); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	got, err := repo.GetBackupManifest("backup-1")
	if err != nil {
		t.Fatalf("get manifest: %v", err)
	}

	if got.Checksum != want.Checksum {
		t.Errorf("checksum = %q, want %q", got.Checksum, want.Checksum)
	}
	if got.UncompressedSize != want.UncompressedSize {
		t.Errorf("uncompressed size = %d, want %d", got.UncompressedSize, want.UncompressedSize)
	}
	if got.DurationMs != want.DurationMs {
		t.Errorf("duration = %d, want %d", got.DurationMs, want.DurationMs)
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}
}

func TestGetBackupManifestMissing(t *testing.T) {
	db := newTestDB(t)
	repo := NewBackupRepository(db)

	if _, err := repo.GetBackupManifest("does-not-exist"); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
