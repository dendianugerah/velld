package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
)

var walName = regexp.MustCompile(`^[0-9A-F]{24}$`)
var backupStartWAL = regexp.MustCompile(`(?m)^START WAL LOCATION: .* \(file ([0-9A-F]{24})\)$`)

const postgresRecoverySettings = `SELECT json_object_agg(name, setting)
	FROM pg_catalog.pg_settings WHERE name IN ('max_connections', 'max_worker_processes',
		'max_wal_senders', 'max_prepared_transactions', 'max_locks_per_transaction')`

func postgresRecoveryArchive(conn *connection.StoredConnection) (string, error) {
	root := os.Getenv("VELLD_PITR_WAL_DIR")
	if root == "" {
		return "", recoveryInputError("Set VELLD_PITR_WAL_DIR to the PostgreSQL WAL archive mounted on the API")
	}
	dir := filepath.Join(root, conn.ID)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return "", recoveryInputError("Create the PostgreSQL WAL archive directory for this connection and configure archive_command")
	}
	return dir, nil
}

func (s *BackupService) capturePostgresBaseline(ctx context.Context, conn *connection.StoredConnection, dir string, base *recoveryBaseline) error {
	if _, err := postgresRecoveryArchive(conn); err != nil {
		return err
	}
	source, closeSource, err := s.recoverySource(conn)
	if err != nil {
		return err
	}
	defer closeSource()
	sslMode := "disable"
	if source.SSL {
		sslMode = "require"
	}
	dsn := url.URL{Scheme: "postgres", Host: net.JoinHostPort(source.Host, strconv.Itoa(source.Port)),
		Path: "/" + source.DatabaseName, User: url.User(source.Username), RawQuery: "sslmode=" + sslMode}
	dbURL := dsn
	dbURL.User = url.UserPassword(source.Username, source.Password)
	db, err := sql.Open("postgres", dbURL.String())
	if err != nil {
		return err
	}
	defer db.Close()
	var archiveMode string
	var unsupported bool
	err = db.QueryRowContext(ctx, `SELECT current_setting('server_version_num'), current_setting('archive_mode'),
		(SELECT rolname FROM pg_catalog.pg_roles WHERE oid = 10),
		pg_catalog.pg_is_in_recovery() OR EXISTS (SELECT 1 FROM pg_catalog.pg_tablespace WHERE oid NOT IN (1663, 1664))
	`).Scan(&base.Version, &archiveMode, &base.Bootstrap, &unsupported)
	if err != nil {
		return err
	}
	if unsupported || archiveMode == "off" {
		return recoveryInputError("Use a PostgreSQL primary with WAL archiving enabled and no custom tablespaces")
	}
	version, err := strconv.Atoi(base.Version)
	if err != nil || version < 130000 {
		return recoveryInputError("Point-in-time recovery requires PostgreSQL 13 or later")
	}
	base.Version = strconv.Itoa(version / 10000)
	var settings []byte
	if err := db.QueryRowContext(ctx, postgresRecoverySettings).Scan(&settings); err != nil {
		return err
	}
	if err := json.Unmarshal(settings, &base.Settings); err != nil {
		return err
	}
	tool, err := exec.LookPath("pg_basebackup")
	if err != nil {
		return err
	}
	data := filepath.Join(dir, "data")
	cmd := exec.CommandContext(ctx, tool, "--no-password", "--dbname="+dsn.String(),
		"--pgdata="+data, "--wal-method=stream", "--checkpoint=fast")
	cmd.Env = append(os.Environ(), "PGPASSWORD="+source.Password)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_basebackup: %w: %s", err, output)
	}
	label, err := os.ReadFile(filepath.Join(data, "backup_label"))
	if err != nil {
		return err
	}
	wal := backupStartWAL.FindSubmatch(label)
	if wal == nil {
		return fmt.Errorf("cannot determine the starting WAL segment")
	}
	base.Log = string(wal[1])
	return db.QueryRowContext(ctx, "SELECT pg_catalog.clock_timestamp()").Scan(&base.From)
}

