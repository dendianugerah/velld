package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	backupBranchLabel = "velld.backup-branch"
	backupBranchTTL   = 24 * time.Hour
)

var (
	errBranchUnavailable = errors.New("only completed backups can be branched")
	errBranchUnsupported = errors.New("branches are supported for PostgreSQL, MySQL, MariaDB, and MongoDB")
)

type backupBranch struct {
	ConnectionString string    `json:"connection_string"`
	ExpiresAt        time.Time `json:"expires_at"`
}

func (h *BackupHandler) CreateBackupBranch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	branch, err := h.backupService.createBackupBranch(r.Context(), mux.Vars(r)["id"], userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		response.SendError(w, http.StatusNotFound, "Backup not found")
	case errors.Is(err, errBranchUnavailable), errors.Is(err, errBranchUnsupported):
		response.SendError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		response.SendError(w, http.StatusInternalServerError, err.Error())
	default:
		response.SendSuccess(w, "Backup branch created", branch)
	}
}

func (s *BackupService) createBackupBranch(ctx context.Context, id string, userID uuid.UUID) (*backupBranch, error) {
	backup, err := s.backupRepo.GetBackup(id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, errors.New("Could not look up the backup")
	}
	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conn.UserID != userID) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, errors.New("Could not look up the backup connection")
	}
	if backup.Status != "completed" {
		return nil, errBranchUnavailable
	}
	if _, ok := restoreTools[conn.Type]; !ok {
		return nil, errBranchUnsupported
	}

	ctx, cancel := context.WithTimeout(ctx, restoreCheckTimeout)
	defer cancel()
	select {
	case s.restoreChecks.slot <- struct{}{}:
		defer func() { <-s.restoreChecks.slot }()
	case <-ctx.Done():
		return nil, errors.New("Timed out waiting to start the backup branch")
	}
	docker := s.restoreChecks.docker
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err = docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		return nil, errors.New("Docker is not reachable from the API; mount /var/run/docker.sock to create backup branches")
	}
	file, temporary, err := s.ensureBackupFileAvailable(backup, userID)
	if err != nil {
		return nil, errors.New("The backup file is unavailable")
	}
	if temporary {
		defer os.Remove(file)
	}
	dump, err := inspectDump(conn.Type, file)
	if err != nil {
		return nil, errors.New("The backup file is unreadable")
	}

	target := restoreCheckTargetFor(conn.Type, dump.serverVersion, backupDatabaseName(backup, conn))
	if conn.Type == "mongodb" {
		target.conn.Username, target.conn.Password, target.conn.MongoAuthSource = "velld", randomPassword(), "admin"
		target.env = []string{"MONGO_INITDB_ROOT_USERNAME=" + target.conn.Username, "MONGO_INITDB_ROOT_PASSWORD=" + target.conn.Password}
	}
	expires := time.Now().UTC().Truncate(time.Second).Add(backupBranchTTL)
	container, err := docker.start(ctx, target.image, target.port, target.env, true)
	if err != nil {
		return nil, errors.New("Could not start the backup branch container")
	}
	ready := false
	defer func() {
		if !ready {
			docker.remove(container.id)
		}
	}()
	target.conn.Host, target.conn.Port = container.host, container.port
	if _, err := s.restoreIntoTarget(ctx, &target, file, dump); err != nil {
		return nil, errors.New("Could not restore the backup into the branch")
	}

	user, password := "velld_"+randomPassword()[:16], randomPassword()
	if err := maskBranch(ctx, &target.conn, user, password); err != nil {
		return nil, errors.New("Could not safely mask the backup branch; no connection was issued")
	}
	ready = true
	return &backupBranch{ConnectionString: branchConnectionString(&target.conn, container.publishedPort, user, password), ExpiresAt: expires}, nil
}

func branchConnectionString(conn *connection.StoredConnection, port int, user, password string) string {
	uri := url.URL{
		Scheme:  conn.Type,
		User:    url.UserPassword(user, password),
		Host:    net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		Path:    "/" + conn.DatabaseName,
		RawPath: "/" + url.PathEscape(conn.DatabaseName),
	}
	switch conn.Type {
	case "postgresql":
		uri.RawQuery = "sslmode=disable"
	case "mariadb":
		uri.Scheme = "mysql"
	}
	return uri.String()
}

func (s *BackupService) CleanExpiredBackupBranches() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), restoreCheckCleanupTimeout)
		s.restoreChecks.docker.removeContainersBefore(ctx, backupBranchLabel, time.Now().Add(-backupBranchTTL))
		cancel()
		<-ticker.C
	}
}

func maskBranch(ctx context.Context, conn *connection.StoredConnection, user, password string) error {
	mask := newBranchMasker()
	switch conn.Type {
	case "mongodb":
		return mask.mongo(ctx, conn, user, password)
	case "postgresql", "mysql", "mariadb":
		return mask.sql(ctx, conn, user, password)
	default:
		return fmt.Errorf("unsupported database type %s", conn.Type)
	}
}
