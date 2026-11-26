package backup

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/go-sql-driver/mysql"
)

var mysqlLogName = regexp.MustCompile(`^([A-Za-z0-9_][A-Za-z0-9_.-]*\.)([0-9]{6})$`)
var mysqlDumpPosition = regexp.MustCompile(`^-- CHANGE MASTER TO MASTER_LOG_FILE='([^']+)', MASTER_LOG_POS=([0-9]+);`)

func mysqlRecoveryDB(conn *connection.StoredConnection) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User, cfg.Passwd, cfg.DBName = conn.Username, conn.Password, conn.DatabaseName
	cfg.Net, cfg.Addr = "tcp", net.JoinHostPort(conn.Host, strconv.Itoa(conn.Port))
	cfg.ParseTime = true
	if conn.SSL {
		cfg.TLSConfig = "true"
	}
	return sql.Open("mysql", cfg.FormatDSN())
}

func mysqlRecoveryCommand(ctx context.Context, tool string, conn *connection.StoredConnection, args ...string) *exec.Cmd {
	options := []string{"--no-defaults", "--protocol=TCP", "--host=" + conn.Host,
		"--port=" + strconv.Itoa(conn.Port), "--user=" + conn.Username}
	switch {
	case tool == "mysqlbinlog" && conn.SSL:
		options = append(options, "--ssl-mode=VERIFY_IDENTITY")
	case tool == "mysqlbinlog":
		options = append(options, "--ssl-mode=DISABLED", "--get-server-public-key")
	case conn.SSL:
		options = append(options, "--ssl-verify-server-cert")
	default:
		options = append(options, "--skip-ssl")
	}
	cmd := exec.CommandContext(ctx, tool, append(options, args...)...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conn.Password, "TZ=UTC")
	return cmd
}

func mysqlRecoveryIdentity(ctx context.Context, db *sql.DB) (id, version string, now time.Time, err error) {
	var on, compressed bool
	var format string
	err = db.QueryRowContext(ctx, `SELECT @@global.server_uuid, VERSION(), UTC_TIMESTAMP(6),
		@@global.log_bin, @@global.binlog_format, @@global.binlog_transaction_compression`).Scan(&id, &version, &now, &on, &format, &compressed)
	if err == nil && (!on || format != "ROW" || compressed) {
		err = recoveryInputError("MySQL recovery needs uncompressed row-based binary logging enabled")
	}
	return
}

func (s *BackupService) captureMySQLBaseline(ctx context.Context, conn *connection.StoredConnection, dir string, base *recoveryBaseline) error {
	source, closeSource, err := s.recoverySource(conn)
	if err != nil {
		return err
	}
	defer closeSource()
	db, err := mysqlRecoveryDB(&source)
	if err != nil {
		return err
	}
	defer db.Close()
	base.SystemID, base.Version, _, err = mysqlRecoveryIdentity(ctx, db)
	if err != nil {
		return err
	}
	var unsupportedTables int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE' AND engine <> 'InnoDB'`, base.Database).Scan(&unsupportedTables); err != nil {
		return err
	}
	if unsupportedTables != 0 {
		return recoveryInputError("MySQL point-in-time recovery requires InnoDB tables")
	}
	dump := filepath.Join(dir, "baseline.sql")
	cmd := mysqlRecoveryCommand(ctx, "mariadb-dump", &source, "--single-transaction", "--master-data=2",
		"--routines", "--events", "--hex-blob", "--result-file="+dump, "--", base.Database)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mariadb-dump: %w: %s", err, output)
	}
	if err := readMySQLDumpPosition(dump, base); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx, "SELECT UTC_TIMESTAMP(6)").Scan(&base.From); err != nil {
		return err
	}
	base.From = base.From.Truncate(time.Second).Add(time.Second)
	return collectMySQLLogs(ctx, &source, db, dir, base)
}

func readMySQLDumpPosition(path string, base *recoveryBaseline) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		if parts := mysqlDumpPosition.FindStringSubmatch(scan.Text()); parts != nil && mysqlLogName.MatchString(parts[1]) {
			base.Log, base.Position = parts[1], parts[2]
			return nil
		}
	}
	if err := scan.Err(); err != nil {
		return err
	}
	return fmt.Errorf("MySQL dump does not contain a binary log position")
}

func (s *BackupService) archiveMySQL(ctx context.Context, conn *connection.StoredConnection, dir string, base *recoveryBaseline) error {
	if base.Database != conn.DatabaseName || base.Engine != conn.Type {
		return recoveryInputError("Remove the recovery baseline before preparing a different source")
	}
	source, closeSource, err := s.recoverySource(conn)
	if err != nil {
		return err
	}
	defer closeSource()
	db, err := mysqlRecoveryDB(&source)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := collectMySQLLogs(ctx, &source, db, dir, base); err != nil {
		return err
	}
	return writeRecoveryBaseline(dir, base)
}

