package backup

import "time"

// Verification status values.
const (
	// VerificationPending means the backup is queued for verification.
	VerificationPending = "pending"
	// VerificationRunning means a verifier is currently checking the backup.
	VerificationRunning = "running"
	// VerificationPassed means every check succeeded.
	VerificationPassed = "passed"
	// VerificationFailed means at least one check failed.
	VerificationFailed = "failed"
)

// Verification methods describe how thoroughly a backup was checked.
const (
	// MethodChecksum re-hashes the stored backup and compares it to the
	// manifest. It proves the bytes are intact but not that they restore.
	MethodChecksum = "checksum"
	// MethodRestore restores the backup and compares schema and row counts.
	// It is the strongest check and the reason this feature exists.
	MethodRestore = "restore"
)

// BackupVerification records a single verification run for a backup. Rows are
// kept rather than overwritten so the dashboard can show verification history
// and "verified at" age.
type BackupVerification struct {
	ID              string     `json:"id"`
	BackupID        string     `json:"backup_id"`
	Status          string     `json:"status"`
	Method          string     `json:"method"`
	ChecksumMatches bool       `json:"checksum_matches"`
	Details         string     `json:"details,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
