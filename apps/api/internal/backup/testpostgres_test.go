//go:build integration

package backup

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// testPostgres is a disposable Postgres instance for integration tests.
type testPostgres struct {
	container string
	host      string
	port      int
	user      string
	password  string
	database  string
}

const (
	testPostgresImage    = "postgres:16-alpine"
	testPostgresUser     = "velld"
	testPostgresPassword = "velld"
	testPostgresDatabase = "velld_test"
)

// startTestPostgres starts a throwaway Postgres container. The test is skipped
// when docker or psql is unavailable so the unit suite still runs everywhere.
func startTestPostgres(t *testing.T) *testPostgres {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available for integration test")
	}
	if _, err := exec.LookPath("psql"); err != nil {
		t.Skip("psql not available for integration test")
	}

	container := fmt.Sprintf("velld-test-pg-%d", time.Now().UnixNano())

	run := exec.Command("docker", "run", "--rm", "-d",
		"--name", container,
		"-e", "POSTGRES_USER="+testPostgresUser,
		"-e", "POSTGRES_PASSWORD="+testPostgresPassword,
		"-e", "POSTGRES_DB="+testPostgresDatabase,
		"-p", "0:5432",
		testPostgresImage,
	)
	if out, err := run.CombinedOutput(); err != nil {
		t.Skipf("could not start postgres container: %v\n%s", err, out)
	}

	pg := &testPostgres{
		container: container,
		host:      "127.0.0.1",
		user:      testPostgresUser,
		password:  testPostgresPassword,
		database:  testPostgresDatabase,
		port:      publishedPort(t, container),
	}

	t.Cleanup(func() {
		exec.Command("docker", "rm", "-f", container).Run()
	})

	pg.waitReady(t)
	return pg
}

func publishedPort(t *testing.T, container string) int {
	t.Helper()

	out, err := exec.Command("docker", "port", container, "5432/tcp").Output()
	if err != nil {
		t.Fatalf("could not read container port: %v", err)
	}

	line := strings.TrimSpace(string(out))
	idx := strings.LastIndex(line, ":")
	if idx < 0 {
		t.Fatalf("unexpected docker port output: %q", line)
	}

	var port int
	if _, err := fmt.Sscanf(line[idx+1:], "%d", &port); err != nil {
		t.Fatalf("could not parse port from %q: %v", line, err)
	}
	return port
}

func (p *testPostgres) waitReady(t *testing.T) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(context.Background(), "psql",
			p.dsn(), "-c", "SELECT 1")
		cmd.Env = append(cmd.Environ(), "PGPASSWORD="+p.password)
		if err := cmd.Run(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}

	t.Fatalf("postgres did not become ready before timeout")
}

func (p *testPostgres) dsn() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		p.user, p.password, p.host, p.port, p.database)
}

func (p *testPostgres) exec(t *testing.T, query string) {
	t.Helper()

	cmd := exec.Command("psql", p.dsn(), "-v", "ON_ERROR_STOP=1", "-c", query)
	cmd.Env = append(cmd.Environ(), "PGPASSWORD="+p.password)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("psql failed: %v\n%s", err, out)
	}
}
