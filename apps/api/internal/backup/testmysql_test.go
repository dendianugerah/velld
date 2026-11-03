package backup

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// testMySQL is a disposable MySQL instance for integration tests.
type testMySQL struct {
	container string
	host      string
	port      int
	user      string
	password  string
	database  string
}

const (
	testMySQLImage    = "mysql:8.0"
	testMySQLUser     = "velld"
	testMySQLPassword = "velldpass"
	testMySQLDatabase = "velld_test"
)

// startTestMySQL starts a throwaway MySQL container, skipping the test when
// docker or the mysql client is unavailable.
func startTestMySQL(t *testing.T) *testMySQL {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available for integration test")
	}
	if _, err := exec.LookPath("mysql"); err != nil {
		t.Skip("mysql client not available for integration test")
	}

	container := fmt.Sprintf("velld-test-mysql-%d", time.Now().UnixNano())

	run := exec.Command("docker", "run", "--rm", "-d",
		"--name", container,
		"-e", "MYSQL_ROOT_PASSWORD=rootpass",
		"-e", "MYSQL_DATABASE="+testMySQLDatabase,
		"-e", "MYSQL_USER="+testMySQLUser,
		"-e", "MYSQL_PASSWORD="+testMySQLPassword,
		"-p", "0:3306",
		testMySQLImage,
	)
	if out, err := run.CombinedOutput(); err != nil {
		t.Skipf("could not start mysql container: %v\n%s", err, out)
	}

	my := &testMySQL{
		container: container,
		host:      "127.0.0.1",
		user:      testMySQLUser,
		password:  testMySQLPassword,
		database:  testMySQLDatabase,
		port:      publishedMySQLPort(t, container),
	}

	t.Cleanup(func() {
		exec.Command("docker", "rm", "-f", container).Run()
	})

	my.waitReady(t)
	return my
}

func publishedMySQLPort(t *testing.T, container string) int {
	t.Helper()

	out, err := exec.Command("docker", "port", container, "3306/tcp").Output()
	if err != nil {
		t.Fatalf("read container port: %v", err)
	}
	line := strings.TrimSpace(string(out))
	idx := strings.LastIndex(line, ":")
	if idx < 0 {
		t.Fatalf("unexpected docker port output: %q", line)
	}
	var port int
	if _, err := fmt.Sscanf(line[idx+1:], "%d", &port); err != nil {
		t.Fatalf("parse port from %q: %v", line, err)
	}
	return port
}

func (m *testMySQL) waitReady(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(context.Background(), "mysql",
			"-h", m.host, "-P", fmt.Sprintf("%d", m.port),
			"-u", m.user, "-p"+m.password, "-e", "SELECT 1")
		if err := cmd.Run(); err == nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("mysql did not become ready before timeout")
}

func (m *testMySQL) exec(t *testing.T, statement string) {
	t.Helper()

	cmd := exec.Command("mysql",
		"-h", m.host, "-P", fmt.Sprintf("%d", m.port),
		"-u", m.user, "-p"+m.password, m.database, "-e", statement)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mysql exec failed: %v\n%s", err, out)
	}
}
