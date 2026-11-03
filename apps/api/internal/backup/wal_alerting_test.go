package backup

import (
	"testing"

	"github.com/dendianugerah/velld/internal/notification"
	"github.com/google/uuid"
)

func TestWALGapNotificationCreated(t *testing.T) {
	svc, repo := newVerificationService(t)
	seedBackupWithType(t, repo, "dddddddd-0000-0000-0000-000000000001", "/tmp/x.sql", "postgresql")

	conn, err := svc.connStorage.GetConnection("conn-1")
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}

	gaps := []string{"000000010000000000000004", "000000010000000000000005"}
	if err := svc.createWALGapNotification(conn, gaps); err != nil {
		t.Fatalf("create notification: %v", err)
	}

	notes, err := svc.notificationRepo.GetUserNotifications(uuid.MustParse("99999999-9999-9999-9999-999999999999"))
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notes))
	}
	if notes[0].Type != notification.WALGapDetected {
		t.Fatalf("type = %q, want %q", notes[0].Type, notification.WALGapDetected)
	}
}

func TestFirstN(t *testing.T) {
	values := []string{"a", "b", "c", "d"}
	if got := firstN(values, 2); len(got) != 2 || got[1] != "b" {
		t.Fatalf("firstN = %v", got)
	}
	if got := firstN(values, 10); len(got) != 4 {
		t.Fatalf("firstN with n > len should return all, got %v", got)
	}
}
