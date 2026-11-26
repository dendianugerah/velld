//go:build integration

package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestDockerSchemaDiff(t *testing.T) {
	for _, tc := range []struct {
		engine, version, generatedID, alterLabel string
	}{
		{"postgresql", "16.4", "SERIAL", "ALTER COLUMN label TYPE VARCHAR(70)"},
		{"mysql", "8.0.40", "INT AUTO_INCREMENT", "MODIFY COLUMN label VARCHAR(70)"},
	} {
		t.Run(tc.engine, func(t *testing.T) {
			settings := restoreCheckTargetFor(tc.engine, tc.version, "schema_diff")
			container, port := startDockerDatabase(t, settings.image, strconv.Itoa(settings.port), settings.env...)
			settings.conn.Host, settings.conn.Port = dockerTestHost, port
			live, err := openSQLTarget(&settings.conn)
			if err != nil {
				t.Fatal(err)
			}
			defer live.Close()
			waitForDockerDatabase(t, container, live.PingContext)
			s := newTestBackupService(t)
			conn := saveTestConnection(t, s, settings.conn)

			execSchemaDiffSQL(t, live,
				"CREATE TABLE customers (id INT PRIMARY KEY, label VARCHAR(30) DEFAULT 'old', obsolete INT)",
				"CREATE TABLE should_disappear (id INT PRIMARY KEY)",
				"CREATE TABLE orders (id INT PRIMARY KEY, customer_id INT, CONSTRAINT order_customer FOREIGN KEY (customer_id) REFERENCES customers(id))",
				"CREATE INDEX old_customer_label ON customers(label)",
				"INSERT INTO customers (id, label) VALUES (7, 'preserved')",
				"INSERT INTO orders (id, customer_id) VALUES (1, 7)")
			source := createSchemaDiffBackup(t, s, conn)

			execSchemaDiffSQL(t, live,
				"ALTER TABLE customers "+tc.alterLabel,
				"ALTER TABLE customers ALTER COLUMN label SET DEFAULT 'new'",
				"ALTER TABLE customers DROP COLUMN obsolete",
				"ALTER TABLE customers ADD COLUMN active INT DEFAULT 1 NOT NULL",
				"DROP TABLE should_disappear",
				"CREATE UNIQUE INDEX new_customer_label ON customers(label)",
				"CREATE TABLE invoices (id "+tc.generatedID+" PRIMARY KEY, customer_id INT, CONSTRAINT invoice_customer FOREIGN KEY (customer_id) REFERENCES customers(id), CONSTRAINT invoice_positive CHECK (customer_id > 0))")
			if tc.engine == "postgresql" {
				execSchemaDiffSQL(t, live, "DROP INDEX old_customer_label",
					"CREATE SCHEMA billing", "CREATE TABLE billing.details (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY)")
			} else {
				execSchemaDiffSQL(t, live, "ALTER TABLE customers DROP INDEX old_customer_label")
			}
			target := createSchemaDiffBackup(t, s, conn)

			handler := NewBackupHandler(s)
			call := func(owner uuid.UUID, sourceID string, body string, status int) string {
				t.Helper()
				ctx := context.WithValue(t.Context(), "user", jwt.MapClaims{"user_id": owner.String()})
				request := httptest.NewRequestWithContext(ctx, http.MethodPost, "/backups/"+sourceID+"/schema-diff", strings.NewReader(body))
				rec := httptest.NewRecorder()
				handler.SchemaDiff(rec, mux.SetURLVars(request, map[string]string{"id": sourceID}))
				if rec.Code != status || rec.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("schema diff returned HTTP %d, want %d: %s", rec.Code, status, rec.Body)
				}
				if status != http.StatusOK {
					return ""
				}
				var response struct {
					Data struct {
						Migration string `json:"migration_sql"`
					}
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				return response.Data.Migration
			}
			liveBody := `{"target_live":true}`
			backupBody := `{"target_backup_id":"` + target.ID.String() + `"}`
			call(uuid.New(), source.ID.String(), backupBody, http.StatusNotFound)
			call(conn.UserID, source.ID.String(), `{"target_live":true,"target_backup_id":"x"}`, http.StatusBadRequest)
			migration := call(conn.UserID, source.ID.String(), backupBody, http.StatusOK)
			if fromLive := call(conn.UserID, source.ID.String(), liveBody, http.StatusOK); migration == "" || migration != fromLive {
				t.Fatalf("backup and live diffs did not agree:\nbackup:\n%s\nlive:\n%s", migration, fromLive)
			}
			if unchanged := call(conn.UserID, target.ID.String(), liveBody, http.StatusOK); unchanged != "" {
				t.Fatalf("unchanged schema produced SQL: %s", unchanged)
			}

			// Apply the migration to an independent copy of the source and compare it with the live target.
			copySettings := restoreCheckTargetFor(tc.engine, tc.version, conn.DatabaseName)
			copyName, copyPort := startDockerDatabase(t, copySettings.image, strconv.Itoa(copySettings.port), copySettings.env...)
			copySettings.conn.Host, copySettings.conn.Port = dockerTestHost, copyPort
			copied, err := openSQLTarget(&copySettings.conn)
			if err != nil {
				t.Fatal(err)
			}
			defer copied.Close()
			waitForDockerDatabase(t, copyName, copied.PingContext)
			if _, err := s.restoreFile(t.Context(), &copySettings.conn, source.Path); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "migration.sql")
			if err := os.WriteFile(path, []byte(migration), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := s.restoreFile(t.Context(), &copySettings.conn, path); err != nil {
				t.Fatalf("apply migration: %v\n%s", err, migration)
			}
			actual, err := readDatabaseSchema(t.Context(), &copySettings.conn)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := readDatabaseSchema(t.Context(), &conn)
			if err != nil {
				t.Fatal(err)
			}
			if remaining, err := generateMigration(tc.engine == "postgresql", actual, expected); err != nil || remaining != "" {
				t.Fatalf("schemas differ after applying migration: %v\n%s", err, remaining)
			}
			var label string
			if err := copied.QueryRowContext(t.Context(), "SELECT label FROM customers WHERE id = 7").Scan(&label); err != nil || label != "preserved" {
				t.Fatalf("migration did not preserve customer row: %q, %v", label, err)
			}
			t.Log("backup and live comparisons agreed; applied migration and verified target schema and preserved data")
		})
	}
}

func createSchemaDiffBackup(t *testing.T, s *BackupService, conn connection.StoredConnection) *Backup {
	t.Helper()
	b, err := s.CreateBackup(conn.ID)
	if err != nil || b.Status != "completed" {
		t.Fatalf("create schema diff backup: %v", err)
	}
	return b
}

func execSchemaDiffSQL(t *testing.T, db *sql.DB, statements ...string) {
	t.Helper()
	for _, statement := range statements {
		if _, err := db.ExecContext(t.Context(), statement); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
}
