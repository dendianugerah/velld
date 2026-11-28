package connection

import (
	"strings"
	"testing"
)

func TestApplyConnectionURL(t *testing.T) {
	for _, test := range []struct {
		raw, engine, host, database string
		port                        int
		ssl                         bool
	}{
		{"postgres://backup%40team:p%3A%40ss@[::1]:5433/app%20data?sslmode=disable", "postgresql", "::1", "app data", 5433, false},
		{"mariadb://backup%40team:p%3A%40ss@db/app", "mysql", "db", "app", 3306, true},
		{"mongodb://backup%40team:p%3A%40ss@db/app?authSource=admin&tls=false", "mongodb", "db", "app", 27017, false},
	} {
		var config ConnectionConfig
		if err := applyConnectionURL(&config, test.raw); err != nil {
			t.Fatal(err)
		}
		if config.Type != test.engine || config.Host != test.host || config.Port != test.port ||
			config.Database != test.database || config.SSL != test.ssl ||
			config.Username != "backup@team" || config.Password != "p:@ss" ||
			(test.engine == "mongodb" && config.MongoAuthSource != "admin") {
			t.Errorf("incorrect parsed connection for %s", test.engine)
		}
	}
}

func TestInvalidConnectionURLDoesNotExposeCredentials(t *testing.T) {
	for _, raw := range []string{
		"postgres://user:private@host:99999",
		"postgres://user:private@host/%ZZ",
		"mysql://user:private@host/database",
		"postgres://user:private@one,two/database",
		"postgres://user:private@host/database?sslmode=%ZZ",
	} {
		err := applyConnectionURL(&ConnectionConfig{Type: "postgresql"}, raw)
		if err == nil || strings.Contains(err.Error(), "private") {
			t.Fatal("expected a URL error without credentials")
		}
	}
}
