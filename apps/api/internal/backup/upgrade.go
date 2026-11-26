package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/common/response"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	upgradeTimeout          = 25 * time.Minute
	upgradeReadyTimeout     = 3 * time.Minute
	upgradeRollbackTimeout  = 3 * time.Minute
	upgradeReadyPollInterval = 500 * time.Millisecond
)

var (
	errContainerNotFound  = errors.New("no running Docker container matched this connection")
	errUpgradeUnsupported = errors.New("safe upgrades are supported for PostgreSQL, MySQL, MariaDB, and MongoDB")
	errInvalidUpgradeReq  = errors.New("invalid upgrade request")
)

type UpgradeInfo struct {
	Supported     bool   `json:"supported"`
	ContainerID   string `json:"container_id,omitempty"`
	ContainerName string `json:"container_name,omitempty"`
	CurrentImage  string `json:"current_image,omitempty"`
	DatabaseType  string `json:"database_type"`
	Hint          string `json:"hint,omitempty"`
}

type UpgradeRequest struct {
	TargetImage string `json:"target_image"`
}

type UpgradeResult struct {
	Status           string `json:"status"`
	ContainerID      string `json:"container_id"`
	ContainerName    string `json:"container_name"`
	PreviousImage    string `json:"previous_image"`
	TargetImage      string `json:"target_image"`
	PreviousVersion  string `json:"previous_version,omitempty"`
	CurrentVersion   string `json:"current_version,omitempty"`
	SnapshotBackupID string `json:"snapshot_backup_id"`
	HealthCheck      string `json:"health_check,omitempty"`
	RolledBack       bool   `json:"rolled_back"`
	RollbackVerified bool   `json:"rollback_verified,omitempty"`
	Error            string `json:"error,omitempty"`
}

type databaseBaseline struct {
	version string
	objects int
	records int64
}

func (h *BackupHandler) GetUpgradeInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	info, err := h.backupService.GetUpgradeInfo(r.Context(), mux.Vars(r)["id"], userID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		response.SendError(w, http.StatusNotFound, "Connection not found")
	case err != nil:
		response.SendError(w, http.StatusInternalServerError, err.Error())
	default:
		response.SendSuccess(w, "Upgrade info retrieved", info)
	}
}

func (h *BackupHandler) UpgradeConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID, err := common.GetUserIDFromContext(r.Context())
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req UpgradeRequest
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid upgrade request body")
		return
	}
	result, err := h.backupService.UpgradeConnection(r.Context(), mux.Vars(r)["id"], userID, req.TargetImage)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		response.SendError(w, http.StatusNotFound, "Connection not found")
	case errors.Is(err, errUpgradeUnsupported), errors.Is(err, errInvalidUpgradeReq), errors.Is(err, errContainerNotFound):
		response.SendError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		response.SendError(w, http.StatusInternalServerError, err.Error())
	case result.RolledBack:
		response.SendSuccess(w, "Upgrade failed and was rolled back", result)
	default:
		response.SendSuccess(w, "Database upgraded successfully", result)
	}
}

func (s *BackupService) GetUpgradeInfo(ctx context.Context, connectionID string, userID uuid.UUID) (*UpgradeInfo, error) {
	conn, err := s.connStorage.GetConnection(connectionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conn.UserID != userID) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("failed to look up connection: %w", err)
	}

	info := &UpgradeInfo{DatabaseType: conn.Type}
	if _, ok := restoreTools[conn.Type]; !ok {
		info.Hint = errUpgradeUnsupported.Error()
		return info, nil
	}

	docker := s.restoreChecks.docker
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err = docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		info.Hint = "Docker is not reachable from the API; mount /var/run/docker.sock to enable safe upgrades."
		return info, nil
	}

	findCtx, findCancel := context.WithTimeout(ctx, 10*time.Second)
	defer findCancel()
	container, _, err := docker.findContainerForConnection(findCtx, conn)
	if err != nil {
		info.Hint = "No running Docker container matched this connection."
		return info, nil
	}

	info.Supported = true
	info.ContainerID = container.ID
	info.ContainerName = strings.TrimPrefix(container.Name, "/")
	info.CurrentImage = container.Config.Image
	return info, nil
}

