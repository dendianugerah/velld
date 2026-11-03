package backup

import (
	"context"
	"fmt"
	"strings"

	"github.com/dendianugerah/velld/internal/connection"
)

// scratchVerifier abstracts the engine-specific parts of restore verification:
// creating and dropping a scratch database, restoring into it, and counting
// rows per table. verifyRestore owns the engine-neutral flow; each database
// type supplies a scratchVerifier.
type scratchVerifier interface {
	createScratch(ctx context.Context, conn *connection.StoredConnection, name string) error
	dropScratch(ctx context.Context, conn *connection.StoredConnection, name string) error
	restoreInto(ctx context.Context, target *connection.StoredConnection, backupPath string) error
	tableRowCounts(ctx context.Context, conn *connection.StoredConnection, database string) (map[string]int64, error)
}

// restoreVerifiable reports whether restore verification is supported for a
// database type.
func restoreVerifiable(dbType string) bool {
	return scratchVerifierFor(dbType) != nil
}

func scratchVerifierFor(dbType string) scratchVerifier {
	switch dbType {
	case "postgresql":
		return &postgresVerifier{svc: nil}
	case "mysql", "mariadb":
		return &mysqlVerifier{svc: nil}
	}
	return nil
}

// verifyRestore restores the backup into a scratch database on the same server,
// compares per-table row counts with the live database, and always drops the
// scratch database. It never writes to the source database.
func (s *BackupService) verifyRestore(ctx context.Context, backup *Backup, conn *connection.StoredConnection, backupPath string) (bool, string, error) {
	verifier := scratchVerifierFor(conn.Type)
	if verifier == nil {
		return false, "", fmt.Errorf("restore verification not supported for %s", conn.Type)
	}
	attachVerifier(verifier, s)

	scratch := scratchDatabaseName(backup.ID.String())

	if err := verifier.dropScratch(ctx, conn, scratch); err != nil {
		return false, "", fmt.Errorf("could not clean up any previous scratch database: %w", err)
	}
	if err := verifier.createScratch(ctx, conn, scratch); err != nil {
		return false, "restore verification requires permission to create a scratch database on the server", fmt.Errorf("create scratch database: %w", err)
	}
	defer func() {
		if err := verifier.dropScratch(context.Background(), conn, scratch); err != nil {
			fmt.Printf("Warning: failed to drop scratch database %s: %v\n", scratch, err)
		}
	}()

	restoreConn := *conn
	restoreConn.DatabaseName = scratch

	if err := verifier.restoreInto(ctx, &restoreConn, backupPath); err != nil {
		return false, "restore into scratch database failed", err
	}

	sourceCounts, err := verifier.tableRowCounts(ctx, conn, conn.DatabaseName)
	if err != nil {
		return false, "", fmt.Errorf("count source rows: %w", err)
	}
	restoredCounts, err := verifier.tableRowCounts(ctx, conn, scratch)
	if err != nil {
		return false, "", fmt.Errorf("count restored rows: %w", err)
	}

	mismatches := compareTableCounts(sourceCounts, restoredCounts)
	if len(mismatches) > 0 {
		return false, strings.Join(mismatches, "; "), nil
	}

	return true, fmt.Sprintf("restored into %s; %d tables match source row counts", scratch, len(sourceCounts)), nil
}

// attachVerifier wires the service into the verifier implementation, which
// keeps the interface free of the service dependency.
func attachVerifier(verifier scratchVerifier, svc *BackupService) {
	switch v := verifier.(type) {
	case *postgresVerifier:
		v.svc = svc
	case *mysqlVerifier:
		v.svc = svc
	}
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

// restoreToolingAvailable reports whether the client tools needed to restore
// a database type are installed. When they are not, verification falls back to
// the checksum check rather than failing.
func (s *BackupService) restoreToolingAvailable(dbType string) error {
	tool, ok := restoreTools[dbType]
	if !ok {
		return fmt.Errorf("no restore tool for %s", dbType)
	}
	if s.findDatabaseRestorePath(dbType) == "" {
		return fmt.Errorf("%s not found", tool)
	}
	return nil
}
