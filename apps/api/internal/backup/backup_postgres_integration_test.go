package backup

import (
	"database/sql"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func testPostgresConnection(t *testing.T, env string) (*url.URL, connection.StoredConnection) {
	t.Helper()
	dsn := os.Getenv(env)
	if dsn == "" {
		t.Skipf("set %s to a disposable PostgreSQL instance to run this test", env)
	}
	address, err := url.Parse(dsn)
	if err != nil || address.User == nil || address.Scheme != "postgres" {
		t.Fatal("expected a postgres URL including a user")
	}
	password, _ := address.User.Password()
	port, err := strconv.Atoi(address.Port())
	if err != nil {
		t.Fatal("specify the PostgreSQL port explicitly")
	}
	t.Setenv("PGSSLMODE", address.Query().Get("sslmode"))
	return address, connection.StoredConnection{Type: "postgresql", Host: address.Hostname(),
		Port: port, Username: address.User.Username(), Password: password, DatabaseName: strings.TrimPrefix(address.Path, "/")}
}

func TestPostgresBackupRestoreIntegration(t *testing.T) {
	address, config := testPostgresConnection(t, "VELLD_TEST_POSTGRES_DSN")
	admin, err := sql.Open("postgres", address.String())
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	dbNames := []string{"velld_source_", "velld_restore_", "velld_legacy_"}
	for i, prefix := range dbNames {
		name := prefix + strings.ReplaceAll(uuid.NewString(), "-", "")
		if _, err := admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(name)); err != nil {
			t.Fatal(err)
		}
		defer admin.Exec("DROP DATABASE " + pq.QuoteIdentifier(name))
		dbNames[i] = name
	}
	openDB := func(name string) *sql.DB {
		t.Helper()
		dsn := *address
		dsn.Path = "/" + name
		db, err := sql.Open("postgres", dsn.String())
		if err != nil {
			t.Fatal(err)
		}
		return db
	}
	source := openDB(dbNames[0])
	defer source.Close()
	if _, err := source.Exec(`CREATE TABLE examples (id INT PRIMARY KEY, label TEXT);
		INSERT INTO examples VALUES (1, 'first'), (2, 'café'), (3, 'third');`); err != nil {
		t.Fatal(err)
	}
	s := newTestBackupService(t)
	config.DatabaseName = dbNames[0]
	sourceConn := saveTestConnection(t, s, config)
	backup, err := s.CreateBackup(sourceConn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(backup.Path, ".sql.gz") || backup.Size == 0 {
		t.Fatalf("expected compressed backup file, got %s (%d bytes)", backup.Path, backup.Size)
	}

	plain, closeInput, err := openPostgresBackup(backup.Path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(plain)
	closeInput()
	if err != nil {
		t.Fatal(err)
	}
	legacy := *backup
	legacy.ID = uuid.New()
	legacy.Path = filepath.Join(s.backupDir, "legacy.sql")
	legacy.Size = int64(len(data))
	if err := os.WriteFile(legacy.Path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.backupRepo.CreateBackup(&legacy); err != nil {
		t.Fatal(err)
	}

	for i, candidate := range []*Backup{backup, &legacy} {
		config.DatabaseName = dbNames[i+1]
		config.UserID = sourceConn.UserID
		target := saveTestConnection(t, s, config)
		if err := s.RestoreBackup(candidate.ID.String(), target.ID); err != nil {
			t.Fatal(err)
		}
		restored := openDB(target.DatabaseName)
		var rows string
		err := restored.QueryRow("SELECT string_agg(id || ':' || label, ',' ORDER BY id) FROM examples").Scan(&rows)
		_ = restored.Close()
		if err != nil || rows != "1:first,2:café,3:third" {
			t.Fatalf("restored rows from %s = %q, error %v", filepath.Base(candidate.Path), rows, err)
		}
		t.Logf("restored %s: 3 rows matched", filepath.Base(candidate.Path))
	}

	keep := 2
	request := ScheduleBackupRequest{ConnectionID: sourceConn.ID, CronSchedule: "0 0 0 1 1 *", RetentionCount: &keep}
	if err := s.ScheduleBackup(&request); err != nil {
		t.Fatal(err)
	}
	schedule, err := s.backupRepo.GetBackupSchedule(sourceConn.ID)
	if err != nil {
		t.Fatal(err)
	}
	keep = 1
	if err := s.ScheduleBackup(&request); err != nil {
		t.Fatal(err)
	}
	s.executeCronBackup(schedule)
	kept, err := s.backupRepo.GetBackupsByConnectionID(sourceConn.ID)
	if err != nil || len(kept) != 1 || kept[0].ScheduleID == nil || *kept[0].ScheduleID != schedule.ID.String() {
		t.Fatalf("expected only the new scheduled backup to remain, got %v (error %v)", kept, err)
	}
	t.Log("scheduled keep-last-1 retained only the newest backup")
}

func TestPostgresOlderClientIntegration(t *testing.T) {
	_, config := testPostgresConnection(t, "VELLD_TEST_NEWER_POSTGRES_DSN")
	s := newTestBackupService(t)
	conn := saveTestConnection(t, s, config)
	if _, err := s.CreateBackup(conn.ID); err == nil || !strings.Contains(err.Error(), "PostgreSQL version mismatch") {
		t.Fatalf("expected version mismatch against real newer PostgreSQL, got %v", err)
	} else {
		t.Log(err)
	}
	files, err := os.ReadDir(filepath.Join(s.backupDir, common.SanitizeConnectionName(conn.Name)))
	if err != nil || len(files) != 0 {
		t.Fatalf("expected no dump artifact after mismatch; files=%v, error=%v", files, err)
	}
	backups, err := s.backupRepo.GetBackupsByConnectionID(conn.ID)
	if err != nil || len(backups) != 0 {
		t.Fatalf("expected no backup record after mismatch; backups=%v, error=%v", backups, err)
	}
}