func (s *BackupService) UpgradeConnection(ctx context.Context, connectionID string, userID uuid.UUID, targetImage string) (*UpgradeResult, error) {
	return s.upgradeConnectionWithTimeout(ctx, connectionID, userID, targetImage, upgradeReadyTimeout)
}

func (s *BackupService) upgradeConnectionWithTimeout(
	ctx context.Context,
	connectionID string,
	userID uuid.UUID,
	targetImage string,
	readyTimeout time.Duration,
) (*UpgradeResult, error) {
	conn, err := s.connStorage.GetConnection(connectionID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && conn.UserID != userID) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("failed to look up connection: %w", err)
	}
	if _, ok := restoreTools[conn.Type]; !ok {
		return nil, errUpgradeUnsupported
	}
	targetImage = strings.TrimSpace(targetImage)
	if targetImage == "" {
		return nil, fmt.Errorf("%w: target_image is required", errInvalidUpgradeReq)
	}

	ctx, cancel := context.WithTimeout(ctx, upgradeTimeout)
	defer cancel()

	select {
	case s.restoreChecks.slot <- struct{}{}:
		defer func() { <-s.restoreChecks.slot }()
	case <-ctx.Done():
		return nil, errors.New("timed out waiting for another Docker operation to finish")
	}

	docker := s.restoreChecks.docker
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err = docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		return nil, errors.New("Docker is not reachable from the API; mount /var/run/docker.sock to run safe upgrades")
	}

	orig, containerPort, err := docker.findContainerForConnection(ctx, conn)
	if err != nil {
		return nil, err
	}
	if err := validateTargetUpgradeImage(conn.Type, orig.Config.Image, targetImage); err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalidUpgradeReq, err)
	}

	// Step 1: Capture pre-upgrade health baseline and take a durable Velld backup snapshot.
	baseline, err := inspectDatabaseHealth(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("pre-upgrade health check failed: %w", err)
	}

	snapshot, err := s.createSingleDatabaseBackup(conn, conn.DatabaseName)
	if err != nil || snapshot == nil || snapshot.Status != "completed" {
		return nil, fmt.Errorf("pre-upgrade snapshot failed: %w", err)
	}

	snapshotFile, isTemp, err := s.ensureBackupFileAvailable(snapshot, userID)
	if err != nil {
		return nil, fmt.Errorf("snapshot file is unavailable: %w", err)
	}
	if isTemp {
		defer os.Remove(snapshotFile)
	}

	dump, err := inspectDump(conn.Type, snapshotFile)
	if err != nil {
		return nil, fmt.Errorf("snapshot file is unreadable: %w", err)
	}

	// Step 2: Pull target image before stopping the live container so pull failures cause zero downtime.
	origImage, _ := docker.inspectImage(ctx, orig.Image)
	if err := docker.ensureImage(ctx, targetImage); err != nil {
		return nil, fmt.Errorf("could not pull target image %s: %w", targetImage, err)
	}

	// Step 3: Stop and park the original container under a rollback name without modifying its data.
	origName := strings.TrimPrefix(orig.Name, "/")
	shortSuffix := randomPassword()[:8]
	rollbackName := origName + "-velld-rollback-" + shortSuffix

	if err := docker.stopContainer(ctx, orig.ID, upgradeStopTimeout); err != nil {
		return nil, fmt.Errorf("failed to stop original container: %w", err)
	}
	if err := docker.renameContainer(ctx, orig.ID, rollbackName); err != nil {
		_ = docker.startContainer(context.Background(), orig.ID)
		return nil, fmt.Errorf("failed to stage original container for rollback: %w", err)
	}

	plan := buildUpgradedContainerPlan(orig, origImage, targetImage, conn.Type, shortSuffix)

	rollback := func(failureReason error) (*UpgradeResult, error) {
		rbCtx, rbCancel := context.WithTimeout(context.Background(), upgradeRollbackTimeout)
		defer rbCancel()

		for _, vol := range plan.createdVolumes {
			docker.removeVolume(vol)
		}
		if renameErr := docker.renameContainer(rbCtx, orig.ID, origName); renameErr != nil {
			return nil, fmt.Errorf("upgrade failed (%v) and rollback rename failed: %w", failureReason, renameErr)
		}
		if startErr := docker.startContainer(rbCtx, orig.ID); startErr != nil {
			return nil, fmt.Errorf("upgrade failed (%v) and restarting original container failed: %w", failureReason, startErr)
		}

		// If the original container used an ephemeral host port, detect its new bound port after restart.
		verifyConn := *conn
		if restarted, inspErr := docker.inspectContainer(rbCtx, orig.ID); inspErr == nil {
			if newPort := resolvePublishedPort(restarted, containerPort); newPort > 0 && newPort != conn.Port {
				verifyConn.Port = newPort
				conn.Port = newPort
				_ = s.connStorage.Update(*conn)
			}
		}

		rbReadyCtx, rbReadyCancel := context.WithTimeout(rbCtx, upgradeReadyTimeout)
		rbErr := waitForUpgradedDatabase(rbReadyCtx, docker, orig.ID, &verifyConn, orig.Config.Env)
		rbReadyCancel()

		verified := false
		var rbHealth string
		if rbErr == nil {
			if current, hErr := verifyUpgradedHealth(rbCtx, docker, orig.ID, &verifyConn, baseline); hErr == nil {
				verified = true
				rbHealth = formatHealthSummary(verifyConn.Type, current)
			}
		}

		return &UpgradeResult{
			Status:           "rolled_back",
			ContainerID:      orig.ID,
			ContainerName:    origName,
			PreviousImage:    orig.Config.Image,
			TargetImage:      targetImage,
			PreviousVersion:  baseline.version,
			CurrentVersion:   baseline.version,
			SnapshotBackupID: snapshot.ID.String(),
			HealthCheck:      rbHealth,
			RolledBack:       true,
			RollbackVerified: verified,
			Error:            failureReason.Error(),
		}, nil
	}

	// Step 4: Create and start the upgraded container, restore the snapshot, and run health checks.
	newID, err := docker.createUpgradedContainer(ctx, plan)
	if err != nil {
		return rollback(fmt.Errorf("create upgraded container: %w", err))
	}

	failAndRollback := func(cause error) (*UpgradeResult, error) {
		docker.remove(newID)
		return rollback(cause)
	}

	if err := docker.startContainer(ctx, newID); err != nil {
		return failAndRollback(fmt.Errorf("start upgraded container: %w", err))
	}

	readyCtx, readyCancel := context.WithTimeout(ctx, readyTimeout)
	err = waitForUpgradedDatabase(readyCtx, docker, newID, conn, orig.Config.Env)
	readyCancel()
	if err != nil {
		return failAndRollback(err)
	}

	if err := prepareUpgradedDatabase(ctx, conn, orig.Config.Env, dump.roles); err != nil {
		return failAndRollback(fmt.Errorf("prepare upgraded database: %w", err))
	}

	if _, err := s.restoreFile(ctx, conn, snapshotFile); err != nil {
		return failAndRollback(fmt.Errorf("restore snapshot into %s: %w", targetImage, err))
	}

	upgradedHealth, err := verifyUpgradedHealth(ctx, docker, newID, conn, baseline)
	if err != nil {
		return failAndRollback(fmt.Errorf("post-upgrade health check failed: %w", err))
	}

	// Step 5: Upgrade verified healthy — remove the parked old container.
	docker.remove(orig.ID)

	return &UpgradeResult{
		Status:           "upgraded",
		ContainerID:      newID,
		ContainerName:    origName,
		PreviousImage:    orig.Config.Image,
		TargetImage:      targetImage,
		PreviousVersion:  baseline.version,
		CurrentVersion:   upgradedHealth.version,
		SnapshotBackupID: snapshot.ID.String(),
		HealthCheck:      formatHealthSummary(conn.Type, upgradedHealth),
		RolledBack:       false,
	}, nil
}

