package backup

import (
	"testing"
	"time"
)

func TestParseCronSchedule(t *testing.T) {
	valid := []string{
		"0 0 * * * *",
		"*/30 * * * * *",
		"0 30 2 * * *",
	}
	for _, expr := range valid {
		t.Run(expr, func(t *testing.T) {
			if _, err := parseCronSchedule(expr); err != nil {
				t.Fatalf("expected %q to be valid: %v", expr, err)
			}
		})
	}

	invalid := []string{
		"",
		"not cron",
		"* * * * *",
		"0 0 0 0 0",
	}
	for _, expr := range invalid {
		t.Run("invalid_"+expr, func(t *testing.T) {
			if _, err := parseCronSchedule(expr); err == nil {
				t.Fatalf("expected %q to be rejected", expr)
			}
		})
	}
}

func TestNextRunTime(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	next, err := nextRunTime("0 0 * * * *", from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := from.Add(time.Hour)
	if !next.Equal(want) {
		t.Fatalf("next run = %v, want %v", next, want)
	}
}
