package backup

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
)

func TestCheckPostgresVersion(t *testing.T) {
	tests := []struct{ client, server, wantErr string }{
		{"pg_dump (PostgreSQL) 16.1 (vendor release)", "160005", ""},
		{"pg_dump (PostgreSQL) 17.0", "160002", ""},
		{"pg_dump (PostgreSQL) 15.8", "160000", "version mismatch"},
		{"pg_dump (PostgreSQL) 9.5.12", "90600", "version mismatch"},
		{"pg_dump (PostgreSQL) 9.6.1", "90624", ""},
		{"pg_dump (PostgreSQL) 18beta1", "180000", ""},
		{"unexpected output", "160000", "unrecognized"},
		{"pg_dump (PostgreSQL) 16.1", "unexpected output", "invalid PostgreSQL"},
	}
	for _, tc := range tests {
		t.Run(tc.client+"/"+tc.server, func(t *testing.T) {
			err := checkPostgresVersion(tc.client, tc.server)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestPostgresDumpChecksVersionBeforeWritingCompressedOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake client commands are shell scripts")
	}
	dir := useTestClientDir(t)
	t.Setenv("VELLD_TEST_SERVER_VERSION", "160001")
	writeTestClient(t, dir, "psql", "#!/bin/sh\nprintf '%s\\n' \"$VELLD_TEST_SERVER_VERSION\"\n")
	const sql = "CREATE TABLE example (id bigint);"
	writeTestClient(t, dir, "pg_dump", `#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "pg_dump (PostgreSQL) 15.12"
  exit 0
fi
touch "$VELLD_TEST_DUMP_STARTED"
printf '%s' '`+sql+`'
`)
	started := filepath.Join(dir, "dump-started")
	t.Setenv("VELLD_TEST_DUMP_STARTED", started)
	path := filepath.Join(dir, "backup.sql.gz")
	conn := &connection.StoredConnection{Type: "postgresql", Host: "localhost", DatabaseName: "example"}
	s := &BackupService{}
	if err := s.runDump(conn, path); err == nil || !strings.Contains(err.Error(), "version mismatch") {
		t.Fatalf("got %v, want version mismatch error", err)
	}
	for _, missing := range []string{started, path} {
		if _, err := os.Stat(missing); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("file exists after incompatible client: %s", missing)
		}
	}

	t.Setenv("VELLD_TEST_SERVER_VERSION", "150016")
	if err := s.runDump(conn, path); err != nil {
		t.Fatal(err)
	}
	compressed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(compressed) < 2 || compressed[0] != 0x1f || compressed[1] != 0x8b {
		t.Fatal("backup is not in gzip format")
	}
	legacyPath := filepath.Join(dir, "legacy.sql")
	if err := os.WriteFile(legacyPath, []byte(sql), 0600); err != nil {
		t.Fatal(err)
	}
	for _, backupPath := range []string{path, legacyPath} {
		input, closeInput, err := openPostgresBackup(backupPath)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(input)
		closeInput()
		if err != nil || string(data) != sql {
			t.Fatalf("read %s: %q, %v", backupPath, data, err)
		}
	}
	failedPath := filepath.Join(dir, "failed.sql.gz")
	if err := writePostgresBackup(exec.Command("sh", "-c", "printf partial; echo 'dump error' >&2; exit 1"), failedPath); err == nil || !strings.Contains(err.Error(), "dump error") {
		t.Fatalf("got %v, want dump error", err)
	}
	if _, err := os.Stat(failedPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("incomplete dump was not removed")
	}
}
