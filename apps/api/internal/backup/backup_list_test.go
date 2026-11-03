package backup

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestListBackupsIncludesVerificationStatus(t *testing.T) {
	repo := NewBackupRepository(newTestDB(t))

	backupID := "77777777-7777-7777-7777-777777777777"
	seedBackup(t, repo, backupID, "/tmp/dump.sql")

	userID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	// Without a verification the field is empty.
	now := time.Now()
	list, _, err := repo.GetAllBackupsWithPagination(BackupListOptions{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("list backups: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(list))
	}
	if list[0].VerificationStatus != "" {
		t.Fatalf("expected no verification status, got %q", list[0].VerificationStatus)
	}

	completed := now
	if err := repo.CreateVerification(&BackupVerification{
		ID:          "ver-1",
		BackupID:    backupID,
		Status:      VerificationPassed,
		Method:      MethodChecksum,
		StartedAt:   now,
		CompletedAt: &completed,
		CreatedAt:   now,
	}); err != nil {
		t.Fatalf("create verification: %v", err)
	}

	list, _, err = repo.GetAllBackupsWithPagination(BackupListOptions{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("list backups: %v", err)
	}
	if list[0].VerificationStatus != VerificationPassed {
		t.Fatalf("verification status = %q, want %q", list[0].VerificationStatus, VerificationPassed)
	}
	if list[0].LastVerifiedAt == "" {
		t.Fatal("expected last_verified_at to be populated")
	}
}
