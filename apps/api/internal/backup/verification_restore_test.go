package backup

import (
	"strings"
	"testing"
)

func TestCompareTableCounts(t *testing.T) {
	source := map[string]int64{"users": 10, "orders": 3, "only_source": 1}

	t.Run("all match", func(t *testing.T) {
		restored := map[string]int64{"users": 10, "orders": 3, "only_source": 1}
		if m := compareTableCounts(source, restored); len(m) != 0 {
			t.Fatalf("expected no mismatches, got %v", m)
		}
	})

	t.Run("row drift is reported", func(t *testing.T) {
		restored := map[string]int64{"users": 9, "orders": 3, "only_source": 1}
		m := compareTableCounts(source, restored)
		if len(m) != 1 || !strings.Contains(m[0], "users") {
			t.Fatalf("expected a users mismatch, got %v", m)
		}
	})

	t.Run("missing and extra tables are reported", func(t *testing.T) {
		restored := map[string]int64{"users": 10, "orders": 3, "only_restored": 1}
		m := compareTableCounts(source, restored)

		joined := strings.Join(m, "\n")
		if !strings.Contains(joined, "only_source") {
			t.Errorf("expected missing source table to be reported: %v", m)
		}
		if !strings.Contains(joined, "only_restored") {
			t.Errorf("expected extra restored table to be reported: %v", m)
		}
	})
}

func TestScratchDatabaseName(t *testing.T) {
	got := scratchDatabaseName("12345678-90ab-cdef-1234-567890abcdef")
	if !strings.HasPrefix(got, "velld_verify_") {
		t.Fatalf("expected velld_verify_ prefix, got %q", got)
	}
	if strings.Contains(got, "-") {
		t.Fatalf("scratch name must be a valid identifier, got %q", got)
	}
	if len(got) > len("velld_verify_")+12 {
		t.Fatalf("scratch name too long: %q", got)
	}
}

func TestQuotePgIdent(t *testing.T) {
	tests := map[string]string{
		"users":        `"users"`,
		`weird"name`:   `"weird""name"`,
		"velld_verify": `"velld_verify"`,
	}
	for input, want := range tests {
		if got := quotePgIdent(input); got != want {
			t.Fatalf("quotePgIdent(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestEscapeMySQLIdent(t *testing.T) {
	tests := map[string]string{
		"users":        "users",
		"weird`name":   "weird``name",
		"velld_verify": "velld_verify",
	}
	for input, want := range tests {
		if got := escapeMySQLIdent(input); got != want {
			t.Fatalf("escapeMySQLIdent(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestQuoteMySQLString(t *testing.T) {
	if got := quoteMySQLString("app"); got != "'app'" {
		t.Fatalf("got %q, want 'app'", got)
	}
	if got := quoteMySQLString("o'brien"); got != `'o\'brien'` {
		t.Fatalf("got %q, want escaped quote", got)
	}
}
