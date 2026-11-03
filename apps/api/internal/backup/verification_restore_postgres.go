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

// postgresVerifier implements scratchVerifier for PostgreSQL using psql.
type postgresVerifier struct {
	svc *BackupService
}

func (v *postgresVerifier) createScratch(ctx context.Context, conn *connection.StoredConnection, name string) error {
	return v.svc.psqlExec(ctx, conn, conn.DatabaseName, "CREATE DATABASE "+quotePgIdent(name))
}

func (v *postgresVerifier) dropScratch(ctx context.Context, conn *connection.StoredConnection, name string) error {
	return v.svc.psqlExec(ctx, conn, conn.DatabaseName, "DROP DATABASE IF EXISTS "+quotePgIdent(name))
}

func (v *postgresVerifier) restoreInto(ctx context.Context, target *connection.StoredConnection, backupPath string) error {
	psql, err := v.psqlPath()
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

func (v *postgresVerifier) tableRowCounts(ctx context.Context, conn *connection.StoredConnection, database string) (map[string]int64, error) {
	tables, err := v.listTables(ctx, conn, database)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(tables))
	for _, table := range tables {
		count, err := v.countRows(ctx, conn, database, table)
		if err != nil {
			return nil, err
		}
		counts[table] = count
	}
	return counts, nil
}

func (v *postgresVerifier) listTables(ctx context.Context, conn *connection.StoredConnection, database string) ([]string, error) {
	query := `SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`
	output, err := v.svc.psqlQuery(ctx, conn, database, query)
	if err != nil {
		return nil, err
	}
	return splitNonEmptyLines(string(output)), nil
}

func (v *postgresVerifier) countRows(ctx context.Context, conn *connection.StoredConnection, database, table string) (int64, error) {
	query := "SELECT count(*) FROM " + quotePgIdent(table)
	output, err := v.svc.psqlQuery(ctx, conn, database, query)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
}

func (v *postgresVerifier) psqlPath() (string, error) {
	dir := v.svc.findDatabaseRestorePath("postgresql")
	if dir == "" {
		return "", fmt.Errorf("psql not found; install the PostgreSQL client tools")
	}
	return filepath.Join(dir, common.GetPlatformExecutableName(restoreTools["postgresql"])), nil
}

// psqlExec runs a statement against a database and fails on a non-zero exit.
func (s *BackupService) psqlExec(ctx context.Context, conn *connection.StoredConnection, database, statement string) error {
	_, err := s.psqlQuery(ctx, conn, database, statement)
	return err
}

// psqlQuery runs a query and returns raw, unaligned output.
func (s *BackupService) psqlQuery(ctx context.Context, conn *connection.StoredConnection, database, query string) ([]byte, error) {
	psql, err := s.resolvePsql()
	if err != nil {
		return nil, err
	}

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

func (s *BackupService) resolvePsql() (string, error) {
	dir := s.findDatabaseRestorePath("postgresql")
	if dir == "" {
		return "", fmt.Errorf("psql not found; install the PostgreSQL client tools")
	}
	return filepath.Join(dir, common.GetPlatformExecutableName(restoreTools["postgresql"])), nil
}

func quotePgIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}
