package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
)

// VerificationResult summarizes a verification run.
type VerificationResult struct {
	Verification *BackupVerification `json:"verification"`
}

// VerifyBackup runs a verification for a backup and records the result. It
// currently performs an integrity check: the stored backup is re-hashed and
// compared with the checksum recorded at creation time. This catches
// truncation and corruption, which is the most common way a backup turns out
// to be worthless.
func (s *BackupService) VerifyBackup(ctx context.Context, backupID string) (*BackupVerification, error) {
	backup, err := s.backupRepo.GetBackup(backupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	conn, err := s.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

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

	matches, details, verifyErr := s.verifyChecksum(ctx, backup, conn.UserID)
	completed := time.Now()

	verification.ChecksumMatches = matches
	verification.Details = details
	verification.CompletedAt = &completed
	if verifyErr != nil || !matches {
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
