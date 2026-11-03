package backup

import (
	"testing"
	"time"
)

func TestVerificationRoundTrip(t *testing.T) {
	repo := NewBackupRepository(newTestDB(t))

	seedBackup(t, repo, "55555555-5555-5555-5555-555555555555", "/tmp/dump.sql")

	started := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	v := &BackupVerification{
		ID:              "ver-1",
		BackupID:        "55555555-5555-5555-5555-555555555555",
		Status:          VerificationRunning,
		Method:          MethodChecksum,
		ChecksumMatches: false,
		StartedAt:       started,
		CreatedAt:       started,
	}
	if err := repo.CreateVerification(v); err != nil {
		t.Fatalf("create verification: %v", err)
	}

	completed := started.Add(2 * time.Second)
	v.Status = VerificationPassed
	v.ChecksumMatches = true
	v.Details = "checksum matches manifest"
	v.CompletedAt = &completed
	if err := repo.UpdateVerification(v); err != nil {
		t.Fatalf("update verification: %v", err)
	}

	got, err := repo.GetLatestVerification(v.BackupID)
	if err != nil {
		t.Fatalf("get latest verification: %v", err)
	}

	if got.Status != VerificationPassed {
		t.Errorf("status = %q, want %q", got.Status, VerificationPassed)
	}
	if !got.ChecksumMatches {
		t.Error("expected checksum_matches to persist as true")
	}
	if got.Details != "checksum matches manifest" {
		t.Errorf("details = %q", got.Details)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to persist")
	}
}

func TestGetLatestVerificationReturnsMostRecent(t *testing.T) {
	repo := NewBackupRepository(newTestDB(t))

	backupID := "66666666-6666-6666-6666-666666666666"
	seedBackup(t, repo, backupID, "/tmp/dump.sql")

	earlier := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	later := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)

	for _, v := range []*BackupVerification{
		{ID: "old", BackupID: backupID, Status: VerificationFailed, Method: MethodChecksum, StartedAt: earlier, CreatedAt: earlier},
		{ID: "new", BackupID: backupID, Status: VerificationPassed, Method: MethodChecksum, StartedAt: later, CreatedAt: later},
	} {
		if err := repo.CreateVerification(v); err != nil {
			t.Fatalf("create verification %s: %v", v.ID, err)
		}
	}

	got, err := repo.GetLatestVerification(backupID)
	if err != nil {
		t.Fatalf("get latest verification: %v", err)
	}
	if got.ID != "new" {
		t.Fatalf("expected most recent verification, got %q", got.ID)
	}
}