func (s *BackupService) recoverPostgres(ctx context.Context, conn *connection.StoredConnection, target *restoreCheckTarget,
	dir string, base *recoveryBaseline, at time.Time, user, password string,
) (*throwawayContainer, error) {
	archive, err := postgresRecoveryArchive(conn)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(archive)
	if err != nil {
		return nil, err
	}
	if !walName.MatchString(base.Log) {
		return nil, fmt.Errorf("invalid starting WAL segment")
	}
	files := []recoveryFile{{name: "velld-archive/", mode: os.ModeDir | 0755}}
	for _, item := range entries {
		name := item.Name()
		if walName.MatchString(name) && name[:8] == base.Log[:8] && name >= base.Log {
			files = append(files, recoveryFile{name: "velld-archive/" + name, path: filepath.Join(archive, name)})
		}
	}

	config := fmt.Sprintf(`listen_addresses = '*'
port = 5432
unix_socket_directories = '/tmp'
hba_file = '/velld-pitr/pg_hba.conf'
ident_file = '/velld-pitr/pg_ident.conf'
archive_mode = off
restore_command = 'cp /velld-archive/%%f %%p'
recovery_target_timeline = 'current'
recovery_target_time = '%s'
recovery_target_action = 'promote'
`, at.UTC().Format("2006-01-02 15:04:05.000000-07:00"))
	for _, name := range []string{"max_connections", "max_worker_processes", "max_wal_senders", "max_prepared_transactions", "max_locks_per_transaction"} {
		value, err := strconv.Atoi(base.Settings[name])
		if err != nil {
			return nil, err
		}
		config += fmt.Sprintf("%s = %d\n", name, value)
	}
	if strings.ContainsAny(base.Database, "\r\n") {
		return nil, recoveryInputError("Database names with line breaks are not supported for PostgreSQL recovery")
	}
	database := "\"" + strings.ReplaceAll(base.Database, "\"", "\"\"") + "\""
	hba := fmt.Sprintf("local all all trust\nhost %s %s all scram-sha-256\nhost all all all reject\n", database, user)
	files = append(files,
		recoveryFile{name: "velld-pitr/postgresql.conf", content: config},
		recoveryFile{name: "velld-pitr/pg_hba.conf", content: hba},
		recoveryFile{name: "velld-pitr/pg_ident.conf"},
		recoveryFile{name: "velld-pitr/data/postgresql.auto.conf"},
		recoveryFile{name: "velld-pitr/data/recovery.signal"},
		recoveryFile{name: "velld-pitr/ready"},
	)
	docker := s.restoreChecks.docker
	spec := map[string]any{
		"Entrypoint": []string{"sh", "-ec"},
		"Cmd": []string{`while [ ! -f /velld-pitr/ready ]; do sleep 1; done
exec docker-entrypoint.sh postgres -c config_file=/velld-pitr/postgresql.conf`},
	}
	container, err := docker.startConfigured(ctx, target.image, target.port, []string{"PGDATA=/velld-pitr/data"}, true, spec)
	if err != nil {
		return nil, err
	}
	ready := false
	defer func() {
		if !ready {
			docker.remove(container.id)
		}
	}()
	err = docker.uploadRecovery(ctx, container.id, filepath.Join(dir, "data"), files)
	if err != nil {
		return nil, err
	}
	query := []string{"psql", "-XAtq", "-v", "ON_ERROR_STOP=1", "--host=/tmp", "--username=" + base.Bootstrap, "--dbname=" + base.Database}
	for {
		result, err := docker.recoveryExec(ctx, container.id, append(query, "--command=SELECT pg_is_in_recovery()"))
		if err == nil && strings.TrimSpace(result) == "f" {
			break
		}
		if !docker.recoveryRunning(ctx, container.id) {
			return nil, recoveryInputError("PostgreSQL could not reach that timestamp; make sure the WAL archive is complete and the recovery image is compatible")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	grant := fmt.Sprintf(`CREATE ROLE %s LOGIN PASSWORD '%s';
DO $$ DECLARE ns text; BEGIN
FOR ns IN SELECT nspname FROM pg_catalog.pg_namespace WHERE nspname !~ '^pg_' AND nspname <> 'information_schema' LOOP
 EXECUTE pg_catalog.format('GRANT USAGE ON SCHEMA %%I TO %s', ns);
 EXECUTE pg_catalog.format('GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %%I TO %s', ns);
 EXECUTE pg_catalog.format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %%I TO %s', ns);
END LOOP;
END $$;`, user, password, user, user, user)
	if _, err := docker.recoveryExec(ctx, container.id, append(query, "--command="+grant)); err != nil {
		return nil, err
	}
	target.conn.Host, target.conn.Port, target.conn.Username, target.conn.Password = container.host, container.port, user, password
	if err := waitForRestoreTarget(ctx, &target.conn); err != nil {
		return nil, err
	}
	ready = true
	return container, nil
}