func envMapFromSlice(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, entry := range env {
		if k, v, ok := strings.Cut(entry, "="); ok {
			m[k] = v
		}
	}
	return m
}

func serverPingConnection(conn *connection.StoredConnection, origEnv []string) connection.StoredConnection {
	pingConn := *conn
	env := envMapFromSlice(origEnv)
	switch conn.Type {
	case "postgresql":
		pingConn.DatabaseName = "postgres"
	case "mysql", "mariadb":
		pingConn.DatabaseName = ""
		if rootPass := env["MYSQL_ROOT_PASSWORD"]; rootPass != "" {
			pingConn.Username = "root"
			pingConn.Password = rootPass
		} else if rootPass := env["MARIADB_ROOT_PASSWORD"]; rootPass != "" {
			pingConn.Username = "root"
			pingConn.Password = rootPass
		}
	}
	return pingConn
}

func waitForUpgradedDatabase(
	ctx context.Context,
	docker *dockerEngine,
	containerID string,
	conn *connection.StoredConnection,
	origEnv []string,
) error {
	pingConn := serverPingConnection(conn, origEnv)
	pingOnce := func() error {
		attemptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if conn.Type == "mongodb" {
			client, err := openMongoTarget(attemptCtx, &pingConn)
			if err != nil {
				return err
			}
			defer client.Disconnect(context.Background())
			return client.Ping(attemptCtx, nil)
		}
		db, err := openSQLTarget(&pingConn)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.PingContext(attemptCtx)
	}

	var lastErr error
	for {
		details, inspErr := docker.inspectContainer(ctx, containerID)
		if inspErr == nil && !details.State.Running {
			logs := docker.containerLogsTail(context.Background(), containerID, 15)
			if logs != "" {
				return fmt.Errorf("container exited unexpectedly (status=%s, exit_code=%d): %s",
					details.State.Status, details.State.ExitCode, logs)
			}
			return fmt.Errorf("container exited unexpectedly (status=%s, exit_code=%d)",
				details.State.Status, details.State.ExitCode)
		}

		if err := pingOnce(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("container did not become ready before timeout: %w", lastErr)
			}
			return fmt.Errorf("container did not become ready before timeout: %w", ctx.Err())
		case <-time.After(upgradeReadyPollInterval):
		}
	}
}

