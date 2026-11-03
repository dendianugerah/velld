package backup

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

// streamingSupported reports whether a database type can dump to stdout, which
// is required for the no-local-disk path. Redis writes RDB files only.
func streamingSupported(dbType string) bool {
	switch dbType {
	case "postgresql", "mysql", "mariadb", "mongodb":
		return true
	}
	return false
}

// createStreamingBackup dumps straight into S3 without writing the dump to
// local disk. It is used when S3 is configured and the database type supports
// stdout streaming.
func (s *BackupService) createStreamingBackup(ctx context.Context, conn *connection.StoredConnection, dbName string) (*Backup, error) {
	s3Storage, err := s.s3StorageForUser(conn.UserID)
	if err != nil {
		return nil, err
	}

	tempConn := *conn
	tempConn.DatabaseName = dbName

	prepared, err := s.dumpCommand(&tempConn, toStdout())
	if err != nil {
		return nil, err
	}

	backupID := uuid.New()
	filename := fmt.Sprintf("%s_%s.dump", dbName, time.Now().Format("20060102_150405"))
	objectKey := s3Storage.ObjectKeyFor(common.SanitizeConnectionName(conn.Name), filename)

	started := time.Now()
	backup := &Backup{
		ID:           backupID,
		ConnectionID: conn.ID,
		StartedTime:  started,
		Status:       "in_progress",
		Path:         filename,
		CreatedAt:    started,
		UpdatedAt:    started,
	}

	var checksum string
	var size int64

	err = s3Storage.UploadStream(ctx, objectKey, func(dst io.Writer) error {
		cw := NewChecksumWriter(dst)
		if err := prepared.RunToWriter(ctx, cw); err != nil {
			return err
		}
		checksum = cw.Checksum()
		size = cw.Size()
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("streaming backup for %s '%s' failed: %w", conn.Type, dbName, err)
	}

	completed := time.Now()
	backup.Size = size
	backup.Status = "completed"
	backup.CompletedTime = &completed
	backup.S3ObjectKey = &objectKey
	backup.UpdatedAt = completed

	if err := s.backupRepo.CreateBackup(backup); err != nil {
		return nil, fmt.Errorf("failed to save backup: %v", err)
	}

	manifest := &BackupManifest{
		BackupID:          backup.ID.String(),
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          checksum,
		UncompressedSize:  size,
		CompressedSize:    size,
		DurationMs:        completed.Sub(started).Milliseconds(),
		CreatedAt:         completed,
	}
	if err := s.backupRepo.CreateBackupManifest(manifest); err != nil {
		fmt.Printf("Warning: failed to record backup manifest: %v\n", err)
	}

	return backup, nil
}
