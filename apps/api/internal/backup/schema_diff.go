package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var errSchemaDiffInput = errors.New("invalid schema comparison")

type schemaDiffRequest struct {
	TargetBackupID string `json:"target_backup_id"`
	TargetLive     bool   `json:"target_live"`
}

func (h *BackupHandler) SchemaDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req schemaDiffRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil ||
		(req.TargetBackupID != "") == req.TargetLive {
		response.SendError(w, http.StatusBadRequest, "Choose either a target backup or the live database")
		return
	}
	migration, err := h.backupService.compareSchemas(r.Context(), userID, mux.Vars(r)["id"], req)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		response.SendError(w, http.StatusNotFound, "Backup not found")
	case errors.Is(err, errSchemaDiffInput):
		response.SendError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		response.SendError(w, http.StatusInternalServerError, err.Error())
	default:
		response.SendSuccess(w, "Schema compared", map[string]string{"migration_sql": migration})
	}
}

func (s *BackupService) schemaDiffBackup(id string, userID uuid.UUID) (*Backup, *connection.StoredConnection, error) {
	b, err := s.backupRepo.GetBackup(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, sql.ErrNoRows
		}
		return nil, nil, errors.New("Could not look up the backup")
	}
	conn, err := s.connStorage.GetConnection(b.ConnectionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conn.UserID != userID) {
		return nil, nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, nil, errors.New("Could not look up the backup connection")
	}
	if b.Status != "completed" {
		return nil, nil, fmt.Errorf("only completed backups can be compared: %w", errSchemaDiffInput)
	}
	if conn.Type != "postgresql" && conn.Type != "mysql" {
		return nil, nil, fmt.Errorf("schema diff is supported only for PostgreSQL and MySQL: %w", errSchemaDiffInput)
	}
	return b, conn, nil
}

func (s *BackupService) compareSchemas(ctx context.Context, userID uuid.UUID, id string, req schemaDiffRequest) (string, error) {
	source, conn, err := s.schemaDiffBackup(id, userID)
	if err != nil {
		return "", err
	}
	var target *Backup
	var targetConn *connection.StoredConnection
	if !req.TargetLive {
		target, targetConn, err = s.schemaDiffBackup(req.TargetBackupID, userID)
		if err != nil {
			return "", err
		}
		if targetConn.Type != conn.Type {
			return "", fmt.Errorf("both backups must use the same database engine: %w", errSchemaDiffInput)
		}
		if source.ID == target.ID {
			return "", nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, restoreCheckTimeout)
	defer cancel()
	select {
	case s.restoreChecks.slot <- struct{}{}:
		defer func() { <-s.restoreChecks.slot }()
	case <-ctx.Done():
		return "", errors.New("Timed out waiting to start the schema comparison")
	}
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err = s.restoreChecks.docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		return "", errors.New("Docker is not reachable from the API; mount /var/run/docker.sock to compare schemas")
	}

	before, err := s.readBackupSchema(ctx, source, conn)
	if err != nil {
		return "", err
	}
	var after databaseSchema
	if req.TargetLive {
		live := *conn
		live.DatabaseName = backupDatabaseName(source, conn)
		tunnel, host, port, err := s.setupSSHTunnelIfNeeded(&live)
		if err != nil {
			return "", errors.New("Could not open the live database SSH tunnel")
		}
		if tunnel != nil {
			defer tunnel.Stop()
		}
		live.Host, live.Port = host, port
		after, err = readDatabaseSchema(ctx, &live)
		if err != nil {
			return "", fmt.Errorf("Could not inspect the live database schema: %w", schemaDiffInspectionError(err))
		}
	} else {
		after, err = s.readBackupSchema(ctx, target, targetConn)
		if err != nil {
			return "", err
		}
	}
	return generateMigration(conn.Type == "postgresql", before, after)
}

func (s *BackupService) readBackupSchema(ctx context.Context, b *Backup, conn *connection.StoredConnection) (databaseSchema, error) {
	file, temporary, err := s.ensureBackupFileAvailable(b, conn.UserID)
	if err != nil {
		return databaseSchema{}, errors.New("The backup file is unavailable")
	}
	if temporary {
		defer os.Remove(file)
	}
	dump, err := inspectDump(conn.Type, file)
	if err != nil {
		return databaseSchema{}, errors.New("The backup file is unreadable")
	}
	target := restoreCheckTargetFor(conn.Type, dump.serverVersion, backupDatabaseName(b, conn))
	docker := s.restoreChecks.docker
	container, err := docker.start(ctx, target.image, target.port, target.env, false)
	if err != nil {
		return databaseSchema{}, errors.New("Could not start the schema comparison container")
	}
	defer docker.remove(container.id)
	target.conn.Host, target.conn.Port = container.host, container.port
	if _, err := s.restoreIntoTarget(ctx, &target, file, dump); err != nil {
		return databaseSchema{}, errors.New("Could not restore the backup for schema comparison")
	}
	schema, err := readDatabaseSchema(ctx, &target.conn)
	if err != nil {
		return databaseSchema{}, fmt.Errorf("Could not inspect the restored schema: %w", schemaDiffInspectionError(err))
	}
	return schema, nil
}

func schemaDiffInspectionError(err error) error {
	if errors.Is(err, errSchemaDiffInput) {
		return err
	}
	return errors.New("verify database availability and metadata access")
}