func prepareUpgradedDatabase(ctx context.Context, conn *connection.StoredConnection, origEnv []string, pgRoles []string) error {
	switch conn.Type {
	case "postgresql":
		adminConn := serverPingConnection(conn, origEnv)
		db, err := openSQLTarget(&adminConn)
		if err != nil {
			return err
		}
		defer db.Close()
		var exists int
		err = db.QueryRowContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", conn.DatabaseName).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			if _, err := db.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(conn.DatabaseName)); err != nil {
				return fmt.Errorf("create database %s: %w", conn.DatabaseName, err)
			}
		} else if err != nil {
			return err
		}
		return createReferencedRoles(ctx, conn, pgRoles)

	case "mysql", "mariadb":
		adminConn := serverPingConnection(conn, origEnv)
		db, err := openSQLTarget(&adminConn)
		if err != nil {
			return err
		}
		defer db.Close()
		quotedDB := "`" + strings.ReplaceAll(conn.DatabaseName, "`", "``") + "`"
		if _, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+quotedDB); err != nil {
			return fmt.Errorf("create database %s: %w", conn.DatabaseName, err)
		}
		return nil

	default:
		return nil
	}
}

func inspectDatabaseHealth(ctx context.Context, conn *connection.StoredConnection) (databaseBaseline, error) {
	if conn.Type == "mongodb" {
		return inspectMongoHealth(ctx, conn)
	}
	return inspectSQLHealth(ctx, conn)
}

