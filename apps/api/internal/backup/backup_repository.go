package backup

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dendianugerah/velld/internal/common"
)

type BackupRepository struct {
	db *sql.DB
}

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{
		db: db,
	}
}

func (r *BackupRepository) CreateBackupSchedule(schedule *BackupSchedule) error {
	var nextRunStr *string
	if schedule.NextRunTime != nil {
		str := schedule.NextRunTime.Format(time.RFC3339)
		nextRunStr = &str
	}

	var lastBackupStr *string
	if schedule.LastBackupTime != nil {
		str := schedule.LastBackupTime.Format(time.RFC3339)
		lastBackupStr = &str
	}

	now := time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		INSERT INTO backup_schedules (
			id, connection_id, enabled, cron_schedule, retention_days, retention_count,
			next_run_time, last_backup_time, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		schedule.ID, schedule.ConnectionID, schedule.Enabled,
		schedule.CronSchedule, schedule.RetentionDays, schedule.RetentionCount,
		nextRunStr, lastBackupStr, now, now)
	return err
}

func (r *BackupRepository) UpdateBackupSchedule(schedule *BackupSchedule) error {
	var nextRunStr *string
	if schedule.NextRunTime != nil {
		str := schedule.NextRunTime.Format(time.RFC3339)
		nextRunStr = &str
	}

	var lastBackupStr *string
	if schedule.LastBackupTime != nil {
		str := schedule.LastBackupTime.Format(time.RFC3339)
		lastBackupStr = &str
	}

	query := `
		UPDATE backup_schedules 
		SET enabled = $1, 
		    cron_schedule = $2, 
		    retention_days = $3, 
		    retention_count = $4,
		    next_run_time = $5,
		    last_backup_time = $6,
		    updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(query,
		schedule.Enabled,
		schedule.CronSchedule,
		schedule.RetentionDays,
		schedule.RetentionCount,
		nextRunStr,
		lastBackupStr,
		time.Now(),
		schedule.ID)
	if err != nil {
		return fmt.Errorf("failed to update backup schedule: %v", err)
	}

	return nil
}

func (r *BackupRepository) recordScheduleRun(id string, nextRun, lastRun time.Time) error {
	_, err := r.db.Exec(`
		UPDATE backup_schedules SET next_run_time = $1, last_backup_time = $2, updated_at = $2
		WHERE id = $3`, nextRun.Format(time.RFC3339), lastRun.Format(time.RFC3339), id)
	return err
}

func (r *BackupRepository) GetBackupSchedule(connectionID string) (*BackupSchedule, error) {
	var (
		nextRunStr    sql.NullString
		lastBackupStr sql.NullString
		createdAtStr  string
		updatedAtStr  string
	)
	schedule := &BackupSchedule{}
	err := r.db.QueryRow(`
		SELECT id, connection_id, enabled, cron_schedule, COALESCE(retention_days, 0), retention_count,
		       next_run_time, last_backup_time, created_at, updated_at 
		FROM backup_schedules 
		WHERE connection_id = $1
		ORDER BY created_at DESC LIMIT 1`,
		connectionID).Scan(
		&schedule.ID, &schedule.ConnectionID, &schedule.Enabled,
		&schedule.CronSchedule, &schedule.RetentionDays, &schedule.RetentionCount,
		&nextRunStr, &lastBackupStr, &createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse next_run_time if not null
	if nextRunStr.Valid {
		nextRun, err := common.ParseTime(nextRunStr.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing next_run_time: %v", err)
		}
		schedule.NextRunTime = &nextRun
	}

	// Parse last_backup_time if not null
	if lastBackupStr.Valid {
		lastBackup, err := common.ParseTime(lastBackupStr.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing last_backup_time: %v", err)
		}
		schedule.LastBackupTime = &lastBackup
	}

	// Parse created_at and updated_at
	createdAt, err := common.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %v", err)
	}
	schedule.CreatedAt = createdAt

	updatedAt, err := common.ParseTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %v", err)
	}
	schedule.UpdatedAt = updatedAt

	return schedule, nil
}

func (r *BackupRepository) GetAllActiveSchedules() ([]*BackupSchedule, error) {
	rows, err := r.db.Query(`
		SELECT id, connection_id, enabled, cron_schedule, COALESCE(retention_days, 0), retention_count,
		       next_run_time, last_backup_time, created_at, updated_at 
		FROM backup_schedules 
		WHERE enabled = true
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*BackupSchedule
	for rows.Next() {
		var (
			nextRunStr    sql.NullString
			lastBackupStr sql.NullString
			createdAtStr  string
			updatedAtStr  string
		)
		schedule := &BackupSchedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.ConnectionID, &schedule.Enabled,
			&schedule.CronSchedule, &schedule.RetentionDays, &schedule.RetentionCount,
			&nextRunStr, &lastBackupStr, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// Parse next_run_time if not null
		if nextRunStr.Valid {
			nextRun, err := common.ParseTime(nextRunStr.String)
			if err != nil {
				return nil, fmt.Errorf("error parsing next_run_time: %v", err)
			}
			schedule.NextRunTime = &nextRun
		}

		// Parse last_backup_time if not null
		if lastBackupStr.Valid {
			lastBackup, err := common.ParseTime(lastBackupStr.String)
			if err != nil {
				return nil, fmt.Errorf("error parsing last_backup_time: %v", err)
			}
			schedule.LastBackupTime = &lastBackup
		}

		// Parse created_at and updated_at
		createdAt, err := common.ParseTime(createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing created_at: %v", err)
		}
		schedule.CreatedAt = createdAt

		updatedAt, err := common.ParseTime(updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated_at: %v", err)
		}
		schedule.UpdatedAt = updatedAt

		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

// Backup Methods

func (r *BackupRepository) CreateBackup(backup *Backup) error {
	_, err := r.db.Exec(`
		INSERT INTO backups (
			id, connection_id, schedule_id, status, path, s3_object_key, size,
			started_time, completed_time, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		backup.ID, backup.ConnectionID, backup.ScheduleID,
		backup.Status, backup.Path, backup.S3ObjectKey, backup.Size,
		backup.StartedTime, backup.CompletedTime,
		backup.CreatedAt, backup.UpdatedAt)
	return err
}

func (r *BackupRepository) UpdateBackupStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE backups SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *BackupRepository) getBackupsOutsideRetention(connectionID string, cutoffTime *time.Time, count int) ([]*Backup, error) {
	rows, err := r.db.Query(`
		WITH ranked AS (
			SELECT id, path, s3_object_key, julianday(created_at) AS created,
			       ROW_NUMBER() OVER (ORDER BY julianday(created_at) DESC, created_at DESC, id DESC) AS backup_rank
			FROM backups
			WHERE connection_id = $1 AND status = 'completed'
		)
		SELECT id, path, s3_object_key FROM ranked
		WHERE created < julianday($2) OR ($3 > 0 AND backup_rank > $3)`,
		connectionID, cutoffTime, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*Backup
	for rows.Next() {
		backup := &Backup{}
		err := rows.Scan(&backup.ID, &backup.Path, &backup.S3ObjectKey)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}
	return backups, rows.Err()
}

func (r *BackupRepository) backupArtifactsAreShared(backup *Backup) (local, remote bool, err error) {
	err = r.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM backups WHERE id <> $1 AND path = $2),
		       EXISTS(SELECT 1 FROM backups WHERE id <> $1 AND s3_object_key = $3)`,
		backup.ID, backup.Path, backup.S3ObjectKey).Scan(&local, &remote)
	return
}

func (r *BackupRepository) DeleteBackup(id string) error {
	_, err := r.db.Exec("DELETE FROM backups WHERE id = $1", id)
	return err
}

func (r *BackupRepository) GetBackup(id string) (*Backup, error) {
	var (
		startedTimeStr   string
		completedTimeStr sql.NullString
		createdAtStr     string
		updatedAtStr     string
	)
	backup := &Backup{}
	err := r.db.QueryRow(`
		SELECT id, connection_id, schedule_id, status, path, s3_object_key, size,
			   started_time, completed_time, created_at, updated_at 
		FROM backups WHERE id = $1`, id).
		Scan(&backup.ID, &backup.ConnectionID, &backup.ScheduleID,
			&backup.Status, &backup.Path, &backup.S3ObjectKey, &backup.Size,
			&startedTimeStr, &completedTimeStr,
			&createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, err
	}

	// Parse started_time
	startedTime, err := common.ParseTime(startedTimeStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing started_time: %v", err)
	}
	backup.StartedTime = startedTime

	// Parse completed_time if not null
	if completedTimeStr.Valid {
		completedTime, err := common.ParseTime(completedTimeStr.String)
		if err != nil {
			return nil, fmt.Errorf("error parsing completed_time: %v", err)
		}
		backup.CompletedTime = &completedTime
	}

	// Parse created_at and updated_at
	createdAt, err := common.ParseTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing created_at: %v", err)
	}
	backup.CreatedAt = createdAt

	updatedAt, err := common.ParseTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing updated_at: %v", err)
	}
	backup.UpdatedAt = updatedAt

	return backup, nil
}

