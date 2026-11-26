package backup

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/settings"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

func newTestBackupService(t *testing.T) *BackupService {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "velld.db")+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../database/migrations"); err != nil {
		t.Fatal(err)
	}
	crypto, err := common.NewEncryptionService(strings.Repeat("42", 32))
	if err != nil {
		t.Fatal(err)
	}
	s := NewBackupService(connection.NewConnectionRepository(db, crypto), t.TempDir(),
		NewBackupRepository(db), settings.NewSettingsService(settings.NewSettingsRepository(db), crypto), crypto)
	s.restoreChecks.automatic = false
	t.Cleanup(func() { <-s.cronManager.Stop().Done() })
	return s
}

func saveTestConnection(t *testing.T, s *BackupService, conn connection.StoredConnection) connection.StoredConnection {
	t.Helper()
	if conn.UserID == uuid.Nil {
		conn.UserID = uuid.New()
		if _, err := s.backupRepo.db.Exec("INSERT INTO users (id, username, password) VALUES (?, ?, ?)", conn.UserID, "test-user", "unused"); err != nil {
			t.Fatal(err)
		}
	}
	conn.ID = uuid.NewString()
	conn.Name = conn.ID
	conn.Status = "connected"
	conn.CreatedAt = time.Now().Format(time.RFC3339)
	conn.UpdatedAt = conn.CreatedAt
	if err := s.connStorage.Save(conn); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestRetentionCountSchedule(t *testing.T) {
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", Host: "localhost"})
	handler := NewBackupHandler(s)
	post := func(policy string, status int) {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/backups/schedule", strings.NewReader(
			`{"connection_id":"`+conn.ID+`","cron_schedule":"0 0 0 1 1 *",`+policy+`}`))
		handler.ScheduleBackup(rec, req)
		if rec.Code != status {
			t.Fatalf("schedule response = %d, want %d: %s", rec.Code, status, rec.Body)
		}
	}
	post(`"retention_count":-1,"retention_days":30`, http.StatusBadRequest)
	post(`"retention_count":0`, http.StatusBadRequest)
	post(`"retention_count":2`, http.StatusOK)

	assertPolicy := func(days, count int) {
		t.Helper()
		stored, err := s.backupRepo.GetBackupSchedule(conn.ID)
		if err != nil || stored.RetentionDays != days || stored.RetentionCount != count {
			t.Fatalf("schedule = %+v, error %v; want days=%d, count=%d", stored, err, days, count)
		}
		active, err := s.backupRepo.GetAllActiveSchedules()
		if err != nil || len(active) != 1 || active[0].RetentionCount != count {
			t.Fatalf("active schedules = %+v, error %v; want count=%d", active, err, count)
		}
		connections, err := s.connStorage.ListByUserID(conn.UserID)
		if err != nil || len(connections) != 1 || connections[0].RetentionCount == nil || *connections[0].RetentionCount != count {
			t.Fatalf("connection list = %+v, error %v; want count=%d", connections, err, count)
		}
	}
	assertPolicy(0, 2)
	post(`"retention_days":30`, http.StatusOK)
	assertPolicy(30, 2)
	zero := 0
	if err := s.UpdateBackupSchedule(conn.ID, &UpdateScheduleRequest{
		CronSchedule: "0 0 0 1 1 *", RetentionDays: 7, RetentionCount: &zero,
	}); err != nil {
		t.Fatal(err)
	}
	assertPolicy(7, 0)
}

func TestRetentionDeletesOnlyCompletedBackupsOutsidePolicy(t *testing.T) {
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", Host: "localhost", S3CleanupOnRetention: true})
	other := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", Host: "localhost", UserID: conn.UserID})
	now := time.Now().UTC()
	add := func(c connection.StoredConnection, status string, age time.Duration, sharedPath string) *Backup {
		t.Helper()
		id := uuid.New()
		point := now.Add(-age)
		path := sharedPath
		if path == "" {
			path = filepath.Join(s.backupDir, id.String())
			if err := os.WriteFile(path, []byte("backup"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		b := &Backup{ID: id, ConnectionID: c.ID, Status: status, Path: path,
			StartedTime: point, CreatedAt: point, UpdatedAt: point, CompletedTime: &point}
		if err := s.backupRepo.CreateBackup(b); err != nil {
			t.Fatal(err)
		}
		return b
	}
	newest := add(conn, "completed", time.Hour, "")
	oldButRankedSecond := add(conn, "completed", 72*time.Hour, "")
	old := add(conn, "completed", 96*time.Hour, "")
	duplicate := add(conn, "completed", 97*time.Hour, newest.Path)
	failed := add(conn, "failed", 0, "")
	otherBackup := add(other, "completed", 96*time.Hour, "")

	deleted := make(chan string, 10)
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			deleted <- r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected S3 request", http.StatusBadRequest)
		}
	}))
	defer remote.Close()
	enabled, insecure := true, false
	endpoint, bucket, key, secret := strings.TrimPrefix(remote.URL, "http://"), "test-backups", "test-key", "test-secret"
	if _, err := s.settingsService.UpdateUserSettings(conn.UserID, &settings.UpdateSettingsRequest{
		S3Enabled: &enabled, S3UseSSL: &insecure, S3Endpoint: &endpoint, S3Bucket: &bucket, S3AccessKey: &key, S3SecretKey: &secret,
	}); err != nil {
		t.Fatal(err)
	}
	for _, b := range []*Backup{newest, old, duplicate} {
		if err := s.backupRepo.UpdateBackupS3ObjectKey(b.ID.String(), filepath.Base(b.Path)); err != nil {
			t.Fatal(err)
		}
	}
	check := func(b *Backup, hasRecord, hasFile bool) {
		t.Helper()
		if _, err := s.backupRepo.GetBackup(b.ID.String()); (hasRecord && err != nil) || (!hasRecord && !errors.Is(err, sql.ErrNoRows)) {
			t.Fatalf("backup %s: GetBackup error = %v, want present=%t", b.ID, err, hasRecord)
		}
		if _, err := os.Stat(b.Path); (hasFile && err != nil) || (!hasFile && !errors.Is(err, os.ErrNotExist)) {
			t.Fatalf("backup %s: Stat error = %v, want file present=%t", b.ID, err, hasFile)
		}
	}
	s.cleanupOldBackups(conn.ID, 0, 2)
	check(old, false, false)
	check(duplicate, false, true)
	check(newest, true, true)
	check(oldButRankedSecond, true, true)
	check(failed, true, true)
	check(otherBackup, true, true)
	select {
	case path := <-deleted:
		if path != "/"+bucket+"/"+filepath.Base(old.Path) {
			t.Fatalf("deleted unexpected remote object %s", path)
		}
	default:
		t.Fatal("old remote object was not deleted")
	}
	select {
	case path := <-deleted:
		t.Fatalf("deleted unexpected additional remote object %s", path)
	default:
	}

	s.cleanupOldBackups(conn.ID, 2, 2)
	check(oldButRankedSecond, false, false)
	check(newest, true, true)
	check(failed, true, true)
	check(otherBackup, true, true)
}
