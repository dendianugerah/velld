package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildPITRConfig(t *testing.T) {
	target := time.Date(2026, 4, 25, 14, 32, 7, 0, time.UTC)
	cfg := buildPITRConfig(target, "cp /mnt/wal/%f %p")

	if !cfg.RecoverySignal {
		t.Error("expected recovery.signal to be enabled")
	}
	if !strings.Contains(cfg.PostgresqlAutoConf, "restore_command = 'cp /mnt/wal/%f %p'") {
		t.Errorf("missing restore_command: %s", cfg.PostgresqlAutoConf)
	}
	if !strings.Contains(cfg.PostgresqlAutoConf, "recovery_target_time = '2026-04-25 14:32:07.000000+00'") {
		t.Errorf("missing recovery_target_time: %s", cfg.PostgresqlAutoConf)
	}
	if !strings.Contains(cfg.PostgresqlAutoConf, "recovery_target_action = 'promote'") {
		t.Errorf("missing recovery_target_action: %s", cfg.PostgresqlAutoConf)
	}
}

func TestBuildPITRConfigEscapesQuotes(t *testing.T) {
	cfg := buildPITRConfig(time.Now().Add(-time.Hour), "echo 'hi'")
	if !strings.Contains(cfg.PostgresqlAutoConf, "restore_command = 'echo ''hi'''") {
		t.Fatalf("quotes not escaped: %s", cfg.PostgresqlAutoConf)
	}
}

func TestValidatePITRTarget(t *testing.T) {
	now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	t.Run("accepts a past RFC3339 time", func(t *testing.T) {
		got, err := validatePITRTarget("2026-04-25T14:32:07Z", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Equal(time.Date(2026, 4, 25, 14, 32, 7, 0, time.UTC)) {
			t.Fatalf("parsed %v", got)
		}
	})

	t.Run("rejects a future time", func(t *testing.T) {
		if _, err := validatePITRTarget("2026-06-01T00:00:00Z", now); err == nil {
			t.Fatal("expected future target to be rejected")
		}
	})

	t.Run("rejects empty", func(t *testing.T) {
		if _, err := validatePITRTarget("  ", now); err == nil {
			t.Fatal("expected empty target to be rejected")
		}
	})

	t.Run("rejects garbage", func(t *testing.T) {
		if _, err := validatePITRTarget("not-a-time", now); err == nil {
			t.Fatal("expected garbage to be rejected")
		}
	})
}

func TestPrepareRecoveryDir(t *testing.T) {
	svc := &BackupService{}
	dir := t.TempDir()
	target := time.Date(2026, 4, 25, 14, 32, 7, 0, time.UTC)

	result, err := svc.PrepareRecoveryDir(dir, target, "cp /mnt/wal/%f %p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TargetTime != "2026-04-25T14:32:07Z" {
		t.Fatalf("target time = %q", result.TargetTime)
	}

	if _, err := os.Stat(filepath.Join(dir, "recovery.signal")); err != nil {
		t.Fatalf("recovery.signal missing: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(dir, "postgresql.auto.conf"))
	if err != nil {
		t.Fatalf("read auto.conf: %v", err)
	}
	if !strings.Contains(string(content), "recovery_target_time") {
		t.Fatalf("auto.conf missing recovery_target_time: %s", content)
	}
}
