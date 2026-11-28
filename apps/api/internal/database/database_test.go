package database

import (
	"path/filepath"
	"testing"

	"github.com/pressly/goose"
)

func TestInitAppliesMigrations(t *testing.T) {
	t.Chdir("../..")
	migrations, err := goose.CollectMigrations("internal/database/migrations", 0, goose.MaxVersion)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) == 0 {
		t.Fatal("no SQLite migrations found")
	}
	want := migrations[len(migrations)-1].Version
	path := filepath.Join(t.TempDir(), "velld.db")

	for _, phase := range []string{"fresh database", "restart"} {
		t.Run(phase, func(t *testing.T) {
			db, err := Init(path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			got, err := goose.GetDBVersion(db)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("migration version = %d, want %d", got, want)
			}
		})
	}
}
