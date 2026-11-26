package backup

import (
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	VerificationPending = "pending"
	VerificationPassed  = "passed"
	VerificationFailed  = "failed"
	VerificationSkipped = "skipped"

	restoreCheckTimeout        = 20 * time.Minute
	restoreCheckReadyTimeout   = 3 * time.Minute
	restoreCheckCleanupTimeout = 30 * time.Second
)

// restoreCheckRunner restores completed backups into throwaway Docker containers, one at a time.
type restoreCheckRunner struct {
	automatic bool
	docker    *dockerEngine
	slot      chan struct{}
	running   sync.WaitGroup
}

func newRestoreCheckRunner() *restoreCheckRunner {
	return &restoreCheckRunner{
		automatic: !strings.EqualFold(os.Getenv("RESTORE_CHECK_ENABLED"), "false"),
		docker:    newDockerEngine(),
		slot:      make(chan struct{}, 1),
	}
}

func (s *BackupService) queueRestoreCheck(backup *Backup) {
	if !s.restoreChecks.automatic {
		return
	}
	pending := VerificationPending
	backup.VerificationStatus = &pending
	if err := s.backupRepo.UpdateBackupVerification(backup.ID.String(), VerificationPending, ""); err != nil {
		fmt.Printf("Warning: failed to queue restore check for backup %s: %v\n", backup.ID, err)
		return
	}
	s.restoreChecks.running.Add(1)
	go func(id string) {
		defer s.restoreChecks.running.Done()
		if _, err := s.VerifyBackup(id); err != nil {
			fmt.Printf("Warning: restore check for backup %s did not run: %v\n", id, err)
		}
	}(backup.ID.String())
}

// ResumeRestoreChecks removes containers left by an interrupted run and re-queues pending checks.
func (s *BackupService) ResumeRestoreChecks() {
	if !s.restoreChecks.automatic {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), restoreCheckCleanupTimeout)
	s.restoreChecks.docker.removeLeftovers(ctx)
	cancel()
	ids, err := s.backupRepo.backupIDsWithVerificationStatus(VerificationPending)
	if err != nil {
		fmt.Printf("Warning: failed to load pending restore checks: %v\n", err)
		return
	}
	for _, id := range ids {
		backup, err := s.backupRepo.GetBackup(id)
		if err == nil {
			s.queueRestoreCheck(backup)
		}
	}
}

// VerifyBackup restores a completed backup into a throwaway container of the same database,
// records whether the restored data matches the backup, and alerts when it does not.
func (s *BackupService) VerifyBackup(backupID string) (*BackupVerification, error) {
	backup, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		return nil, err
	}
	if backup.Status != "completed" {
		return nil, fmt.Errorf("backup %s is %s, only completed backups can be verified", backupID, backup.Status)
	}
	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %v", err)
	}

	s.restoreChecks.slot <- struct{}{}
	status, message := s.runRestoreCheck(backup, conn)
	<-s.restoreChecks.slot

	if err := s.backupRepo.UpdateBackupVerification(backupID, status, message); err != nil {
		return nil, fmt.Errorf("failed to record restore check: %v", err)
	}
	if status == VerificationFailed {
		dbName := backupDatabaseName(backup, conn)
		if err := s.createRestoreCheckFailureNotification(conn.ID, dbName, errors.New(message)); err != nil {
			fmt.Printf("Error creating restore check failure notification: %v\n", err)
		}
	}
	updated, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		return nil, err
	}
	return &updated.BackupVerification, nil
}

