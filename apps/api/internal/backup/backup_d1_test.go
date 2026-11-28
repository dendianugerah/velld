package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

func TestD1RestoreCheckRecordedWithoutDocker(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix://"+filepath.Join(t.TempDir(), "missing.sock"))
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, connection.StoredConnection{
		Type: "d1", D1AccountID: "account", D1DatabaseID: "database", D1APIToken: "test-d1-token",
	})
	var encrypted string
	if err := s.backupRepo.db.QueryRow("SELECT d1_api_token FROM connections WHERE id = ?", conn.ID).Scan(&encrypted); err != nil {
		t.Fatal(err)
	}
	stored, err := s.connStorage.GetConnection(conn.ID)
	if err != nil {
		t.Fatal(err)
	}
	response, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "" || encrypted == conn.D1APIToken || stored.D1Credentials() != conn.D1Credentials() ||
		strings.Contains(string(response), conn.D1APIToken) {
		t.Fatal("D1 credentials did not round-trip encrypted with the token hidden")
	}

	for _, tc := range []struct{ name, dump, status, message string }{
		{"valid", "CREATE TABLE items (id INTEGER);\nINSERT INTO items VALUES (1),(2);", VerificationPassed, "1 tables, 2 rows restored"},
		{"invalid", "CREATE TABLE items (id INTEGER);\nINSERT INTO items VALUES (", VerificationFailed, "D1 SQL import failed"},
		{"external-file", "ATTACH DATABASE 'not-a-backup.db' AS external;", VerificationFailed, "not authorized"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, now := uuid.New(), time.Now()
			path := filepath.Join(s.backupDir, id.String()+".sql")
			if err := os.WriteFile(path, []byte(tc.dump), 0600); err != nil {
				t.Fatal(err)
			}
			if err := s.backupRepo.CreateBackup(&Backup{ID: id, ConnectionID: conn.ID, Status: "completed",
				Path: path, StartedTime: now, CompletedTime: &now, CreatedAt: now, UpdatedAt: now}); err != nil {
				t.Fatal(err)
			}
			assertRestoreCheck(t, s, id.String(), tc.status, tc.message)
		})
	}
}
