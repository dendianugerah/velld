//go:build integration

package backup

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/google/uuid"
)

// TestDockerRestoreCheckFailureAlerts proves broken backups fail their restore check and alert.
func TestDockerRestoreCheckFailureAlerts(t *testing.T) {
	for _, tc := range []struct {
		dbType, extension string
		content           string
		wantMessage       string
	}{
		{"postgresql", ".sql.gz", "-- Dumped from database version 16.4\n" +
			"CREATE TABLE public.orders (id integer REFERENCES public.missing(id));\n" +
			"COPY public.orders (id) FROM stdin;\n1\n\\.\n", "Restore into postgres:16 failed"},
		{"mysql", ".sql", "-- Server version\t8.0.40\n" +
			"CREATE TABLE `orders` (id int, CONSTRAINT bad FOREIGN KEY (id) REFERENCES missing(id)) ENGINE=InnoDB;\n",
			"Restore into mysql:8.0 failed"},
		{"mongodb", ".archive", "this is not a mongodump archive", "Restore into mongo:7.0 failed"},
	} {
		t.Run(tc.dbType, func(t *testing.T) {
			alerts := make(chan map[string]any, 1)
			webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload map[string]any
				_ = json.NewDecoder(r.Body).Decode(&payload)
				alerts <- payload
			}))
			defer webhook.Close()

			s := newTestBackupService(t)
			conn := saveTestConnection(t, s, connection.StoredConnection{
				Type: tc.dbType, Host: "127.0.0.1", Port: 1, DatabaseName: "shop", Username: "velld",
			})
			enabled, url := true, webhook.URL
			if _, err := s.settingsService.UpdateUserSettings(conn.UserID,
				&settings.UpdateSettingsRequest{NotifyWebhook: &enabled, WebhookURL: &url}); err != nil {
				t.Fatal(err)
			}

			id := uuid.New()
			path := filepath.Join(s.backupDir, backupFilename(tc.dbType, "shop", "20260101_000000", id))
			writeRestoreCheckFixture(t, path, tc.content)
			now := time.Now()
			if err := s.backupRepo.CreateBackup(&Backup{ID: id, ConnectionID: conn.ID, Status: "completed",
				Path: path, Size: int64(len(tc.content)), StartedTime: now, CompletedTime: &now,
				CreatedAt: now, UpdatedAt: now}); err != nil {
				t.Fatal(err)
			}

			assertRestoreCheck(t, s, id.String(), VerificationFailed, tc.wantMessage)
			select {
			case alert := <-alerts:
				if alert["event"] != notificationRestoreCheckFailed || alert["database_name"] != "shop" {
					t.Fatalf("unexpected alert %+v", alert)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("no restore check failure alert was sent")
			}
		})
	}
}

func writeRestoreCheckFixture(t *testing.T, path, content string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if !strings.HasSuffix(path, ".gz") {
		_, err = file.WriteString(content)
	} else {
		writer := gzip.NewWriter(file)
		if _, err = writer.Write([]byte(content)); err == nil {
			err = writer.Close()
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}
