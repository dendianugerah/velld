package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose"
)

func Init(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		db.Close()
		return nil, err
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// resolveMigrationsDir finds the migrations directory whether the process runs
// from the repository, from apps/api, or from a container image. This keeps the
// binary independent of the working directory.
func resolveMigrationsDir() (string, error) {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
		return "", fmt.Errorf("MIGRATIONS_DIR %q is not a directory", dir)
	}

	candidates := []string{
		filepath.Join("internal", "database", "migrations"),
		filepath.Join("apps", "api", "internal", "database", "migrations"),
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "internal", "database", "migrations"))
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("could not locate migrations directory (looked in %v)", candidates)
}
