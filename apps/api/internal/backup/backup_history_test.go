package backup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestBackupHistoryFiltersBeforePagination(t *testing.T) {
	s := newTestBackupService(t)
	pg := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", DatabaseName: "ledger"})
	my := saveTestConnection(t, s, connection.StoredConnection{UserID: pg.UserID, Type: "mysql", DatabaseName: "warehouse"})
	other := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", DatabaseName: "ledger"})
	now := time.Now().UTC()
	add := func(conn connection.StoredConnection, status string, age time.Duration, verification string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		created := now.Add(-age).Format(time.RFC3339)
		_, err := s.backupRepo.db.Exec(`INSERT INTO backups (id, connection_id, status, path, size, started_time, created_at, updated_at) VALUES (?, ?, ?, ?, 128, ?, ?, ?)`, id, conn.ID, status, "/backups/snapshot.sql", created, created, created)
		if err != nil {
			t.Fatal(err)
		}
		if verification != "" {
			if err := s.backupRepo.UpdateBackupVerification(id.String(), verification, "check detail"); err != nil {
				t.Fatal(err)
			}
		}
		return id
	}
	// Fill more than one unfiltered page with unrelated, newer backups.
	for i := 0; i < 12; i++ {
		add(my, "completed", time.Duration(i)*time.Minute, "")
	}
	wanted := add(pg, "completed", 2*time.Hour, "failed")
	add(pg, "completed", 3*time.Hour, "passed")
	add(pg, "failed", 4*time.Hour, "")
	add(pg, "completed", 40*24*time.Hour, "passed")
	add(other, "completed", time.Hour, "failed")
	from := now.Add(-24 * time.Hour)
	for _, tc := range []struct {
		name  string
		opts  BackupListOptions
		total int
	}{
		{"all", BackupListOptions{}, 16},
		{"engine", BackupListOptions{DatabaseType: "postgresql"}, 4},
		{"status", BackupListOptions{Status: "failed"}, 1},
		{"date", BackupListOptions{DateFrom: &from}, 15},
		{"database search", BackupListOptions{Search: "LEDGER"}, 4},
		{"connection search", BackupListOptions{Search: pg.Name}, 4},
		{"filename search", BackupListOptions{Search: "snapshot"}, 16},
		{"connection", BackupListOptions{ConnectionID: pg.ID}, 4},
		{"check failed", BackupListOptions{Verification: "failed"}, 1},
		{"not checked includes old backups", BackupListOptions{Verification: "skipped"}, 13},
		{"combined", BackupListOptions{Status: "completed", DatabaseType: "postgresql", DateFrom: &from, Verification: "failed", Search: "ledger", ConnectionID: pg.ID}, 1},
		{"no match", BackupListOptions{Search: "not-present"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts.UserID = pg.UserID
			tc.opts.Limit = 1
			for offset := 0; offset <= tc.total; offset++ {
				tc.opts.Offset = offset
				rows, total, err := s.GetAllBackupsWithPagination(tc.opts)
				if err != nil || total != tc.total {
					t.Fatalf("total=%d want=%d err=%v", total, tc.total, err)
				}
				wantLen := 1
				if offset == tc.total {
					wantLen = 0
				}
				if len(rows) != wantLen {
					t.Fatalf("offset=%d got %d rows, want %d", offset, len(rows), wantLen)
				}
				if len(rows) == 1 && rows[0].ConnectionName == "" {
					t.Fatal("missing connection name")
				}
				if tc.name == "combined" && len(rows) == 1 && rows[0].ID != wanted {
					t.Fatal("combined filters did not find the older matching backup")
				}
			}
		})
	}
}

func TestBackupDetailOwnershipAndHistoryHTTP(t *testing.T) {
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", DatabaseName: "ledger"})
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.backupRepo.db.Exec(`INSERT INTO backups (id, connection_id, status, path, size, created_at, updated_at) VALUES (?, ?, 'completed', 'ledger.sql', 12, ?, ?)`, id, conn.ID, now, now); err != nil {
		t.Fatal(err)
	}
	h := NewBackupHandler(s)
	for _, tc := range []struct {
		name string
		user uuid.UUID
		want int
	}{
		{"owner", conn.UserID, http.StatusOK}, {"another user", uuid.New(), http.StatusNotFound}, {"anonymous", uuid.Nil, http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := mux.SetURLVars(httptest.NewRequest("GET", "/api/backups/"+id, nil), map[string]string{"id": id})
			if tc.user != uuid.Nil {
				req = req.WithContext(context.WithValue(req.Context(), "user", jwt.MapClaims{"user_id": tc.user.String()}))
			}
			rec := httptest.NewRecorder()
			h.GetBackup(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status %d want %d: %s", rec.Code, tc.want, rec.Body)
			}
			if tc.want == http.StatusOK {
				var body struct {
					Data BackupList `json:"data"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Data.DatabaseName != "ledger" || body.Data.ConnectionName != conn.Name {
					t.Fatalf("missing detail metadata: %s (%v)", rec.Body, err)
				}
			}
		})
	}
	for _, tc := range []struct {
		query string
		want  int
		total int
	}{
		{"?status=completed&engine=postgresql&date=24hours&connection_id=" + conn.ID, 200, 1},
		{"?status=failed", 200, 0}, {"?date=invalid", 400, 0}, {"?verification=skipped", 200, 1},
	} {
		req := httptest.NewRequest("GET", "/api/backups"+tc.query, nil)
		req = req.WithContext(context.WithValue(req.Context(), "user", jwt.MapClaims{"user_id": conn.UserID.String()}))
		rec := httptest.NewRecorder()
		h.ListBackups(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s status %d want %d: %s", tc.query, rec.Code, tc.want, rec.Body)
		}
		if rec.Code == 200 {
			var body struct {
				Pagination struct {
					Total int `json:"total"`
				} `json:"pagination"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Pagination.Total != tc.total {
				t.Fatalf("%s wrong pagination: %s (%v)", tc.query, rec.Body, err)
			}
		}
	}
}
