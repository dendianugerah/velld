package backup

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

func assertRestoreCheck(t *testing.T, s *BackupService, backupID, wantStatus, wantMessage string) {
	t.Helper()
	result, err := s.VerifyBackup(backupID)
	if err != nil {
		t.Fatalf("verify backup: %v", err)
	}
	if result.VerificationStatus == nil || *result.VerificationStatus != wantStatus ||
		result.VerificationMessage == nil || !strings.Contains(*result.VerificationMessage, wantMessage) ||
		result.VerifiedAt == nil {
		t.Fatalf("restore check = %s: %s, want %s containing %q", deref(result.VerificationStatus),
			deref(result.VerificationMessage), wantStatus, wantMessage)
	}
	listed, _, err := s.backupRepo.GetAllBackupsWithPagination(BackupListOptions{UserID: mustBackupOwner(t, s, backupID), Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	for _, backup := range listed {
		if backup.ID.String() == backupID {
			if backup.VerificationStatus == nil || *backup.VerificationStatus != wantStatus {
				t.Fatalf("backup history shows restore check %v, want %s", backup.VerificationStatus, wantStatus)
			}
			t.Logf("restore check %s: %s", wantStatus, *result.VerificationMessage)
			return
		}
	}
	t.Fatalf("backup %s missing from backup history", backupID)
}

func mustBackupOwner(t *testing.T, s *BackupService, backupID string) uuid.UUID {
	t.Helper()
	backup, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		t.Fatal(err)
	}
	return conn.UserID
}

func TestInspectDump(t *testing.T) {
	dir := t.TempDir()
	pgPath := filepath.Join(dir, "shop.sql.gz")
	file, err := os.Create(pgPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	_, _ = writer.Write([]byte("-- Dumped from database version 16.4 (Debian 16.4-1)\n" +
		"CREATE TABLE public.a (id integer);\nCREATE UNLOGGED TABLE public.b (id integer);\n" +
		"ALTER TABLE public.a OWNER TO app;\nGRANT SELECT ON TABLE public.a TO \"Read Only\", PUBLIC;\n" +
		"COPY public.a (id) FROM stdin;\n1\n2\n\\.\nCOPY public.b (id) FROM stdin;\n3\n\\.\n"))
	_ = writer.Close()
	_ = file.Close()
	pg, err := inspectDump("postgresql", pgPath)
	if err != nil || pg.serverVersion != "16.4" || pg.tables != 2 || pg.rows != 3 ||
		!slices.Equal(pg.roles, []string{"app", "Read Only"}) {
		t.Fatalf("postgres dump summary = %+v, %v", pg, err)
	}

	mysqlPath := filepath.Join(dir, "shop.sql")
	if err := os.WriteFile(mysqlPath, []byte("-- Server version\t10.11.6-MariaDB-0+deb12u1\n"+
		"CREATE TABLE `a` (id int);\n/*!50001 CREATE VIEW `v` AS SELECT 1 */;\nINSERT INTO `a` VALUES (1),(2);\n"), 0600); err != nil {
		t.Fatal(err)
	}
	my, err := inspectDump("mysql", mysqlPath)
	if err != nil || my.serverVersion != "10.11.6-MariaDB-0+deb12u1" || my.tables != 1 || len(my.roles) != 0 {
		t.Fatalf("mysql dump summary = %+v, %v", my, err)
	}

	empty := filepath.Join(dir, "shop.archive")
	if err := os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectDump("mongodb", empty); err == nil {
		t.Fatal("expected an empty MongoDB archive to be rejected")
	}
}

func TestRestoreCheckTargetMatchesSourceServer(t *testing.T) {
	for _, tc := range []struct{ dbType, version, image string }{
		{"postgresql", "16.4 (Debian 16.4-1)", "postgres:16"},
		{"postgresql", "", "postgres:17"},
		{"mysql", "8.0.40", "mysql:8.0"},
		{"mysql", "10.11.6-MariaDB-0+deb12u1", "mariadb:10.11"},
		{"mariadb", "", "mariadb:11.4"},
		{"mysql", "", "mysql:8.4"},
		{"mongodb", "", "mongo:7.0"},
	} {
		target := restoreCheckTargetFor(tc.dbType, tc.version, "shop")
		if target.image != tc.image || target.conn.DatabaseName != "shop" || target.conn.Type != tc.dbType {
			t.Errorf("%s %q: target %s/%+v, want %s", tc.dbType, tc.version, target.image, target.conn, tc.image)
		}
	}
}

func TestBackupDatabaseName(t *testing.T) {
	id := uuid.New()
	conn := &connection.StoredConnection{DatabaseName: "fallback"}
	backup := &Backup{ID: id, Path: "/backups/conn/my_shop_db_20260101_120000_" + id.String() + ".sql.gz"}
	if got := backupDatabaseName(backup, conn); got != "my_shop_db" {
		t.Fatalf("database name = %q", got)
	}
	backup.Path = "/backups/legacy.sql"
	if got := backupDatabaseName(backup, conn); got != "fallback" {
		t.Fatalf("legacy database name = %q", got)
	}
}

func TestRestoreCheckSkippedWithoutDocker(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix://"+filepath.Join(t.TempDir(), "missing.sock"))
	s := newTestBackupService(t)
	s.restoreChecks.docker = newDockerEngine()
	conn := saveTestConnection(t, s, connection.StoredConnection{Type: "postgresql", DatabaseName: "shop"})
	id := uuid.New()
	now := time.Now()
	if err := s.backupRepo.CreateBackup(&Backup{ID: id, ConnectionID: conn.ID, Status: "completed",
		Path: filepath.Join(s.backupDir, "missing.sql.gz"), StartedTime: now, CompletedTime: &now,
		CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	assertRestoreCheck(t, s, id.String(), VerificationSkipped, "Docker is not reachable")
	if !strings.Contains(*mustVerification(t, s, id.String()).VerificationMessage, "docker.sock") {
		t.Fatal("skip message should explain how to enable restore checks")
	}
}

func mustVerification(t *testing.T, s *BackupService, id string) *BackupVerification {
	t.Helper()
	backup, err := s.backupRepo.GetBackup(id)
	if err != nil {
		t.Fatal(err)
	}
	return &backup.BackupVerification
}

func deref(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}