func (r *BackupRepository) GetAllBackupsWithPagination(opts BackupListOptions) ([]*BackupList, int, error) {
	whereClause := "WHERE c.user_id = $1"
	args := []interface{}{opts.UserID}
	argCount := 2

	if opts.Search != "" {
		whereClause += fmt.Sprintf(" AND (LOWER(b.path) LIKE $%d OR LOWER(b.status) LIKE $%d)", argCount, argCount)
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		argCount++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM backups b
		INNER JOIN connections c ON b.connection_id = c.id
		%s`, whereClause)

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT 
			b.id, b.connection_id, c.type, b.schedule_id, b.status, b.path, b.s3_object_key, b.size,
			b.started_time, b.completed_time, b.created_at, b.updated_at,
			c.database_name
		FROM backups b
		INNER JOIN connections c ON b.connection_id = c.id
		%s
		ORDER BY b.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, opts.Limit, opts.Offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	backups := make([]*BackupList, 0)
	for rows.Next() {
		var (
			startedTimeStr   sql.NullString
			completedTimeStr sql.NullString
			createdAtStr     string
			updatedAtStr     string
		)
		backup := &BackupList{}
		err := rows.Scan(
			&backup.ID, &backup.ConnectionID, &backup.DatabaseType,
			&backup.ScheduleID, &backup.Status, &backup.Path, &backup.S3ObjectKey, &backup.Size,
			&startedTimeStr, &completedTimeStr,
			&createdAtStr, &updatedAtStr,
			&backup.DatabaseName,
		)
		if err != nil {
			return nil, 0, err
		}

		backup.StartedTime = startedTimeStr.String
		backup.CompletedTime = completedTimeStr.String
		backup.CreatedAt = createdAtStr
		backup.UpdatedAt = updatedAtStr

		backups = append(backups, backup)
	}

	return backups, total, rows.Err()
}

func (r *BackupRepository) UpdateBackupStatusAndSchedule(id string, status string, scheduleID string) error {
	_, err := r.db.Exec(`
		UPDATE backups 
		SET status = $1, schedule_id = $2, updated_at = $3 
		WHERE id = $4`,
		status, scheduleID, time.Now().Format(time.RFC3339), id)
	return err
}

func (r *BackupRepository) GetBackupsByConnectionID(connectionID string) ([]*Backup, error) {
	rows, err := r.db.Query(`
		SELECT id, connection_id, schedule_id, status, path, s3_object_key, size,
		       started_time, completed_time, created_at, updated_at
		FROM backups
		WHERE connection_id = $1
		ORDER BY created_at DESC`,
		connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []*Backup
	for rows.Next() {
		backup := &Backup{}
		var startedTimeStr, createdAtStr, updatedAtStr string
		var completedTimeStr sql.NullString

		err := rows.Scan(
			&backup.ID, &backup.ConnectionID, &backup.ScheduleID, &backup.Status,
			&backup.Path, &backup.S3ObjectKey, &backup.Size,
			&startedTimeStr, &completedTimeStr, &createdAtStr, &updatedAtStr,
		)
		if err != nil {
			return nil, err
		}

		backup.StartedTime, err = common.ParseTime(startedTimeStr)
		if err != nil {
			return nil, err
		}

		if completedTimeStr.Valid {
			completedTime, err := common.ParseTime(completedTimeStr.String)
			if err != nil {
				return nil, err
			}
			backup.CompletedTime = &completedTime
		}

		backup.CreatedAt, err = common.ParseTime(createdAtStr)
		if err != nil {
			return nil, err
		}

		backup.UpdatedAt, err = common.ParseTime(updatedAtStr)
		if err != nil {
			return nil, err
		}

		backups = append(backups, backup)
	}

	return backups, rows.Err()
}

func (r *BackupRepository) UpdateBackupS3ObjectKey(backupID string, s3ObjectKey string) error {
	_, err := r.db.Exec(`
		UPDATE backups 
		SET s3_object_key = $1, updated_at = datetime('now') 
		WHERE id = $2`,
		s3ObjectKey, backupID)
	return err
}
