package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
)

// mysqlVerifier implements scratchVerifier for MySQL and MariaDB using the
// mysql client.
type mysqlVerifier struct {
	svc *BackupService
}

func (v *mysqlVerifier) createScratch(ctx context.Context, conn *connection.StoredConnection, name string) error {
	return v.exec(ctx, conn, "CREATE DATABASE `"+escapeMySQLIdent(name)+"`")
}

func (v *mysqlVerifier) dropScratch(ctx context.Context, conn *connection.StoredConnection, name string) error {
	return v.exec(ctx, conn, "DROP DATABASE IF EXISTS `"+escapeMySQLIdent(name)+"`")
}

func (v *mysqlVerifier) restoreInto(ctx context.Context, target *connection.StoredConnection, backupPath string) error {
	mysqlBin, err := v.svc.resolveToolPath(target.Type)
	if err != nil {
		return err
	}

	optionFile, cleanup, err := v.credentialFile(target)
	if err != nil {
		return err
	}
	defer cleanup()

	args := []string{
		"-h", target.Host,
		"-P", strconv.Itoa(target.Port),
		"-u", target.Username,
		target.DatabaseName,
	}
	if optionFile != "" {
		args = append([]string{"--defaults-extra-file=" + optionFile}, args...)
	}

	cmd := exec.CommandContext(ctx, mysqlBin, args...)
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()
	cmd.Stdin = file

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return nil
}

func (v *mysqlVerifier) tableRowCounts(ctx context.Context, conn *connection.StoredConnection, database string) (map[string]int64, error) {
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

func (v *mysqlVerifier) listTables(ctx context.Context, conn *connection.StoredConnection, database string) ([]string, error) {
	query := fmt.Sprintf(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = %s ORDER BY table_name",
		quoteMySQLString(database),
	)
	output, err := v.query(ctx, conn, "information_schema", query)
	if err != nil {
		return nil, err
	}
	return splitNonEmptyLines(output), nil
}

func (v *mysqlVerifier) countRows(ctx context.Context, conn *connection.StoredConnection, database, table string) (int64, error) {
	query := "SELECT COUNT(*) FROM `" + escapeMySQLIdent(database) + "`.`" + escapeMySQLIdent(table) + "`"
	output, err := v.query(ctx, conn, database, query)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(output), 10, 64)
}

func (v *mysqlVerifier) query(ctx context.Context, conn *connection.StoredConnection, database, query string) (string, error) {
	mysqlBin, err := v.svc.resolveToolPath(conn.Type)
	if err != nil {
		return "", err
	}

	optionFile, cleanup, err := v.credentialFile(conn)
	if err != nil {
		return "", err
	}
	defer cleanup()

	args := []string{
		"-h", conn.Host,
		"-P", strconv.Itoa(conn.Port),
		"-u", conn.Username,
		"-N", "-B",
		"-e", query,
		database,
	}
	if optionFile != "" {
		args = append([]string{"--defaults-extra-file=" + optionFile}, args...)
	}

	output, err := exec.CommandContext(ctx, mysqlBin, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return string(output), nil
}

// exec runs a statement against the connection's default database.
func (v *mysqlVerifier) exec(ctx context.Context, conn *connection.StoredConnection, statement string) error {
	mysqlBin, err := v.svc.resolveToolPath(conn.Type)
	if err != nil {
		return err
	}

	optionFile, cleanup, err := v.credentialFile(conn)
	if err != nil {
		return err
	}
	defer cleanup()

	args := []string{
		"-h", conn.Host,
		"-P", strconv.Itoa(conn.Port),
		"-u", conn.Username,
		"-e", statement,
	}
	if optionFile != "" {
		args = append([]string{"--defaults-extra-file=" + optionFile}, args...)
	}

	output, err := exec.CommandContext(ctx, mysqlBin, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, trimStderr(output))
	}
	return nil
}

// credentialFile writes a mode 0600 option file when a password is set and
// returns its path plus a cleanup function. The empty path means no password.
func (v *mysqlVerifier) credentialFile(conn *connection.StoredConnection) (string, func(), error) {
	if conn.Password == "" {
		return "", func() {}, nil
	}
	return writeMySQLOptionFile(conn.Password)
}

func escapeMySQLIdent(name string) string {
	return strings.ReplaceAll(name, "`", "``")
}

func quoteMySQLString(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return "'" + replacer.Replace(s) + "'"
}
