package backup

import "testing"

func TestGenerateDiff(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		target      string
		wantAdded   int
		wantRemoved int
	}{
		{
			name:   "identical input has no changes",
			source: "a\nb\nc",
			target: "a\nb\nc",
		},
		{
			name:      "appended lines are added",
			source:    "a\nb",
			target:    "a\nb\nc",
			wantAdded: 1,
		},
		{
			name:        "trailing lines removed",
			source:      "a\nb\nc",
			target:      "a\nb",
			wantRemoved: 1,
		},
		{
			name:      "single line modified becomes remove plus add",
			source:    "keep1\nold\nkeep2",
			target:    "keep1\nnew\nkeep2",
			wantAdded: 1,

			wantRemoved: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDiff(tt.source, tt.target)
			if got.Added != tt.wantAdded {
				t.Errorf("Added = %d, want %d", got.Added, tt.wantAdded)
			}
			if got.Removed != tt.wantRemoved {
				t.Errorf("Removed = %d, want %d", got.Removed, tt.wantRemoved)
			}
			if got.Modified != 0 {
				t.Errorf("Modified = %d, want 0 (modifications are reported as remove plus add)", got.Modified)
			}
		})
	}
}

func TestGenerateDiffCountsEveryLine(t *testing.T) {
	got := generateDiff("a\nb\nc", "a\nb\nc")

	if total := got.Added + got.Removed + got.Unchanged; total != 3 {
		t.Fatalf("expected all 3 lines accounted for, got %d", total)
	}
}

func TestGenerateDiffCapsChanges(t *testing.T) {
	source := ""
	target := ""
	for i := 0; i < 5000; i++ {
		source += "line\n"
		target += "changed\n"
	}

	got := generateDiff(source, target)
	if len(got.Changes) > 1001 {
		t.Fatalf("expected changes to be capped near 1000, got %d", len(got.Changes))
	}
}
