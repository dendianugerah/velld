package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/dendianugerah/velld/internal/connection"
	"github.com/google/uuid"
)

const testEncryptionKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func newVerificationService(t *testing.T) (*BackupService, *BackupRepository) {
	t.Helper()
	db := newTestDB(t)
	repo := NewBackupRepository(db)

	crypto, err := common.NewEncryptionService(testEncryptionKey)
	if err != nil {
		t.Fatalf("create encryption service: %v", err)
	}

	svc := &BackupService{
		backupRepo:  repo,
		connStorage: connection.NewConnectionRepository(db, crypto),
	}
	return svc, repo
}

// seedBackup inserts a connection with encrypted credentials and a completed
// backup pointing at path.
func seedBackup(t *testing.T, repo *BackupRepository, id, path string) {
	t.Helper()
	seedBackupWithType(t, repo, id, path, "postgresql")
}

func seedBackupWithType(t *testing.T, repo *BackupRepository, id, path, dbType string) {
	t.Helper()

	encUser, encPass := encryptForTest(t, "user"), encryptForTest(t, "pass")
	// Deliberately leave nullable columns unset so this doubles as a regression
	// test for NULL-tolerant scanning.
	if _, err := repo.db.Exec(`
		INSERT INTO connections (id, name, type, host, port, username, password, database_name, user_id)
		VALUES ('conn-1', 'primary', ?, 'localhost', 5432, ?, ?, 'app', '99999999-9999-9999-9999-999999999999');
		INSERT INTO backups (id, connection_id, status, path)
		VALUES (?, 'conn-1', 'completed', ?);
	`, dbType, encUser, encPass, id, path); err != nil {
		t.Fatalf("seed backup: %v", err)
	}
}

func encryptForTest(t *testing.T, plaintext string) string {
	t.Helper()
	crypto, err := common.NewEncryptionService(testEncryptionKey)
	if err != nil {
		t.Fatalf("create encryption service: %v", err)
	}
	enc, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return enc
}

func writeBackupFile(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dump.sql")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	return path
}

func TestVerifyChecksumPassesForIntactFile(t *testing.T) {
	svc, repo := newVerificationService(t)

	content := []byte("CREATE TABLE t (id int);\nINSERT INTO t VALUES (1);\n")
	path := writeBackupFile(t, content)
	seedBackup(t, repo, "11111111-1111-1111-1111-111111111111", path)

	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "11111111-1111-1111-1111-111111111111",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex(content),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	backup, err := repo.GetBackup("11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("get backup: %v", err)
	}

	matches, details, err := svc.verifyChecksum(context.Background(), backup, uuid.Nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !matches {
		t.Fatalf("expected checksum to match, details: %s", details)
	}
}

func TestVerifyChecksumDetectsCorruption(t *testing.T) {
	svc, repo := newVerificationService(t)

	path := writeBackupFile(t, []byte("corrupted contents"))
	seedBackup(t, repo, "22222222-2222-2222-2222-222222222222", path)

	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "22222222-2222-2222-2222-222222222222",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex([]byte("original contents")),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	backup, err := repo.GetBackup("22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("get backup: %v", err)
	}

	matches, details, err := svc.verifyChecksum(context.Background(), backup, uuid.Nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches {
		t.Fatal("expected checksum mismatch to be detected")
	}
	if details == "" {
		t.Fatal("expected detail message on mismatch")
	}
}

func TestVerifyBackupRecordsResult(t *testing.T) {
	svc, repo := newVerificationService(t)

	content := []byte("payload")
	path := writeBackupFile(t, content)
	seedBackupWithType(t, repo, "33333333-3333-3333-3333-333333333333", path, "mysql")
	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "33333333-3333-3333-3333-333333333333",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex(content),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	verification, err := svc.VerifyBackup(context.Background(), "33333333-3333-3333-3333-333333333333")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if verification.Status != VerificationPassed {
		t.Fatalf("status = %q, want %q", verification.Status, VerificationPassed)
	}

	latest, err := repo.GetLatestVerification("33333333-3333-3333-3333-333333333333")
	if err != nil {
		t.Fatalf("get latest verification: %v", err)
	}
	if latest.Status != VerificationPassed {
		t.Fatalf("persisted status = %q, want %q", latest.Status, VerificationPassed)
	}
}

func TestVerifyBackupMissingManifestFails(t *testing.T) {
	svc, repo := newVerificationService(t)

	path := writeBackupFile(t, []byte("payload"))
	seedBackupWithType(t, repo, "44444444-4444-4444-4444-444444444444", path, "mysql")

	verification, err := svc.VerifyBackup(context.Background(), "44444444-4444-4444-4444-444444444444")
	if err == nil {
		t.Fatal("expected error when manifest is missing")
	}
	if verification == nil || verification.Status != VerificationFailed {
		t.Fatalf("expected failed verification, got %+v", verification)
	}
}

var _ = sql.ErrNoRows
