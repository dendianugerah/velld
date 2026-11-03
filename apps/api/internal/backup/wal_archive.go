package backup

import (
	"context"
	"fmt"
	"io"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

// WALCommandConfig describes the archive/restore commands a PostgreSQL server
// should use so its WAL lands in Velld's S3 bucket and can be replayed.
type WALCommandConfig struct {
	ArchiveCommand string `json:"archive_command"`
	RestoreCommand string `json:"restore_command"`
	WALPrefix      string `json:"wal_prefix"`
}

// walObjectKey returns the S3 key for a WAL segment of a connection.
func walObjectKey(connectionName, segment string) string {
	return walPrefix(common.SanitizeConnectionName(connectionName)) + "/" + segment
}

// PushWALSegment stores one WAL segment under the connection's WAL prefix. The
// API exposes this so a server's archive_command can ship segments to Velld.
func (s *BackupService) PushWALSegment(ctx context.Context, connectionID, segment string, content io.Reader) (string, error) {
	if _, err := ParseWALSegment(segment); err != nil {
		return "", err
	}

	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return "", fmt.Errorf("load connection: %w", err)
	}
	if conn.Type != "postgresql" {
		return "", fmt.Errorf("WAL archiving is only supported for PostgreSQL")
	}

	storage, err := s.s3StorageForUser(conn.UserID)
	if err != nil {
		return "", fmt.Errorf("WAL archiving requires S3 storage: %w", err)
	}

	key := walObjectKey(conn.Name, segment)
	if err := storage.UploadStream(ctx, key, func(dst io.Writer) error {
		_, err := io.Copy(dst, content)
		return err
	}); err != nil {
		return "", fmt.Errorf("upload WAL segment: %w", err)
	}

	return key, nil
}

// ArchivedWALSegments lists the WAL segments archived for a connection and
// returns any gaps that would break point-in-time recovery.
func (s *BackupService) ArchivedWALSegments(ctx context.Context, connectionID string, segmentSizeBytes int64) ([]WALSegment, []string, error) {
	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return nil, nil, fmt.Errorf("load connection: %w", err)
	}

	storage, err := s.s3StorageForUser(conn.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("WAL archiving requires S3 storage: %w", err)
	}

	keys, err := storage.ListFilesWithPrefix(ctx, walPrefix(common.SanitizeConnectionName(conn.Name))+"/")
	if err != nil {
		return nil, nil, err
	}

	var segments []WALSegment
	for _, key := range keys {
		name := lastPathSegment(key)
		seg, err := ParseWALSegment(name)
		if err != nil {
			continue
		}
		segments = append(segments, seg)
	}
	SortWALSegments(segments)

	return segments, FindGaps(segments, segmentSizeBytes), nil
}

// buildWALCommandConfig renders the commands a server administrator pastes into
// postgresql.conf. archive_command streams the finished segment to Velld; the
// restore_command copies it back from the mounted archive for recovery.
func buildWALCommandConfig(conn *connection.StoredConnection, apiBaseURL, archivePath string) WALCommandConfig {
	prefix := walPrefix(common.SanitizeConnectionName(conn.Name))

	return WALCommandConfig{
		WALPrefix: prefix,
		ArchiveCommand: fmt.Sprintf(
			`velld wal-push --connection %s --api %s %%p`,
			conn.ID, apiBaseURL,
		),
		RestoreCommand: fmt.Sprintf(
			`cp %s/%%f %%p`,
			archivePath,
		),
	}
}

func lastPathSegment(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '/' {
			return key[i+1:]
		}
	}
	return key
}