func (s *BackupService) runRestoreCheck(backup *Backup, conn *connection.StoredConnection) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), restoreCheckTimeout)
	defer cancel()

	if _, ok := restoreTools[conn.Type]; !ok {
		return VerificationSkipped, fmt.Sprintf("Restore checks are not supported for %s", conn.Type)
	}
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	err := s.restoreChecks.docker.ping(pingCtx)
	pingCancel()
	if err != nil {
		return VerificationSkipped, "Docker is not reachable from the API; mount /var/run/docker.sock to enable restore checks"
	}

	filePath, isTemp, err := s.ensureBackupFileAvailable(backup, conn.UserID)
	if err != nil {
		return VerificationFailed, fmt.Sprintf("Backup file is unavailable: %v", err)
	}
	if isTemp {
		defer os.Remove(filePath)
	}

	dump, err := inspectDump(conn.Type, filePath)
	if err != nil {
		return VerificationFailed, fmt.Sprintf("Backup file is unreadable: %v", err)
	}

	target := restoreCheckTargetFor(conn.Type, dump.serverVersion, backupDatabaseName(backup, conn))
	container, err := s.restoreChecks.docker.start(ctx, target.image, target.port, target.env, false)
	if err != nil {
		return VerificationFailed, fmt.Sprintf("Could not start %s restore check container: %v", target.image, err)
	}
	defer s.restoreChecks.docker.remove(container.id)
	target.conn.Host, target.conn.Port = container.host, container.port
	summary, err := s.restoreIntoTarget(ctx, &target, filePath, dump)
	if err != nil {
		return VerificationFailed, err.Error()
	}
	return VerificationPassed, fmt.Sprintf("Restored %s into %s", summary, target.image)
}

func (s *BackupService) restoreIntoTarget(ctx context.Context, target *restoreCheckTarget, filePath string, dump dumpSummary) (string, error) {
	readyCtx, readyCancel := context.WithTimeout(ctx, restoreCheckReadyTimeout)
	defer readyCancel()
	if err := waitForRestoreTarget(readyCtx, &target.conn); err != nil {
		return "", fmt.Errorf("%s restore check container did not become ready: %w", target.image, err)
	}

	if err := createReferencedRoles(ctx, &target.conn, dump.roles); err != nil {
		return "", fmt.Errorf("Could not prepare %s for the restore: %w", target.image, err)
	}

	output, err := s.restoreFile(ctx, &target.conn, filePath)
	if err != nil {
		return "", fmt.Errorf("Restore into %s failed: %w", target.image, err)
	}

	summary, err := confirmRestoredData(ctx, &target.conn, dump, output)
	if err != nil {
		return "", fmt.Errorf("Restore into %s finished but the data does not match the backup: %w", target.image, err)
	}
	return summary, nil
}

func backupDatabaseName(backup *Backup, conn *connection.StoredConnection) string {
	const timestampSuffix = len("_20060102_150405")
	name := filepath.Base(backup.Path)
	if i := strings.LastIndex(name, "_"+backup.ID.String()); i > timestampSuffix {
		if _, err := time.Parse("20060102_150405", name[i-timestampSuffix+1:i]); err == nil {
			return name[:i-timestampSuffix]
		}
	}
	return conn.DatabaseName
}

type dumpSummary struct {
	serverVersion string
	tables        int
	rows          int64
	roles         []string
}

var (
	pgDumpServerVersion = regexp.MustCompile(`^-- Dumped from database version (\S+)`)
	mysqlDumpVersion    = regexp.MustCompile(`^-- Server version\s+(\S+)`)
	createTableLine     = regexp.MustCompile("^CREATE (UNLOGGED )?TABLE ")
	pgRoleReference     = regexp.MustCompile(`^(?:ALTER .* OWNER TO|GRANT .* TO) (.+);$`)
)

func inspectDump(dbType, filePath string) (dumpSummary, error) {
	var summary dumpSummary
	var input io.Reader
	switch dbType {
	case "postgresql":
		reader, closeReader, err := openPostgresBackup(filePath)
		if err != nil {
			return summary, err
		}
		defer closeReader()
		input = reader
	case "mysql", "mariadb":
		file, err := os.Open(filePath)
		if err != nil {
			return summary, err
		}
		defer file.Close()
		input = file
	default:
		info, err := os.Stat(filePath)
		if err == nil && info.Size() == 0 {
			err = errors.New("backup file is empty")
		}
		return summary, err
	}

	reader := bufio.NewReaderSize(input, 1<<20)
	inCopy := false
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			line = strings.TrimRight(line, "\r\n")
			switch {
			case inCopy:
				if line == `\.` {
					inCopy = false
				} else {
					summary.rows++
				}
			case strings.HasPrefix(line, "COPY ") && strings.HasSuffix(line, "FROM stdin;"):
				inCopy = true
			case createTableLine.MatchString(line):
				summary.tables++
			case dbType == "postgresql" && pgRoleReference.MatchString(line):
				for _, role := range strings.Split(pgRoleReference.FindStringSubmatch(line)[1], ", ") {
					if strings.HasPrefix(role, `"`) {
						role = strings.ReplaceAll(strings.Trim(role, `"`), `""`, `"`)
					}
					if role != "PUBLIC" && !slices.Contains(summary.roles, role) {
						summary.roles = append(summary.roles, role)
					}
				}
			case summary.serverVersion == "" && strings.HasPrefix(line, "--"):
				if m := pgDumpServerVersion.FindStringSubmatch(line); m != nil {
					summary.serverVersion = m[1]
				} else if m := mysqlDumpVersion.FindStringSubmatch(line); m != nil {
					summary.serverVersion = m[1]
				}
			}
		}
		if err == io.EOF {
			return summary, nil
		}
		if err != nil {
			return summary, err
		}
	}
}

