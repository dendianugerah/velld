package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	pg18Version = "18"
	pg18Label   = "velld.pg18-rehearsal"
)

type pg18Issue struct {
	Category string `json:"category"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type pg18Report struct {
	TargetVersion string      `json:"target_version"`
	SourceVersion string      `json:"source_version,omitempty"`
	Status        string      `json:"status"`
	Message       string      `json:"message,omitempty"`
	DurationMS    int64       `json:"duration_ms"`
	Issues        []pg18Issue `json:"issues"`
	OmittedIssues int         `json:"omitted_issues,omitempty"`
}

func newPG18Report(status string) pg18Report {
	return pg18Report{TargetVersion: pg18Version, Status: status, Issues: []pg18Issue{}}
}

func (r *BackupRepository) queuePG18Report(id string, report pg18Report) (bool, error) {
	encoded, err := json.Marshal(report)
	if err != nil {
		return false, err
	}
	result, err := r.db.Exec(`UPDATE backups SET pg18_status = ?, pg18_report = ?
		WHERE id = ? AND (pg18_status IS NULL OR pg18_status <> 'pending')`, report.Status, string(encoded), id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func (r *BackupRepository) pg18Report(id string) (pg18Report, error) {
	var encoded string
	err := r.db.QueryRow("SELECT pg18_report FROM backups WHERE id = ? AND pg18_report IS NOT NULL", id).Scan(&encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return newPG18Report("not_run"), nil
	}
	var report pg18Report
	if err != nil {
		return report, err
	}
	err = json.Unmarshal([]byte(encoded), &report)
	return report, err
}

func (s *BackupService) InitializePG18Rehearsals() error {
	report := newPG18Report(VerificationFailed)
	report.Message = "The API restarted before the rehearsal finished. Run it again to get a report."
	encoded, err := json.Marshal(report)
	if err != nil {
		return err
	}
	if _, err := s.backupRepo.db.Exec("UPDATE backups SET pg18_status = ?, pg18_report = ? WHERE pg18_status = 'pending'",
		report.Status, string(encoded)); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.restoreChecks.docker.removeContainersBefore(ctx, pg18Label, time.Now())
	return nil
}

func (h *BackupHandler) PG18Rehearsal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	s := h.backupService
	id := mux.Vars(r)["id"]
	var engine, status string
	err = s.backupRepo.db.QueryRow(`SELECT c.type, b.status FROM backups b
		JOIN connections c ON c.id = b.connection_id WHERE b.id = ? AND c.user_id = ?`, id, userID).Scan(&engine, &status)
	if errors.Is(err, sql.ErrNoRows) {
		response.SendError(w, http.StatusNotFound, "Backup not found")
		return
	}
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Could not look up the backup")
		return
	}
	if engine != "postgresql" || status != "completed" {
		response.SendError(w, http.StatusBadRequest, "Only completed PostgreSQL backups can be tested on PG 18")
		return
	}

	if r.Method == http.MethodPost {
		if err := s.startPG18Rehearsal(id, userID); err != nil {
			response.SendError(w, http.StatusInternalServerError, "Could not start the PG 18 rehearsal")
			return
		}
	}
	report, err := s.backupRepo.pg18Report(id)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Could not read the PG 18 rehearsal report")
		return
	}
	response.SendSuccess(w, "PG 18 rehearsal report", report)
}

func (s *BackupService) startPG18Rehearsal(id string, userID uuid.UUID) error {
	report := newPG18Report(VerificationPending)
	queued, err := s.backupRepo.queuePG18Report(id, report)
	if err != nil || !queued {
		return err
	}
	go func() {
		start := time.Now()
		backup, err := s.backupRepo.GetBackup(id)
		if err == nil {
			s.runPG18Rehearsal(backup, userID, &report)
		} else {
			report.Status = VerificationFailed
			report.Message = "The backup is no longer available."
		}
		report.DurationMS = time.Since(start).Milliseconds()
		encoded, err := json.Marshal(report)
		if err == nil {
			_, err = s.backupRepo.db.Exec("UPDATE backups SET pg18_status = ?, pg18_report = ? WHERE id = ?",
				report.Status, string(encoded), id)
		}
		if err != nil {
			log.Printf("Could not save PG 18 rehearsal report for backup %s: %v", id, err)
		}
	}()
	return nil
}

func (s *BackupService) runPG18Rehearsal(backup *Backup, userID uuid.UUID, report *pg18Report) {
	report.Status = VerificationFailed
	ctx, cancel := context.WithTimeout(context.Background(), restoreCheckTimeout)
	defer cancel()
	select {
	case s.restoreChecks.slot <- struct{}{}:
		defer func() { <-s.restoreChecks.slot }()
	case <-ctx.Done():
		report.Message = "Timed out waiting for an available restore check slot."
		return
	}

	docker := s.restoreChecks.docker
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err := docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		report.Message = "Docker is not reachable from the API; mount /var/run/docker.sock to run the rehearsal."
		return
	}
	file, temporary, err := s.ensureBackupFileAvailable(backup, userID)
	if err != nil {
		report.Message = fmt.Sprintf("Backup file is unavailable: %v", err)
		return
	}
	if temporary {
		defer os.Remove(file)
	}
	dump, err := inspectDump("postgresql", file)
	if err != nil {
		report.Message = fmt.Sprintf("Backup file is unreadable: %v", err)
		return
	}
	report.SourceVersion = dump.serverVersion
	if dump.serverVersion == "" {
		report.Message = "The PostgreSQL dump is missing its source version header."
		return
	}
	target := restoreCheckTargetFor("postgresql", pg18Version, "velld_rehearsal")
	container, err := docker.startConfigured(ctx, target.image, target.port, target.env, false,
		map[string]any{"Labels": map[string]string{pg18Label: "true"}})
	if err != nil {
		report.Message = fmt.Sprintf("Could not start %s: %v", target.image, err)
		return
	}
	defer docker.remove(container.id)
	target.conn.Host, target.conn.Port = container.host, container.port
	if err := prepareRestoreTarget(ctx, &target, dump); err != nil {
		report.Message = err.Error()
		if strings.Contains(report.Message, "create role ") {
			report.Issues = append(report.Issues, pg18Issue{"role", "error", report.Message})
		}
		return
	}

	output, err := s.executeRestore(ctx, &target.conn, file, true)
	failed := report.addPostgresDiagnostics(string(output))
	if err != nil {
		report.Message = fmt.Sprintf("Restore could not finish: %v", err)
		return
	}
	if failed {
		report.Message = "PostgreSQL reported errors while restoring this backup."
		return
	}
	summary, err := confirmRestoredData(ctx, &target.conn, dump, output)
	if err != nil {
		report.Message = fmt.Sprintf("Restored data could not be verified: %v", err)
		return
	}
	report.Status = VerificationPassed
	report.Message = "Restored " + summary + " into " + target.image + "."
}

var pg18Diagnostic = regexp.MustCompile(`(?i)^(?:psql:(?:[^:]*:\d+:)?\s*)?(ERROR|FATAL|PANIC|WARNING|NOTICE|INFO|LOG|DEBUG):\s*(.+)$`)

func (r *pg18Report) addPostgresDiagnostics(output string) bool {
	const issueLimit = 100
	failed := false
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		match := pg18Diagnostic.FindStringSubmatch(line)
		if match == nil {
			if !strings.HasPrefix(line, "psql:") {
				continue
			}
			match = []string{line, "ERROR", line}
		}
		severity := "error"
		switch strings.ToUpper(match[1]) {
		case "WARNING":
			severity = "warning"
		case "ERROR", "FATAL", "PANIC":
			failed = true
		default:
			continue
		}
		if len(r.Issues) >= issueLimit {
			r.OmittedIssues++
			continue
		}
		message := match[2]
		if len(message) > 1000 {
			message = message[:1000] + "…"
		}
		r.Issues = append(r.Issues, pg18Issue{pg18IssueCategory(message), severity, message})
	}
	return failed
}

func pg18IssueCategory(message string) string {
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(msg, "extension"):
		return "extension"
	case strings.Contains(msg, "role "), strings.Contains(msg, "permission denied"), strings.Contains(msg, "must be owner"):
		return "role"
	case strings.Contains(msg, "unrecognized configuration parameter"), strings.Contains(msg, "not supported"),
		strings.Contains(msg, "syntax error"),
		(strings.HasPrefix(msg, "function ") || strings.HasPrefix(msg, "column ")) && strings.Contains(msg, "does not exist"):
		return "feature"
	default:
		return "restore"
	}
}
