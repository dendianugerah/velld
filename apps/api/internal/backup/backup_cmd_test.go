package backup

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

func useTestClientDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	previous := common.CommonBinaryPaths[runtime.GOOS]
	common.CommonBinaryPaths[runtime.GOOS] = []string{dir}
	t.Cleanup(func() { common.CommonBinaryPaths[runtime.GOOS] = previous })
	return dir
}

func writeTestClient(t *testing.T, dir, name, script string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, common.GetPlatformExecutableName(name)), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
}

func TestBackupAndRestorePassCredentialsOutsideArgs(t *testing.T) {
	dir := useTestClientDir(t)
	for _, name := range []string{"pg_dump", "psql", "mysqldump", "mysql", "mongodump", "mongorestore"} {
		writeTestClient(t, dir, name, "")
	}
	password := " test-'#colon: \"secret\"\n\\ends "
	t.Setenv("PGPASSWORD", "inherited")
	t.Setenv("MYSQL_PWD", "inherited")
	s := &BackupService{}
	conn := &connection.StoredConnection{Host: "db.example", Port: 1234, Username: "test-user", DatabaseName: "test_db", Password: password}
	checkArgs := func(cmd *exec.Cmd) {
		t.Helper()
		if cmd == nil {
			t.Fatal("no command was created")
		}
		for _, arg := range cmd.Args {
			if strings.Contains(arg, password) || arg == "--password" {
				t.Fatal("password passed as a command argument")
			}
		}
	}
	checkEnv := func(cmd *exec.Cmd, key, value string) {
		t.Helper()
		checkArgs(cmd)
		var matches []string
		for _, pair := range cmd.Environ() {
			if strings.HasPrefix(pair, key+"=") {
				matches = append(matches, pair)
			}
		}
		if len(matches) != 1 || matches[0] != key+"="+value {
			t.Fatalf("%s was not set to the connection credential", key)
		}
	}
	checkEnv(s.createPgDumpCmd(conn), "PGPASSWORD", password)
	checkEnv(s.createPsqlRestoreCmd(conn), "PGPASSWORD", password)
	for _, dbType := range []string{"mysql", "mariadb"} {
		conn.Type = dbType
		checkEnv(s.createMySQLDumpCmd(conn, "backup.sql"), "MYSQL_PWD", password)
		checkEnv(s.createMySQLRestoreCmd(conn), "MYSQL_PWD", password)
	}
	conn.Password = ""
	checkEnv(s.createMySQLDumpCmd(conn, "backup.sql"), "MYSQL_PWD", "")
	checkEnv(s.createMySQLRestoreCmd(conn), "MYSQL_PWD", "")
	conn.Password = password

	checkConfig := func(cmd *exec.Cmd, cleanup func(), err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		checkArgs(cmd)
		if len(cmd.Args) < 3 || cmd.Args[len(cmd.Args)-2] != "--config" {
			t.Fatal("MongoDB command does not use a config file")
		}
		path := cmd.Args[len(cmd.Args)-1]
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
			t.Fatalf("config file mode = %o, want 0600", info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var decoded string
		if err := json.Unmarshal([]byte(strings.TrimPrefix(string(data), "password: ")), &decoded); err != nil || decoded != password {
			t.Fatal("config file did not preserve the credential")
		}
		cleanup()
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("MongoDB credentials file was not removed")
		}
	}
	checkConfig(s.createMongoDumpCmd(conn, "backup.archive"))
	checkConfig(s.createMongoRestoreCmd(conn, "backup.archive"))
}