type restoreCheckTarget struct {
	image string
	port  int
	env   []string
	conn  connection.StoredConnection
}

func restoreCheckTargetFor(dbType, serverVersion, dbName string) restoreCheckTarget {
	password := randomPassword()
	conn := connection.StoredConnection{Type: dbType, DatabaseName: dbName, Password: password}
	majorMinor := func(version string) string {
		parts := strings.SplitN(strings.SplitN(version, "-", 2)[0], ".", 3)
		if len(parts) < 2 {
			return ""
		}
		return parts[0] + "." + parts[1]
	}

	switch dbType {
	case "postgresql":
		conn.Username = "velld"
		image := "postgres:17"
		if major := strings.SplitN(serverVersion, ".", 2)[0]; major != "" {
			if _, err := strconv.Atoi(major); err == nil {
				image = "postgres:" + major
			}
		}
		return restoreCheckTarget{image: image, port: 5432, conn: conn,
			env: []string{"POSTGRES_USER=velld", "POSTGRES_PASSWORD=" + password, "POSTGRES_DB=" + dbName}}
	case "mysql", "mariadb":
		conn.Username = "root"
		image := "mysql:8.4"
		if strings.Contains(strings.ToLower(serverVersion), "mariadb") || (dbType == "mariadb" && serverVersion == "") {
			image = "mariadb:11.4"
			if version := majorMinor(serverVersion); version != "" {
				image = "mariadb:" + version
			}
		} else if version := majorMinor(serverVersion); version != "" {
			image = "mysql:" + version
		}
		return restoreCheckTarget{image: image, port: 3306, conn: conn,
			env: []string{"MYSQL_ROOT_PASSWORD=" + password, "MYSQL_ROOT_HOST=%", "MYSQL_DATABASE=" + dbName,
				"MARIADB_ROOT_PASSWORD=" + password, "MARIADB_ROOT_HOST=%", "MARIADB_DATABASE=" + dbName}}
	default:
		conn.Password = ""
		return restoreCheckTarget{image: "mongo:7.0", port: 27017, conn: conn}
	}
}

func randomPassword() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func openSQLTarget(conn *connection.StoredConnection) (*sql.DB, error) {
	address := net.JoinHostPort(conn.Host, strconv.Itoa(conn.Port))
	if conn.Type == "postgresql" {
		dsn := (&url.URL{Scheme: "postgres", User: url.UserPassword(conn.Username, conn.Password),
			Host: address, Path: "/" + conn.DatabaseName, RawQuery: "sslmode=disable"}).String()
		return sql.Open("postgres", dsn)
	}
	config := mysql.NewConfig()
	config.User, config.Passwd, config.DBName = conn.Username, conn.Password, conn.DatabaseName
	config.Net, config.Addr, config.AllowNativePasswords = "tcp", address, true
	config.Logger = mysql.Logger(log.New(io.Discard, "", 0))
	connector, err := mysql.NewConnector(config)
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(connector), nil
}

func openMongoTarget(ctx context.Context, conn *connection.StoredConnection) (*mongo.Client, error) {
	opts := options.Client().
		SetHosts([]string{net.JoinHostPort(conn.Host, strconv.Itoa(conn.Port))}).
		SetDirect(true).SetServerSelectionTimeout(5 * time.Second)
	if conn.Username != "" {
		opts.SetAuth(options.Credential{
			Username: conn.Username, Password: conn.Password, AuthSource: conn.MongoAuthSource,
		})
	}
	return mongo.Connect(ctx, opts)
}

