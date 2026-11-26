package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type recoveryBaseline struct {
	From      time.Time         `json:"from"`
	Through   time.Time         `json:"through,omitzero"`
	Engine    string            `json:"engine"`
	Version   string            `json:"version"`
	Database  string            `json:"database"`
	SystemID  string            `json:"system_id"`
	Log       string            `json:"log"`
	LastLog   string            `json:"last_log,omitempty"`
	Position  string            `json:"position,omitempty"`
	Bootstrap string            `json:"bootstrap,omitempty"`
	Settings  map[string]string `json:"settings,omitempty"`
}

var recoveryBusy sync.Map

type recoveryInputError string

func (e recoveryInputError) Error() string { return string(e) }

func (h *BackupHandler) PointInTimeRecovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	owner, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	s := h.backupService
	id := mux.Vars(r)["id"]
	if _, err := uuid.Parse(id); err != nil {
		response.SendError(w, http.StatusNotFound, "Connection not found")
		return
	}
	conn, err := s.connStorage.GetConnection(id)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conn.UserID != owner) {
		response.SendError(w, http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Could not look up the connection")
		return
	}
	if conn.Type != "postgresql" && conn.Type != "mysql" {
		response.SendError(w, http.StatusBadRequest, "Point-in-time recovery supports PostgreSQL and MySQL")
		return
	}
	if _, busy := recoveryBusy.LoadOrStore(conn.ID, true); busy {
		response.SendError(w, http.StatusConflict, "A point-in-time recovery operation is already running for this connection")
		return
	}
	defer recoveryBusy.Delete(conn.ID)

	root := filepath.Join(s.backupDir, ".pitr")
	dir := filepath.Join(root, conn.ID)
	base, err := readRecoveryBaseline(dir)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Could not read the recovery baseline")
		return
	}
	if r.Method == http.MethodGet {
		response.SendSuccess(w, "Point-in-time recovery status", recoveryStatus(base))
		return
	}
	if r.Method == http.MethodDelete {
		if err := os.RemoveAll(dir); err != nil {
			response.SendError(w, http.StatusInternalServerError, "Could not remove the recovery baseline")
			return
		}
		response.SendSuccess(w, "Recovery baseline removed", nil)
		return
	}
	var request struct {
		Target *time.Time `json:"target"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&request); err != nil {
		response.SendError(w, http.StatusBadRequest, "Send a JSON object with an optional RFC3339 target timestamp")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), restoreCheckTimeout)
	defer cancel()
	var data any
	if request.Target == nil {
		if base == nil {
			base, err = s.captureRecoveryBaseline(ctx, root, dir, conn)
		} else if conn.Type == "mysql" {
			err = s.archiveMySQL(ctx, conn, dir, base)
		}
		data = recoveryStatus(base)
	} else {
		data, err = s.recoverAt(ctx, conn, dir, base, request.Target.UTC())
	}
	var inputErr recoveryInputError
	switch {
	case errors.As(err, &inputErr):
		response.SendError(w, http.StatusBadRequest, inputErr.Error())
	case err != nil:
		response.SendError(w, http.StatusInternalServerError, "Point-in-time recovery failed; check the server tools, source configuration and Docker logs")
	default:
		response.SendSuccess(w, "Point-in-time recovery operation completed", data)
	}
}

func recoveryStatus(base *recoveryBaseline) any {
	if base == nil {
		return struct{}{}
	}
	return struct {
		From    time.Time `json:"from"`
		Through time.Time `json:"through,omitzero"`
	}{base.From, base.Through}
}

func readRecoveryBaseline(dir string) (*recoveryBaseline, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "baseline.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var base recoveryBaseline
	return &base, json.Unmarshal(raw, &base)
}

func writeRecoveryBaseline(dir string, base *recoveryBaseline) error {
	raw, err := json.Marshal(base)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, ".baseline.json")
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "baseline.json"))
}

func (s *BackupService) captureRecoveryBaseline(ctx context.Context, root, dir string, conn *connection.StoredConnection) (*recoveryBaseline, error) {
	if conn.DatabaseName == "" {
		return nil, recoveryInputError("Set a database name on the connection first")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(root, ".baseline-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	base := &recoveryBaseline{Engine: conn.Type, Database: conn.DatabaseName}
	switch conn.Type {
	case "postgresql":
		err = s.capturePostgresBaseline(ctx, conn, tmp, base)
	case "mysql":
		err = s.captureMySQLBaseline(ctx, conn, tmp, base)
	default:
		err = recoveryInputError("Unsupported database type")
	}
	if err != nil {
		return nil, err
	}
	if err := writeRecoveryBaseline(tmp, base); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, dir); err != nil {
		return nil, err
	}
	return base, nil
}

func (s *BackupService) recoverAt(ctx context.Context, conn *connection.StoredConnection, dir string, base *recoveryBaseline, target time.Time) (*backupBranch, error) {
	if base == nil || base.Engine != conn.Type || base.Database != conn.DatabaseName {
		return nil, recoveryInputError("Create a recovery baseline for this database first")
	}
	if target.Before(base.From) || !target.Before(time.Now()) {
		return nil, recoveryInputError("Choose a past timestamp at or after the baseline")
	}
	if base.Engine == "mysql" && target.Nanosecond() != 0 {
		return nil, recoveryInputError("Choose a whole-second timestamp for MySQL recovery")
	}
	if base.Engine == "mysql" && target.After(base.Through) {
		if err := s.archiveMySQL(ctx, conn, dir, base); err != nil {
			return nil, err
		}
		if target.After(base.Through) {
			return nil, recoveryInputError("The MySQL archive has not reached the requested timestamp")
		}
	}
	select {
	case s.restoreChecks.slot <- struct{}{}:
		defer func() { <-s.restoreChecks.slot }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	docker := s.restoreChecks.docker
	if err := docker.ping(ctx); err != nil {
		return nil, recoveryInputError("Mount /var/run/docker.sock on the API to recover into a temporary database")
	}
	db := restoreCheckTargetFor(base.Engine, base.Version, base.Database)
	user, password := "velld_"+randomPassword()[:16], randomPassword()
	var container *throwawayContainer
	var err error
	switch base.Engine {
	case "postgresql":
		container, err = s.recoverPostgres(ctx, conn, &db, dir, base, target, user, password)
	case "mysql":
		container, err = s.recoverMySQL(ctx, &db, dir, base, target, user, password)
	default:
		return nil, fmt.Errorf("invalid recovery engine")
	}
	if err != nil {
		return nil, err
	}
	return &backupBranch{
		ConnectionString: branchConnectionString(&db.conn, container.publishedPort, user, password),
		ExpiresAt:        time.Now().UTC().Truncate(time.Second).Add(backupBranchTTL),
	}, nil
}

func (s *BackupService) recoverySource(conn *connection.StoredConnection) (connection.StoredConnection, func(), error) {
	effective := *conn
	tunnel, host, port, err := s.setupSSHTunnelIfNeeded(conn)
	if err != nil {
		return effective, nil, err
	}
	effective.Host, effective.Port = host, port
	return effective, func() {
		if tunnel != nil {
			tunnel.Stop()
		}
	}, nil
}
