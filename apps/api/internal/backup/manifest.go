package backup

import (
	"io"
	"os"
	"time"
)

// BackupManifest captures the integrity and provenance details of a finished
// backup. It is what lets Velld prove a backup is complete and later verify a
// restore against the original bytes.
type BackupManifest struct {
	BackupID          string    `json:"backup_id"`
	ChecksumAlgorithm string    `json:"checksum_algorithm"`
	Checksum          string    `json:"checksum"`
	Compression       string    `json:"compression,omitempty"`
	UncompressedSize  int64     `json:"uncompressed_size"`
	CompressedSize    int64     `json:"compressed_size"`
	ServerVersion     string    `json:"server_version,omitempty"`
	ToolVersion       string    `json:"tool_version,omitempty"`
	DurationMs        int64     `json:"duration_ms"`
	CreatedAt         time.Time `json:"created_at"`
}

// recordManifest hashes the finished dump and stores its integrity metadata.
func (s *BackupService) recordManifest(backup *Backup, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	cw := NewChecksumWriter(io.Discard)
	if _, err := io.Copy(cw, file); err != nil {
		return err
	}

	manifest := &BackupManifest{
		BackupID:          backup.ID.String(),
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          cw.Checksum(),
		UncompressedSize:  cw.Size(),
		CompressedSize:    cw.Size(),
		CreatedAt:         time.Now(),
	}

	return s.backupRepo.CreateBackupManifest(manifest)
}
