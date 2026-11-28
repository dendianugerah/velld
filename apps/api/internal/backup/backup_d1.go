package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/dendianugerah/velld/internal/d1"
)

func (s *BackupService) runD1RestoreCheck(ctx context.Context, backup *Backup, conn *connection.StoredConnection) (string, string) {
	path, temporary, err := s.ensureBackupFileAvailable(backup, conn.UserID)
	if err != nil {
		return VerificationFailed, fmt.Sprintf("Backup file is unavailable: %v", err)
	}
	if temporary {
		defer os.Remove(path)
	}
	summary, err := d1.CheckRestore(ctx, path)
	if err != nil {
		return VerificationFailed, err.Error()
	}
	return VerificationPassed, fmt.Sprintf("SQLite integrity check passed: %d tables, %d rows restored", summary.Tables, summary.Rows)
}

func writeD1Backup(conn *connection.StoredConnection, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	file, err := os.CreateTemp(filepath.Dir(path), ".d1-export-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	exportErr := d1.NewClient().Export(ctx, conn.D1Credentials(), file)
	closeErr := file.Close()
	if exportErr != nil {
		return exportErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}
