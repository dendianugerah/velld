package backup

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestPG18Diagnostics(t *testing.T) {
	report := newPG18Report(VerificationFailed)
	output := `SET
psql:<stdin>:12: ERROR:  extension "postgis" is not available
DETAIL:  Could not open extension control file.
psql:<stdin>:13: ERROR:  unrecognized configuration parameter "old_setting"
psql:<stdin>:14: ERROR:  role "app" does not exist
psql:<stdin>:15: WARNING:  no privileges were granted for "items"
psql:<stdin>:16: NOTICE:  relation "items" already exists, skipping
psql:<stdin>:17: error: invalid command \something
psql:<stdin>:18: ERROR:  column "buffers_backend" does not exist
`
	if !report.addPostgresDiagnostics(output) {
		t.Fatal("restore errors must fail rehearsal even when psql exits zero")
	}
	want := []string{"extension/error", "feature/error", "role/error", "restore/warning", "restore/error", "feature/error"}
	if len(report.Issues) != len(want) {
		t.Fatalf("issues: %+v", report.Issues)
	}
	for i, issue := range report.Issues {
		if issue.Category+"/"+issue.Severity != want[i] {
			t.Errorf("issue %d: %+v, want %s", i, issue, want[i])
		}
	}
	report = newPG18Report(VerificationPassed)
	warnings := strings.Repeat("psql:<stdin>:1: WARNING: test warning\n", 101)
	if report.addPostgresDiagnostics(warnings) || len(report.Issues) != 100 || report.OmittedIssues != 1 {
		t.Fatalf("warning-only report: %d issues, %d omitted", len(report.Issues), report.OmittedIssues)
	}
}

func TestPG18ReportPersistence(t *testing.T) {
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql"})
	id := uuid.New()
	now := time.Now()
	if err := s.backupRepo.CreateBackup(&Backup{ID: id, ConnectionID: conn.ID, Status: "completed",
		Path: "shop.sql.gz", CreatedAt: now, UpdatedAt: now, StartedTime: now, CompletedTime: &now}); err != nil {
		t.Fatal(err)
	}
	initial, err := s.backupRepo.pg18Report(id.String())
	if err != nil || initial.Status != "not_run" {
		t.Fatalf("initial report: %+v, %v", initial, err)
	}
	pending := newPG18Report(VerificationPending)
	if queued, err := s.backupRepo.queuePG18Report(id.String(), pending); err != nil || !queued {
		t.Fatalf("first queue: %v, %v", queued, err)
	}
	if queued, err := s.backupRepo.queuePG18Report(id.String(), pending); err != nil || queued {
		t.Fatalf("duplicate queue: %v, %v", queued, err)
	}
	t.Setenv("DOCKER_HOST", "unix://"+t.TempDir()+"/missing.sock")
	s.restoreChecks.docker = newDockerEngine()
	if err := s.InitializePG18Rehearsals(); err != nil {
		t.Fatal(err)
	}
	report, err := s.backupRepo.pg18Report(id.String())
	if err != nil || report.Status != VerificationFailed || report.TargetVersion != "18" || !strings.Contains(report.Message, "restarted") {
		t.Fatalf("interrupted report: %+v, %v", report, err)
	}
	if queued, err := s.backupRepo.queuePG18Report(id.String(), pending); err != nil || !queued {
		t.Fatalf("rerun queue: %v, %v", queued, err)
	}
	req := httptest.NewRequest(http.MethodGet, "/backups/"+id.String()+"/pg18-rehearsal", nil)
	rec := httptest.NewRecorder()
	NewBackupHandler(s).PG18Rehearsal(rec, mux.SetURLVars(req, map[string]string{"id": id.String()}))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous report request status: %d", rec.Code)
	}
}
