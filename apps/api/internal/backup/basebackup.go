package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
)

// BaseBackupRequest asks for a physical PostgreSQL base backup, which is the
// starting point for point-in-time recovery.
type BaseBackupRequest struct {
	ConnectionID string `json:"connection_id"`
}

// createPgBaseBackupCmd builds a pg_basebackup that streams a tar archive of the
// cluster to stdout. Streaming to stdout lets the bytes go straight to S3,
// matching the logical dump path.
func (s *BackupService) createPgBaseBackupCmd(conn *connection.StoredConnection) (*preparedCommand, error) {
	dir := common.FindBinaryPath("postgresql", "pg_basebackup")
	if dir == "" {
		return nil, fmt.Errorf("pg_basebackup not found; install the PostgreSQL client tools")
	}
	binPath := filepath.Join(dir, common.GetPlatformExecutableName("pg_basebackup"))

	cmd := exec.Command(binPath,
		"-h", conn.Host,
		"-p", fmt.Sprintf("%d", conn.Port),
		"-U", conn.Username,
		"-D", "-",
		"--format=tar",
		"--wal-method=stream",
		"--no-password",
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", conn.Password))
	return &preparedCommand{cmd: cmd}, nil
}

// CreateBaseBackup streams a physical base backup of a PostgreSQL connection
// into S3 under a timestamped prefix. It returns the object key prefix so a
// later recovery can point at it.
func (s *BackupService) CreateBaseBackup(ctx context.Context, connectionID string) (string, error) {
	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		return "", fmt.Errorf("load connection: %w", err)
	}

	if conn.Type != "postgresql" {
		return "", fmt.Errorf("base backups are only supported for PostgreSQL")
	}

	storage, err := s.s3StorageForUser(conn.UserID)
	if err != nil {
		return "", fmt.Errorf("base backup requires S3 storage: %w", err)
	}

	prepared, err := s.createPgBaseBackupCmd(conn)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/base/%s/base.tar",
		common.SanitizeConnectionName(conn.Name), time.Now().Format("20060102_150405"))

	err = storage.UploadStream(ctx, key, func(dst io.Writer) error {
		return prepared.RunToWriter(ctx, dst)
	})
	if err != nil {
		return "", fmt.Errorf("base backup failed: %w", err)
	}

	return key, nil
}
