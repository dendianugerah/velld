package backup

import (
	"time"

	"github.com/google/uuid"
)

// BackupSchedule represents a backup schedule configuration
type BackupSchedule struct {
	ID             uuid.UUID  `json:"id"`
	ConnectionID   string     `json:"connection_id"`
	Enabled        bool       `json:"enabled"`
	CronSchedule   string     `json:"cron_schedule"`
	RetentionDays  int        `json:"retention_days"`
	NextRunTime    *time.Time `json:"next_run_time"`
	LastBackupTime *time.Time `json:"last_backup_time"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Backup represents a single backup record
type Backup struct {
	ID            uuid.UUID  `json:"id"`
	ConnectionID  string     `json:"connection_id"`
	ScheduleID    *string    `json:"schedule_id"`
	Status        string     `json:"status"`
	Path          string     `json:"path"`
	S3ObjectKey   *string    `json:"s3_object_key"`
	Size          int64      `json:"size"`
	StartedTime   time.Time  `json:"started_time"`
	CompletedTime *time.Time `json:"completed_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BackupList represents a backup in list view with additional info
type BackupList struct {
	ID            uuid.UUID `json:"id"`
	ConnectionID  string    `json:"connection_id"`
	DatabaseType  string    `json:"database_type"`
	DatabaseName  string    `json:"database_name"`
	ScheduleID    *string   `json:"schedule_id"`
	Status        string    `json:"status"`
	Path          string    `json:"path"`
	S3ObjectKey   *string   `json:"s3_object_key"`
	Size          int64     `json:"size"`
	StartedTime   string    `json:"started_time"`
	CompletedTime string    `json:"completed_time"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

// BackupRequest represents a request to create a backup
type BackupRequest struct {
	ConnectionID string `json:"connection_id"`
}

// ScheduleBackupRequest represents a request to create a backup schedule
type ScheduleBackupRequest struct {
	ConnectionID  string `json:"connection_id"`
	CronSchedule  string `json:"cron_schedule"`
	RetentionDays int    `json:"retention_days"`
}

// BackupStats represents backup statistics
type BackupStats struct {
	TotalBackups    int     `json:"total_backups"`
	FailedBackups   int     `json:"failed_backups"`
	TotalSize       int64   `json:"total_size"`
	AverageDuration float64 `json:"average_duration"`
	SuccessRate     float64 `json:"success_rate"`
}

// PaginatedBackupResponse represents a paginated response of backups
type PaginatedBackupResponse struct {
	Data       []*BackupList `json:"data"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// BackupListRequest represents a request to list backups with pagination
type BackupListRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search"`
}

// BackupListOptions represents options for listing backups
type BackupListOptions struct {
	UserID uuid.UUID
	Limit  int
	Offset int
	Search string
}

type UpdateScheduleRequest struct {
	CronSchedule  string `json:"cron_schedule"`
	RetentionDays int    `json:"retention_days"`
}