func waitForRestoreTarget(ctx context.Context, conn *connection.StoredConnection) error {
	ping := func() error {
		attemptCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if conn.Type == "mongodb" {
			client, err := openMongoTarget(attemptCtx, conn)
			if err != nil {
				return err
			}
			defer client.Disconnect(context.Background())
			return client.Ping(attemptCtx, nil)
		}
		db, err := openSQLTarget(conn)
		if err != nil {
			return err
		}
		defer db.Close()
		return db.PingContext(attemptCtx)
	}
	for {
		err := ping()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(time.Second):
		}
	}
}

// createReferencedRoles creates the roles a PostgreSQL dump assigns ownership or grants to.
func createReferencedRoles(ctx context.Context, conn *connection.StoredConnection, roles []string) error {
	if len(roles) == 0 {
		return nil
	}
	db, err := openSQLTarget(conn)
	if err != nil {
		return err
	}
	defer db.Close()
	for _, role := range roles {
		if role == conn.Username {
			continue
		}
		if _, err := db.ExecContext(ctx, "CREATE ROLE "+pq.QuoteIdentifier(role)); err != nil {
			return fmt.Errorf("create role %s: %w", role, err)
		}
	}
	return nil
}

var mongoRestoreCounts = regexp.MustCompile(`(\d+) document\(s\) restored successfully\. (\d+) document\(s\) failed to restore`)

func confirmRestoredData(ctx context.Context, conn *connection.StoredConnection, dump dumpSummary, output []byte) (string, error) {
	if conn.Type == "mongodb" {
		return confirmRestoredMongo(ctx, conn, output)
	}
	db, err := openSQLTarget(conn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var tables int
	var rows int64
	if conn.Type == "postgresql" {
		const userTables = `FROM information_schema.tables WHERE table_type = 'BASE TABLE'
			AND table_schema NOT IN ('pg_catalog', 'information_schema')`
		if err := db.QueryRowContext(ctx, "SELECT count(*) "+userTables).Scan(&tables); err != nil {
			return "", fmt.Errorf("count restored tables: %w", err)
		}
		if err := db.QueryRowContext(ctx, `SELECT coalesce(sum((xpath('/row/c/text()', query_to_xml(
			format('SELECT count(*) AS c FROM ONLY %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint), 0) `+
			userTables).Scan(&rows); err != nil {
			return "", fmt.Errorf("count restored rows: %w", err)
		}
	} else if err := db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`, conn.DatabaseName).Scan(&tables); err != nil {
		return "", fmt.Errorf("count restored tables: %w", err)
	}

	if tables < dump.tables {
		return "", fmt.Errorf("backup defines %d tables but %d were restored", dump.tables, tables)
	}
	if conn.Type == "postgresql" {
		if rows < dump.rows {
			return "", fmt.Errorf("backup holds %d rows but %d were restored", dump.rows, rows)
		}
		return fmt.Sprintf("%d tables and %d rows", tables, rows), nil
	}
	return fmt.Sprintf("%d tables", tables), nil
}

func confirmRestoredMongo(ctx context.Context, conn *connection.StoredConnection, output []byte) (string, error) {
	client, err := openMongoTarget(ctx, conn)
	if err != nil {
		return "", err
	}
	defer client.Disconnect(context.Background())

	db := client.Database(conn.DatabaseName)
	names, err := db.ListCollectionNames(ctx, bson.D{{Key: "type", Value: "collection"}})
	if err != nil {
		return "", fmt.Errorf("list restored collections: %w", err)
	}
	var collections int
	var documents int64
	for _, name := range names {
		if strings.HasPrefix(name, "system.") {
			continue
		}
		count, err := db.Collection(name).CountDocuments(ctx, bson.D{})
		if err != nil {
			return "", fmt.Errorf("count documents in %s: %w", name, err)
		}
		collections++
		documents += count
	}

	if m := mongoRestoreCounts.FindSubmatch(output); m != nil {
		restored, _ := strconv.ParseInt(string(m[1]), 10, 64)
		if failed, _ := strconv.ParseInt(string(m[2]), 10, 64); failed > 0 {
			return "", fmt.Errorf("%d documents failed to restore", failed)
		}
		if documents != restored {
			return "", fmt.Errorf("mongorestore restored %d documents but %d are readable", restored, documents)
		}
	}
	return fmt.Sprintf("%d collections and %d documents", collections, documents), nil
}
