package common

import (
	"runtime"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "RFC3339",
			input: "2025-01-02T15:04:05Z",
			want:  time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			name:  "SQLite with positive offset",
			input: "2025-01-02 15:04:05+07:00",
			want:  time.Date(2025, 1, 2, 15, 4, 5, 0, time.FixedZone("", 7*60*60)),
		},
		{
			name:  "SQLite without timezone",
			input: "2025-01-02 15:04:05",
			want:  time.Date(2025, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			name:    "invalid",
			input:   "not a timestamp",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeConnectionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple", input: "My Database", want: "my_database"},
		{name: "already clean", input: "prod_db", want: "prod_db"},
		{name: "special characters", input: "Prod / DB (v2)!", want: "prod_db_v2"},
		{name: "leading and trailing separators", input: "  --db--  ", want: "db"},
		{name: "empty falls back to backup", input: "", want: "backup"},
		{name: "only separators falls back to backup", input: "---", want: "backup"},
		{name: "uppercase is lowercased", input: "PRODUCTION", want: "production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeConnectionName(tt.input); got != tt.want {
				t.Fatalf("SanitizeConnectionName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeConnectionNameTruncatesLongInput(t *testing.T) {
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'a'
	}

	got := SanitizeConnectionName(string(long))
	if len(got) > 200 {
		t.Fatalf("expected length <= 200, got %d", len(got))
	}
}

func TestGetPlatformExecutableName(t *testing.T) {
	got := GetPlatformExecutableName("pg_dump")
	want := "pg_dump"
	if runtime.GOOS == "windows" {
		want = "pg_dump.exe"
	}
	if got != want {
		t.Fatalf("GetPlatformExecutableName = %q, want %q", got, want)
	}
}
