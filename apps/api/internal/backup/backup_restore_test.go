package backup

import (
	"errors"
	"strings"
	"testing"
)

func TestIsCriticalPostgreSQLError(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "permission warning is not critical", line: "WARNING: must be member of role", want: false},
		{name: "privileges notice is not critical", line: "ERROR: no privileges to grant", want: false},
		{name: "notice is not critical", line: "NOTICE: relation already exists", want: false},
		{name: "syntax error is critical", line: "ERROR: syntax error at or near \"foo\"", want: true},
		{name: "already exists is critical", line: "ERROR: relation \"users\" already exists", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCriticalPostgreSQLError(tt.line); got != tt.want {
				t.Fatalf("isCriticalPostgreSQLError(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestValidatePostgreSQLRestore(t *testing.T) {
	s := &BackupService{}

	t.Run("clean output succeeds", func(t *testing.T) {
		if err := s.validatePostgreSQLRestore([]byte("SET\nCREATE TABLE\n"), nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-critical errors are ignored", func(t *testing.T) {
		output := []byte("ERROR: no privileges to grant\nERROR: must be member of role\n")
		if err := s.validatePostgreSQLRestore(output, nil); err != nil {
			t.Fatalf("non-critical errors should be ignored, got: %v", err)
		}
	})

	t.Run("critical errors fail the restore", func(t *testing.T) {
		err := s.validatePostgreSQLRestore([]byte("ERROR: syntax error at end of input\n"), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "syntax error") {
			t.Fatalf("expected error to mention the cause, got: %v", err)
		}
	})

	t.Run("existing objects produce an actionable message", func(t *testing.T) {
		err := s.validatePostgreSQLRestore([]byte("ERROR: relation \"users\" already exists\n"), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "must be empty") {
			t.Fatalf("expected guidance about empty target, got: %v", err)
		}
	})

	t.Run("command error with clean output is reported", func(t *testing.T) {
		cmdErr := errors.New("exit status 1")
		err := s.validatePostgreSQLRestore(nil, cmdErr)
		if err == nil || !strings.Contains(err.Error(), "exit status 1") {
			t.Fatalf("expected wrapped command error, got: %v", err)
		}
	})
}

func TestValidateMySQLError(t *testing.T) {
	s := &BackupService{}

	if err := s.validateMySQLRestore("app", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := s.validateMySQLRestore("app", []byte("access denied"), errors.New("exit status 1"))
	if err == nil || !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected output in error, got: %v", err)
	}

	err = s.validateMongoDBRestore("app", nil, errors.New("exit status 1"))
	if err == nil || !strings.Contains(err.Error(), "exit status 1") {
		t.Fatalf("expected fallback to command error, got: %v", err)
	}
}
