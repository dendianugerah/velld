//go:build integration

package backup

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

func TestVerifyRestoreMySQLComparesRowCounts(t *testing.T) {
	db := startTestMySQL(t)

	db.exec(t, "CREATE TABLE accounts (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(64))")
	db.exec(t, "INSERT INTO accounts (name) VALUES ('a'), ('b'), ('c'), ('d')")

	conn := &connection.StoredConnection{
		Type:         "mysql",
		Host:         db.host,
		Port:         db.port,
		Username:     db.user,
		Password:     db.password,
		DatabaseName: db.database,
	}

	dir := t.TempDir()
	dumpPath := filepath.Join(dir, "dump.sql")

	svc := &BackupService{}
	prepared, err := svc.createMySQLDumpCmd(conn, toFile(dumpPath))
	if err != nil {
		t.Fatalf("build dump command: %v", err)
	}
	if out, err := prepared.Run(context.Background()); err != nil {
		t.Fatalf("dump: %v\n%s", err, out)
	}

	backup := &Backup{ID: uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")}

	matches, details, err := svc.verifyRestore(context.Background(), backup, conn, dumpPath)
	if err != nil {
		t.Fatalf("verify restore: %v", err)
	}
	if !matches {
		t.Fatalf("expected restore to match, details: %s", details)
	}
}
