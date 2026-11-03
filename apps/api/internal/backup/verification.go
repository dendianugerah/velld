package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

// VerificationResult summarizes a verification run.
type VerificationResult struct {
	Verification *BackupVerification `json:"verification"`
}

// VerifyBackup runs a verification for a backup synchronously and records the
// result. The HTTP handler uses EnqueueVerification instead so a slow restore
// does not hold the request open.
func (s *BackupService) VerifyBackup(ctx context.Context, backupID string) (*BackupVerification, error) {
	verification, err := s.createPendingVerification(backupID)
	if err != nil {
		return nil, err
	}

	backup, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return s.executeVerification(ctx, verification, backup, conn)
}

// createPendingVerification records a verification in the running state and
// returns it, so callers can report progress before the work finishes.
func (s *BackupService) createPendingVerification(backupID string) (*BackupVerification, error) {
	started := time.Now()
	verification := &BackupVerification{
		ID:        uuid.New().String(),
		BackupID:  backupID,
		Status:    VerificationRunning,
		Method:    MethodChecksum,
		StartedAt: started,
		CreatedAt: started,
	}
	if err := s.backupRepo.CreateVerification(verification); err != nil {
		return nil, fmt.Errorf("failed to record verification start: %w", err)
	}
	return verification, nil
}

func (s *BackupService) executeVerification(ctx context.Context, verification *BackupVerification, backup *Backup, conn *connection.StoredConnection) (*BackupVerification, error) {
	passed, method, details, verifyErr := s.runVerification(ctx, backup, conn)

	completed := time.Now()
	verification.Method = method
	verification.ChecksumMatches = passed
	verification.Details = details
	verification.CompletedAt = &completed
	if verifyErr != nil || !passed {
		verification.Status = VerificationFailed
	} else {
		verification.Status = VerificationPassed
	}

	if err := s.backupRepo.UpdateVerification(verification); err != nil {
		return verification, fmt.Errorf("failed to record verification result: %w", err)
	}

	if verifyErr != nil {
		return verification, verifyErr
	}

	return verification, nil
}

// runVerification performs the strongest check available for the connection.
// Checksum verification always runs first because it is cheap and catches
// corruption; a restore-based check then runs when supported and the checksum
// passed, upgrading the method to MethodRestore.
func (s *BackupService) runVerification(ctx context.Context, backup *Backup, conn *connection.StoredConnection) (bool, string, string, error) {
	matches, details, err := s.verifyChecksum(ctx, backup, conn.UserID)
	if err != nil || !matches {
		return matches, MethodChecksum, details, err
	}

	if !restoreVerifiable(conn.Type) {
		return true, MethodChecksum, details, nil
	}

	path, isTemp, err := s.ensureBackupFileAvailable(backup, conn.UserID)
	if err != nil {
		return true, MethodChecksum, details + " (restore verification skipped: backup unavailable)", nil
	}
	if isTemp {
		defer os.Remove(path)
	}

	restoreMatches, restoreDetails, restoreErr := s.verifyRestore(ctx, backup, conn, path)
	if restoreErr != nil {
		return false, MethodRestore, restoreDetails, restoreErr
	}
	if !restoreMatches {
		return false, MethodRestore, restoreDetails, nil
	}

	return true, MethodRestore, details + "; " + restoreDetails, nil
}

// verifyChecksum re-hashes the backup and compares it with its manifest. It
// returns whether the checksums match and a human-readable detail string.
func (s *BackupService) verifyChecksum(ctx context.Context, backup *Backup, userID uuid.UUID) (bool, string, error) {
	manifest, err := s.backupRepo.GetBackupManifest(backup.ID.String())
	if err != nil {
		return false, "no manifest recorded; cannot verify integrity", fmt.Errorf("failed to load manifest: %w", err)
	}

	reader, cleanup, err := s.openBackupForRead(ctx, backup, userID)
	if err != nil {
		return false, "", err
	}
	defer cleanup()

	cw := NewChecksumWriter(io.Discard)
	if _, err := io.Copy(cw, reader); err != nil {
		return false, "", fmt.Errorf("failed to read backup: %w", err)
	}

	actual := cw.Checksum()
	if actual != manifest.Checksum {
		return false, fmt.Sprintf("checksum mismatch: expected %s, got %s", manifest.Checksum, actual), nil
	}

	return true, "checksum matches manifest", nil
}

// openBackupForRead returns a reader over the backup bytes, preferring the
// local file and falling back to an S3 download into a temp file. The returned
// cleanup removes any temporary file.
func (s *BackupService) openBackupForRead(ctx context.Context, backup *Backup, userID uuid.UUID) (io.Reader, func(), error) {
	path, isTemp, err := s.ensureBackupFileAvailable(backup, userID)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if isTemp {
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: failed to remove temp file %s: %v\n", path, err)
			}
		}
	}

	file, err := os.Open(path)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to open backup: %w", err)
	}

	return file, func() {
		file.Close()
		cleanup()
	}, nil
}

// GetLatestVerification returns the most recent verification for a backup.
func (s *BackupService) GetLatestVerification(backupID string) (*BackupVerification, error) {
	return s.backupRepo.GetLatestVerification(backupID)
}

// AuthorizeBackupOwner returns an error unless the backup's connection belongs
// to userID.
func (s *BackupService) AuthorizeBackupOwner(backup *Backup, userID uuid.UUID) error {
	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		return fmt.Errorf("failed to load connection: %w", err)
	}
	if conn.UserID != userID {
		return fmt.Errorf("not authorized to access this backup")
	}
	return nil
}
