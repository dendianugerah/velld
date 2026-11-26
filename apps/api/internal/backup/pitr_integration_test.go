//go:build integration

package backup

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestDockerPointInTimeRecovery(t *testing.T) {
	t.Run("PostgreSQL", func(t *testing.T) {
		s := newTestBackupService(t)
		conn := saveTestConnection(t, s, connection.StoredConnection{
			Type: "postgresql", Host: dockerTestHost, DatabaseName: "recovery_test", Username: "source_owner", Password: "test_pw",
		})
		archiveRoot := t.TempDir()
		t.Setenv("VELLD_PITR_WAL_DIR", archiveRoot)
		archive := filepath.Join(archiveRoot, conn.ID)
		if err := os.Mkdir(archive, 0777); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(archive, 0777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(archive, "pg_hba.conf"), []byte("local all all trust\nhost all all all scram-sha-256\nhost replication all all scram-sha-256\n"), 0644); err != nil {
			t.Fatal(err)
		}
		container, port := startDockerDatabaseWithOptions(t, "postgres:14-alpine", "5432",
			[]string{"POSTGRES_USER=" + conn.Username, "POSTGRES_PASSWORD=" + conn.Password, "POSTGRES_DB=" + conn.DatabaseName},
			[]string{"--mount=type=bind,src=" + archive + ",dst=/velld-wal"},
			[]string{"-c", "hba_file=/velld-wal/pg_hba.conf", "-c", "archive_mode=on", "-c", `archive_command=cp "%p" "/velld-wal/%f.tmp" && chmod 644 "/velld-wal/%f.tmp" && mv "/velld-wal/%f.tmp" "/velld-wal/%f"`})
		conn.Port = port
		source := prepareRecoveryTestSource(t, container, &conn)
		root, dir := recoveryTestPaths(s, conn)
		base, err := s.captureRecoveryBaseline(t.Context(), root, dir, &conn)
		if err != nil {
			t.Fatal(err)
		}
		execRecoverySQL(t, source, "INSERT INTO events VALUES (1)")
		var target time.Time
		if err := source.QueryRow("SELECT clock_timestamp()").Scan(&target); err != nil {
			t.Fatal(err)
		}
		execRecoverySQL(t, source, "INSERT INTO events VALUES (2)")
		var lastWAL string
		if err := source.QueryRow("SELECT pg_walfile_name(pg_current_wal_lsn())").Scan(&lastWAL); err != nil {
			t.Fatal(err)
		}
		execRecoverySQL(t, source, "SELECT pg_switch_wal()")
		waitForRecoveryTest(t, func() bool {
			_, err := os.Stat(filepath.Join(archive, lastWAL))
			return err == nil
		})
		assertPointInTimeCopy(t, s, &conn, dir, base, target)
		assertRecoveryRows(t, source, "0,1,2")
	})
	t.Run("MySQL", func(t *testing.T) {
		s := newTestBackupService(t)
		conn := saveTestConnection(t, s, connection.StoredConnection{
			Type: "mysql", Host: dockerTestHost, DatabaseName: "recovery_test", Username: "root", Password: "test_pw",
		})
		container, port := startDockerDatabase(t, "mysql:8.0", "3306", "MYSQL_ROOT_HOST=%",
			"MYSQL_ROOT_PASSWORD="+conn.Password, "MYSQL_DATABASE="+conn.DatabaseName)
		conn.Port = port
		source := prepareRecoveryTestSource(t, container, &conn)
		root, dir := recoveryTestPaths(s, conn)
		base, err := s.captureRecoveryBaseline(t.Context(), root, dir, &conn)
		if err != nil {
			t.Fatal(err)
		}
		execRecoverySQL(t, source, "INSERT INTO events VALUES (1)")
		waitForRecoveryTest(t, func() bool { return time.Now().After(base.From.Add(time.Second)) })
		target := time.Now().UTC().Truncate(time.Second)
		execRecoverySQL(t, source, "FLUSH BINARY LOGS")
		execRecoverySQL(t, source, "INSERT INTO events VALUES (2)")
		execRecoverySQL(t, source, "FLUSH BINARY LOGS")
		execRecoverySQL(t, source, "INSERT INTO events VALUES (3)")
		if err := s.archiveMySQL(t.Context(), &conn, dir, base); err != nil {
			t.Fatal(err)
		}
		assertRecoveryRows(t, source, "0,1,2,3")
		if output, err := exec.CommandContext(t.Context(), "docker", "stop", container).CombinedOutput(); err != nil {
			t.Fatalf("stop source: %v: %s", err, output)
		}
		base, err = readRecoveryBaseline(dir)
		if err != nil {
			t.Fatal(err)
		}
		assertPointInTimeCopy(t, s, &conn, dir, base, target)
	})
}

func prepareRecoveryTestSource(t *testing.T, container string, conn *connection.StoredConnection) *sql.DB {
	t.Helper()
	source, err := openSQLTarget(conn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Close() })
	waitForDockerDatabase(t, container, source.PingContext)
	execRecoverySQL(t, source, "CREATE TABLE events (id INT PRIMARY KEY)")
	execRecoverySQL(t, source, "INSERT INTO events VALUES (0)")
	return source
}

func recoveryTestPaths(s *BackupService, conn connection.StoredConnection) (string, string) {
	root := filepath.Join(s.backupDir, ".pitr")
	return root, filepath.Join(root, conn.ID)
}

func assertPointInTimeCopy(t *testing.T, s *BackupService, conn *connection.StoredConnection, dir string, base *recoveryBaseline, target time.Time) {
	t.Helper()
	handler := NewBackupHandler(s)
	ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": uuid.NewString()})
	request := httptest.NewRequestWithContext(ctx, http.MethodGet, "/connections/"+conn.ID+"/pitr", nil)
	r := httptest.NewRecorder()
	handler.PointInTimeRecovery(r, mux.SetURLVars(request, map[string]string{"id": conn.ID}))
	if r.Code != http.StatusNotFound {
		t.Fatalf("another user's recovery baseline returned %d", r.Code)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	result, err := s.recoverAt(ctx, conn, dir, base, target)
	if err != nil {
		t.Fatal(err)
	}
	uri, err := url.Parse(result.ConnectionString)
	if err != nil {
		t.Fatal(err)
	}
	cleanupBranchContainer(t, s.restoreChecks.docker, uri.Port())
	cfg := branchSQLConfig(t, uri, conn.Type)
	recovered, err := openSQLTarget(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	assertRecoveryRows(t, recovered, "0,1")
	query := "CREATE USER forbidden_pitr_admin"
	if conn.Type == "mysql" {
		query = "CREATE USER 'forbidden_pitr_admin'@'%'"
	}
	if _, err := recovered.ExecContext(t.Context(), query); err == nil {
		t.Fatal("recovery account can administer users")
	}
	t.Logf("restored baseline and first write only at %s", target)
}

func assertRecoveryRows(t *testing.T, db *sql.DB, want string) {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), "SELECT id FROM events ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var actual []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatal(err)
		}
		actual = append(actual, value)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(actual, ","); got != want {
		t.Fatalf("rows = %s, want %s", got, want)
	}
}

func execRecoverySQL(t *testing.T, db *sql.DB, stmt string) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), stmt); err != nil {
		t.Fatal(err)
	}
}

func waitForRecoveryTest(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for recovery test prerequisite")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
