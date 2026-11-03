package backup

import (
	"context"
	"testing"
	"time"

	"github.com/dendianugerah/velld/internal/notification"
	"github.com/google/uuid"
)

// newWorkerTestService returns a service with a real migrated DB and a
// connection/backup seeded as described by seedBackupWithType.
func newWorkerTestService(t *testing.T) *BackupService {
	t.Helper()
	svc, _ := newVerificationService(t)
	return svc
}

func TestWorkerVerifiesQueuedBackup(t *testing.T) {
	svc := newWorkerTestService(t)
	repo := svc.backupRepo

	content := []byte("payload")
	path := writeBackupFile(t, content)
	seedBackupWithType(t, repo, "aaaaaaaa-0000-0000-0000-000000000001", path, "mysql")
	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "aaaaaaaa-0000-0000-0000-000000000001",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex(content),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	worker := newVerificationWorker(svc, 1, 0, time.Millisecond)
	defer worker.Stop(context.Background())

	if !worker.Enqueue("aaaaaaaa-0000-0000-0000-000000000001") {
		t.Fatal("expected enqueue to succeed")
	}

	waitForVerification(t, repo, "aaaaaaaa-0000-0000-0000-000000000001")

	latest, err := repo.GetLatestVerification("aaaaaaaa-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("get latest verification: %v", err)
	}
	if latest.Status != VerificationPassed {
		t.Fatalf("status = %q, want %q", latest.Status, VerificationPassed)
	}
}

func TestWorkerRecordsFailedVerification(t *testing.T) {
	svc := newWorkerTestService(t)
	repo := svc.backupRepo

	path := writeBackupFile(t, []byte("tampered"))
	seedBackupWithType(t, repo, "aaaaaaaa-0000-0000-0000-000000000002", path, "mysql")
	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "aaaaaaaa-0000-0000-0000-000000000002",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex([]byte("original")),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	worker := newVerificationWorker(svc, 1, 0, time.Millisecond)
	defer worker.Stop(context.Background())

	worker.Enqueue("aaaaaaaa-0000-0000-0000-000000000002")
	waitForVerification(t, repo, "aaaaaaaa-0000-0000-0000-000000000002")

	latest, err := repo.GetLatestVerification("aaaaaaaa-0000-0000-0000-000000000002")
	if err != nil {
		t.Fatalf("get latest verification: %v", err)
	}
	if latest.Status != VerificationFailed {
		t.Fatalf("status = %q, want %q", latest.Status, VerificationFailed)
	}
}

func TestWorkerStopDrainsQueue(t *testing.T) {
	svc := newWorkerTestService(t)
	repo := svc.backupRepo

	content := []byte("payload")
	path := writeBackupFile(t, content)
	seedBackupWithType(t, repo, "aaaaaaaa-0000-0000-0000-000000000003", path, "mysql")
	if err := repo.CreateBackupManifest(&BackupManifest{
		BackupID:          "aaaaaaaa-0000-0000-0000-000000000003",
		ChecksumAlgorithm: defaultChecksumAlgorithm,
		Checksum:          sha256Hex(content),
	}); err != nil {
		t.Fatalf("create manifest: %v", err)
	}

	worker := newVerificationWorker(svc, 1, 0, time.Millisecond)
	worker.Enqueue("aaaaaaaa-0000-0000-0000-000000000003")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := worker.Stop(ctx); err != nil {
		t.Fatalf("stop: %v", err)
	}

	if _, err := repo.GetLatestVerification("aaaaaaaa-0000-0000-0000-000000000003"); err != nil {
		t.Fatalf("expected job to be processed before stop returned: %v", err)
	}
}

func waitForVerification(t *testing.T, repo *BackupRepository, backupID string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		v, err := repo.GetLatestVerification(backupID)
		if err == nil && v.Status != VerificationRunning {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("verification did not complete for %s", backupID)
}

func TestVerificationFailureCreatesNotification(t *testing.T) {
	svc := newWorkerTestService(t)
	repo := svc.backupRepo

	backupID := "aaaaaaaa-0000-0000-0000-000000000004"
	path := writeBackupFile(t, []byte("payload"))
	seedBackupWithType(t, repo, backupID, path, "mysql")

	if err := svc.createVerificationFailureNotification(backupID, errTestVerification); err != nil {
		t.Fatalf("create notification: %v", err)
	}

	notes, err := svc.notificationRepo.GetUserNotifications(uuid.MustParse("99999999-9999-9999-9999-999999999999"))
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notes))
	}
	if notes[0].Type != notification.VerificationFailed {
		t.Fatalf("type = %q, want %q", notes[0].Type, notification.VerificationFailed)
	}
}

type verificationError struct{}

func (verificationError) Error() string { return "checksum mismatch" }

var errTestVerification = verificationError{}