func inspectSQLHealth(ctx context.Context, conn *connection.StoredConnection) (databaseBaseline, error) {
	db, err := openSQLTarget(conn)
	if err != nil {
		return databaseBaseline{}, err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return databaseBaseline{}, err
	}

	var b databaseBaseline
	if conn.Type == "postgresql" {
		if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&b.version); err != nil {
			return databaseBaseline{}, fmt.Errorf("query PostgreSQL version: %w", err)
		}
		const userTables = `FROM information_schema.tables WHERE table_type = 'BASE TABLE'
			AND table_schema NOT IN ('pg_catalog', 'information_schema')`
		if err := db.QueryRowContext(ctx, "SELECT count(*) "+userTables).Scan(&b.objects); err != nil {
			return databaseBaseline{}, fmt.Errorf("count tables: %w", err)
		}
		if err := db.QueryRowContext(ctx, `SELECT coalesce(sum((xpath('/row/c/text()', query_to_xml(
			format('SELECT count(*) AS c FROM ONLY %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint), 0) `+
			userTables).Scan(&b.records); err != nil {
			return databaseBaseline{}, fmt.Errorf("count rows: %w", err)
		}
		return b, nil
	}

	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&b.version); err != nil {
		return databaseBaseline{}, fmt.Errorf("query MySQL version: %w", err)
	}
	rows, err := db.QueryContext(ctx, `SELECT table_name FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`, conn.DatabaseName)
	if err != nil {
		return databaseBaseline{}, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return databaseBaseline{}, err
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Err(); err != nil {
		return databaseBaseline{}, err
	}

	b.objects = len(tableNames)
	for _, name := range tableNames {
		quoted := "`" + strings.ReplaceAll(name, "`", "``") + "`"
		var count int64
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&count); err != nil {
			return databaseBaseline{}, fmt.Errorf("count rows in %s: %w", name, err)
		}
		b.records += count
	}
	return b, nil
}

func inspectMongoHealth(ctx context.Context, conn *connection.StoredConnection) (databaseBaseline, error) {
	client, err := openMongoTarget(ctx, conn)
	if err != nil {
		return databaseBaseline{}, err
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, nil); err != nil {
		return databaseBaseline{}, err
	}

	var b databaseBaseline
	var buildInfo struct {
		Version string `bson:"version"`
	}
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo); err == nil {
		b.version = buildInfo.Version
	} else if err := client.Database(conn.DatabaseName).RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo); err == nil {
		b.version = buildInfo.Version
	}

	db := client.Database(conn.DatabaseName)
	names, err := db.ListCollectionNames(ctx, bson.D{{Key: "type", Value: "collection"}})
	if err != nil {
		return databaseBaseline{}, fmt.Errorf("list collections: %w", err)
	}
	for _, name := range names {
		if strings.HasPrefix(name, "system.") {
			continue
		}
		count, err := db.Collection(name).CountDocuments(ctx, bson.D{})
		if err != nil {
			return databaseBaseline{}, fmt.Errorf("count documents in %s: %w", name, err)
		}
		b.objects++
		b.records += count
	}
	return b, nil
}

func verifyUpgradedHealth(
	ctx context.Context,
	docker *dockerEngine,
	containerID string,
	conn *connection.StoredConnection,
	baseline databaseBaseline,
) (databaseBaseline, error) {
	details, err := docker.inspectContainer(ctx, containerID)
	if err != nil {
		return databaseBaseline{}, fmt.Errorf("inspect container: %w", err)
	}
	if !details.State.Running || details.State.OOMKilled {
		return databaseBaseline{}, fmt.Errorf("container is not healthy (status=%s, oom=%v)", details.State.Status, details.State.OOMKilled)
	}

	current, err := inspectDatabaseHealth(ctx, conn)
	if err != nil {
		return databaseBaseline{}, err
	}
	if current.objects < baseline.objects {
		unit := "tables"
		if conn.Type == "mongodb" {
			unit = "collections"
		}
		return current, fmt.Errorf("expected at least %d %s after upgrade, found %d", baseline.objects, unit, current.objects)
	}
	if current.records < baseline.records {
		unit := "rows"
		if conn.Type == "mongodb" {
			unit = "documents"
		}
		return current, fmt.Errorf("expected at least %d %s after upgrade, found %d", baseline.records, unit, current.records)
	}
	return current, nil
}

func formatHealthSummary(dbType string, health databaseBaseline) string {
	if dbType == "mongodb" {
		return fmt.Sprintf("%d collections and %d documents verified", health.objects, health.records)
	}
	return fmt.Sprintf("%d tables and %d rows verified", health.objects, health.records)
}
