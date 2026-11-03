package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func setTestEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	t.Setenv("ENCRYPTION_KEY", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	t.Setenv("ALLOW_REGISTER", "true")
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "test.db"))
	t.Setenv("BACKUP_DIR", t.TempDir())
	t.Setenv("MIGRATIONS_DIR", migrationsDir(t))
}

// migrationsDir resolves the repository migrations directory from this file's
// location, so the test does not depend on the working directory.
func migrationsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve caller path")
	}
	return filepath.Join(filepath.Dir(file), "..", "database", "migrations")
}

func TestBuildServesHealth(t *testing.T) {
	setTestEnv(t)

	app, err := Build()
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	defer app.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatalf("/health route is not registered")
	}
}

func TestBuildProtectsApiRoutes(t *testing.T) {
	setTestEnv(t)

	app, err := Build()
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	defer app.Close(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/api/backups", nil)
	rec := httptest.NewRecorder()
	app.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated /api/backups, got %d", rec.Code)
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("VELD_TEST_KEY", "")
	if got := envOrDefault("VELD_TEST_KEY", "fallback"); got != "fallback" {
		t.Fatalf("got %q, want fallback", got)
	}

	t.Setenv("VELD_TEST_KEY", "value")
	if got := envOrDefault("VELD_TEST_KEY", "fallback"); got != "value" {
		t.Fatalf("got %q, want value", got)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	setTestEnv(t)

	app, err := Build()
	if err != nil {
		t.Fatalf("build app: %v", err)
	}

	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("second close should be a no-op: %v", err)
	}
}

var _ = os.Getenv
