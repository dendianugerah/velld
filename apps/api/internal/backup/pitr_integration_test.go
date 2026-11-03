//go:build integration

package backup

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestPointInTimeRecovery runs a full base-backup + WAL-archive + recover-to-
// timestamp cycle inside a postgres container, then asserts the recovered
// cluster contains the pre-target row and not the post-target one.
func TestPointInTimeRecovery(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available for integration test")
	}

	scriptDir := testdataDir(t)

	cmd := exec.Command("docker", "run", "--rm",
		"--user", "postgres",
		"-v", scriptDir+":/testdata:ro",
		"--entrypoint", "sh",
		"postgres:18-alpine",
		"/testdata/pitr.sh",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PITR exercise failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "PITR_OK") {
		t.Fatalf("expected PITR_OK marker, got:\n%s", out)
	}
}

// testdataDir resolves the package testdata directory from this file's
// location so the test is independent of the working directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve caller path")
	}
	return filepath.Join(filepath.Dir(file), "testdata")
}