func collectMySQLLogs(ctx context.Context, conn *connection.StoredConnection, db *sql.DB, dir string, base *recoveryBaseline) error {
	id, _, cutoff, err := mysqlRecoveryIdentity(ctx, db)
	if err != nil {
		return err
	}
	if id != base.SystemID {
		return recoveryInputError("The MySQL server changed; prepare a new recovery baseline")
	}
	rows, err := db.QueryContext(ctx, "SHOW BINARY LOGS")
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil || len(columns) < 2 {
		return fmt.Errorf("unexpected MySQL binary log listing: %w", err)
	}
	start := mysqlLogName.FindStringSubmatch(base.Log)
	if start == nil {
		return fmt.Errorf("invalid baseline log name")
	}
	first, _ := strconv.Atoi(start[2])
	available := make(map[int]bool)
	last := first - 1
	values := make([]any, len(columns))
	for i := range values {
		values[i] = new(sql.RawBytes)
	}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}
		match := mysqlLogName.FindStringSubmatch(string(*values[0].(*sql.RawBytes)))
		if match == nil || match[1] != start[1] {
			return recoveryInputError("MySQL binary log names changed; prepare a new recovery baseline")
		}
		index, _ := strconv.Atoi(match[2])
		available[index] = true
		last = max(index, last)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	previous, _ := mysqlLogIndex(base.LastLog)
	if last < first || last < previous {
		return recoveryInputError("MySQL no longer has the binary logs for this baseline")
	}
	logsDir := filepath.Join(dir, "binlogs")
	if err := os.MkdirAll(logsDir, 0700); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(dir, ".binlogs-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	for index := first; index <= last; index++ {
		name := fmt.Sprintf("%s%06d", start[1], index)
		// The last file in a previous capture was active, so it must be fetched again.
		if index < previous {
			if _, err := os.Stat(filepath.Join(logsDir, name)); err == nil {
				continue
			}
		}
		if !available[index] {
			return recoveryInputError("A required MySQL binary log was purged before it was archived")
		}
		cmd := mysqlRecoveryCommand(ctx, "mysqlbinlog", conn, "--read-from-remote-server", "--raw",
			"--result-file="+staging+string(os.PathSeparator), name)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("archive binary log %s: %w: %s", name, err, output)
		}
		if old, err := os.Stat(filepath.Join(logsDir, name)); err == nil {
			replacement, err := os.Stat(filepath.Join(staging, name))
			if err != nil || replacement.Size() < old.Size() {
				return recoveryInputError("MySQL binary logs were reset; prepare a new recovery baseline")
			}
		}
	}
	entries, err := os.ReadDir(staging)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.Rename(filepath.Join(staging, entry.Name()), filepath.Join(logsDir, entry.Name())); err != nil {
			return err
		}
	}
	base.LastLog = fmt.Sprintf("%s%06d", start[1], last)
	base.Through = cutoff.Truncate(time.Second)
	return nil
}

func mysqlLogIndex(name string) (int, string) {
	match := mysqlLogName.FindStringSubmatch(name)
	if match == nil {
		return 0, ""
	}
	index, _ := strconv.Atoi(match[2])
	return index, match[1]
}

func (s *BackupService) recoverMySQL(ctx context.Context, target *restoreCheckTarget, dir string,
	base *recoveryBaseline, at time.Time, user, password string,
) (*throwawayContainer, error) {
	docker := s.restoreChecks.docker
	container, err := docker.startConfigured(ctx, target.image, target.port, target.env, true,
		map[string]any{"Cmd": []string{"--event-scheduler=OFF", "--skip-log-bin"}})
	if err != nil {
		return nil, err
	}
	ready := false
	defer func() {
		if !ready {
			docker.remove(container.id)
		}
	}()
	target.conn.Host, target.conn.Port = container.host, container.port
	if err := waitForRestoreTarget(ctx, &target.conn); err != nil {
		return nil, err
	}
	dump, err := os.Open(filepath.Join(dir, "baseline.sql"))
	if err != nil {
		return nil, err
	}
	defer dump.Close()
	restore := mysqlRecoveryCommand(ctx, "mariadb", &target.conn, "--binary-mode", "--", base.Database)
	restore.Stdin = dump
	if output, err := restore.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("restore MySQL baseline: %w: %s", err, output)
	}

	first, prefix := mysqlLogIndex(base.Log)
	last, tailPrefix := mysqlLogIndex(base.LastLog)
	if first == 0 || first > last || prefix != tailPrefix {
		return nil, fmt.Errorf("invalid archived MySQL log span")
	}
	args := []string{"--no-defaults", "--verify-binlog-checksum", "--database=" + base.Database,
		"--start-position=" + base.Position, "--stop-datetime=" + at.Format("2006-01-02 15:04:05")}
	for index := first; index <= last; index++ {
		args = append(args, filepath.Join(dir, "binlogs", fmt.Sprintf("%s%06d", prefix, index)))
	}
	sqlFile, err := os.CreateTemp(dir, ".replay-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(sqlFile.Name())
	defer sqlFile.Close()
	replay := exec.CommandContext(ctx, "mysqlbinlog", args...)
	replay.Env = append(os.Environ(), "TZ=UTC")
	replay.Stdout = sqlFile
	var stderr strings.Builder
	replay.Stderr = &stderr
	if err := replay.Run(); err != nil {
		return nil, fmt.Errorf("decode MySQL binary logs: %w: %s", err, stderr.String())
	}
	if _, err := sqlFile.Seek(0, 0); err != nil {
		return nil, err
	}
	apply := mysqlRecoveryCommand(ctx, "mariadb", &target.conn, "--binary-mode", "--", base.Database)
	apply.Stdin = sqlFile
	if output, err := apply.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("replay MySQL binary logs: %w: %s", err, output)
	}
	db, err := openSQLTarget(&target.conn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := grantBranchSQLAccess(ctx, db, false, nil, base.Database, user, password); err != nil {
		return nil, err
	}
	ready = true
	return container, nil
}
