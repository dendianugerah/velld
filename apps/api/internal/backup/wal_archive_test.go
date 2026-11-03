package backup

import (
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
)

func TestWALObjectKey(t *testing.T) {
	got := walObjectKey("My DB", "000000010000000000000001")
	if got != "my_db/wal/000000010000000000000001" {
		t.Fatalf("got %q", got)
	}
}

func TestLastPathSegment(t *testing.T) {
	cases := map[string]string{
		"prod/wal/000000010000000000000001": "000000010000000000000001",
		"segment":                           "segment",
		"a/b/":                              "",
	}
	for input, want := range cases {
		if got := lastPathSegment(input); got != want {
			t.Fatalf("lastPathSegment(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildWALCommandConfig(t *testing.T) {
	conn := &connection.StoredConnection{ID: "conn-1", Name: "Prod DB"}
	cfg := buildWALCommandConfig(conn, "https://velld.example.com", "/mnt/velld-wal")

	if cfg.WALPrefix != "prod_db/wal" {
		t.Fatalf("prefix = %q", cfg.WALPrefix)
	}
	if !strings.Contains(cfg.ArchiveCommand, "velld wal-push") ||
		!strings.Contains(cfg.ArchiveCommand, "conn-1") ||
		!strings.Contains(cfg.ArchiveCommand, "https://velld.example.com") ||
		!strings.Contains(cfg.ArchiveCommand, "%p") {
		t.Fatalf("archive command missing pieces: %q", cfg.ArchiveCommand)
	}
	if !strings.Contains(cfg.RestoreCommand, "/mnt/velld-wal") || !strings.Contains(cfg.RestoreCommand, "%f") {
		t.Fatalf("restore command missing pieces: %q", cfg.RestoreCommand)
	}
}
