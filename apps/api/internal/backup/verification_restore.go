package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

// restoreVerifiable reports whether we can restore-verify a database type.
// Only PostgreSQL is supported for now; the scratch-database lifecycle and
// row-count comparison are written against psql.
func restoreVerifiable(dbType string) bool {
	return dbType == "postgresql"
}

// scratchDatabaseName returns a predictable, prefixed name for a verification
// database so it is obvious where it came from and easy to clean up.
func scratchDatabaseName(backupID string) string {
	compact := strings.ReplaceAll(backupID, "-", "")
	if len(compact) > 12 {
		compact = compact[:12]
	}
	return "velld_verify_" + compact
}

// verifyRestore restores the backup into a scratch database on the same server,
// compares table and row counts with the live database, and always drops the
// scratch database afterwards. It never writes to the source database.
func (s *BackupService) verifyRestore(ctx context.Context, backup *Backup, conn *connection.StoredConnection, backupPath string) (bool, string, error) {
	psql, err := s.resolvePsql()
	if err != nil {
		return false, "", err
	}

	scratch := scratchDatabaseName(backup.ID.String())

	if err := s.psqlExec(ctx, psql, conn, "", "DROP DATABASE IF EXISTS "+quoteIdent(scratch)); err != nil {
		return false, "", fmt.Errorf("could not clean up any previous scratch database: %w", err)
	}
	if err := s.psqlExec(ctx, psql, conn, "", "CREATE DATABASE "+quoteIdent(scratch)); err != nil {
		return false, "restore verification requires CREATE DATABASE privilege on the server", fmt.Errorf("create scratch database: %w", err)
	}
	defer func() {
		if err := s.psqlExec(context.Background(), psql, conn, "", "DROP DATABASE IF EXISTS "+quoteIdent(scratch)); err != nil {
			fmt.Printf("Warning: failed to drop scratch database %s: %v\n", scratch, err)
		}
	}()

	restoreConn := *conn
	restoreConn.DatabaseName = scratch

	if err := s.restoreIntoDatabase(ctx, conn.Type, &restoreConn, backupPath); err != nil {
		return false, "restore into scratch database failed", err
	}

	sourceCounts, err := s.tableRowCounts(ctx, psql, conn, conn.DatabaseName)
	if err != nil {
		return false, "", fmt.Errorf("count source rows: %w", err)
	}
	restoredCounts, err := s.tableRowCounts(ctx, psql, conn, scratch)
	if err != nil {
		return false, "", fmt.Errorf("count restored rows: %w", err)
	}

	mismatches := compareTableCounts(sourceCounts, restoredCounts)
	if len(mismatches) > 0 {
		return false, strings.Join(mismatches, "; "), nil
	}

	return true, fmt.Sprintf("restored into %s; %d tables match source row counts", scratch, len(sourceCounts)), nil
}

func (s *BackupService) resolvePsql() (string, error) {
	dir := s.findDatabaseRestorePath("postgresql")
	if dir == "" {
		return "", fmt.Errorf("psql not found; install the PostgreSQL client tools")
	}
	return filepath.Join(dir, common.GetPlatformExecutableName(restoreTools["postgresql"])), nil
}

// restoreIntoDatabase restores the dump into target for verification. Unlike
// the user-facing restore it does not stop on the first error: client/server
// version skew (for example a newer pg_dump emitting SET statements the server
// does not know) produces benign errors that would otherwise abort a perfectly
// good restore. Whether the data actually landed is decided by the row-count
// comparison that follows, which is the real test.
func (s *BackupService) restoreIntoDatabase(ctx context.Context, dbType string, target *connection.StoredConnection, backupPath string) error {
	psql, err := s.resolvePsql()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, psql,
		"-h", target.Host,
		"-p", strconv.Itoa(target.Port),
		"-U", target.Username,
		"-d", target.DatabaseName,
		"-f", backupPath,
		"-q",
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+target.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return nil
}

// psqlExec runs a single SQL statement against a database ("" uses the
// connection's default) and returns an error on non-zero exit.
func (s *BackupService) psqlExec(ctx context.Context, psql string, conn *connection.StoredConnection, database, statement string) error {
	args := []string{
		"-h", conn.Host,
		"-p", strconv.Itoa(conn.Port),
		"-U", conn.Username,
		"-v", "ON_ERROR_STOP=1",
		"-t", "-A",
		"-c", statement,
	}
	if database != "" {
		args = append(args, "-d", database)
	} else {
		args = append(args, "-d", conn.DatabaseName)
	}

	cmd := exec.CommandContext(ctx, psql, args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+conn.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return nil
}

// tableRowCounts returns an exact row count per user table. It uses count(*)
// rather than pg_stat_user_tables.n_live_tup because the statistics are
// estimates that can be stale immediately after a restore.
func (s *BackupService) tableRowCounts(ctx context.Context, psql string, conn *connection.StoredConnection, database string) (map[string]int64, error) {
	tables, err := s.listUserTables(ctx, psql, conn, database)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(tables))
	for _, table := range tables {
		count, err := s.countRows(ctx, psql, conn, database, table)
		if err != nil {
			return nil, err
		}
		counts[table] = count
	}
	return counts, nil
}

func (s *BackupService) listUserTables(ctx context.Context, psql string, conn *connection.StoredConnection, database string) ([]string, error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`

	output, err := s.psqlQuery(ctx, psql, conn, database, query)
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, line := range strings.Split(string(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			tables = append(tables, line)
		}
	}
	return tables, nil
}

func (s *BackupService) countRows(ctx context.Context, psql string, conn *connection.StoredConnection, database, table string) (int64, error) {
	query := "SELECT count(*) FROM " + quoteIdent(table)

	output, err := s.psqlQuery(ctx, psql, conn, database, query)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse row count for %q: %w", table, err)
	}
	return count, nil
}

// compareTableCounts returns a human-readable list of differences. Missing
// tables on either side and row-count drift both count as mismatches.
func compareTableCounts(source, restored map[string]int64) []string {
	var mismatches []string

	for table, srcCount := range source {
		restoredCount, ok := restored[table]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("table %q missing after restore", table))
			continue
		}
		if srcCount != restoredCount {
			mismatches = append(mismatches, fmt.Sprintf("table %q: source %d rows, restored %d", table, srcCount, restoredCount))
		}
	}

	for table := range restored {
		if _, ok := source[table]; !ok {
			mismatches = append(mismatches, fmt.Sprintf("table %q present after restore but not in source", table))
		}
	}

	return mismatches
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// psqlQuery runs a query and returns raw, unaligned output.
func (s *BackupService) psqlQuery(ctx context.Context, psql string, conn *connection.StoredConnection, database, query string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, psql,
		"-h", conn.Host,
		"-p", strconv.Itoa(conn.Port),
		"-U", conn.Username,
		"-d", database,
		"-t", "-A",
		"-v", "ON_ERROR_STOP=1",
		"-c", query,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+conn.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return output, nil
}
