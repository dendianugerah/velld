package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func (s *BackupService) ScheduleBackup(req *ScheduleBackupRequest) error {
	// Check if a schedule already exists for this connection
	existingSchedule, err := s.backupRepo.GetBackupSchedule(req.ConnectionID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing schedule: %v", err)
	}

	schedule, err := parseCronSchedule(req.CronSchedule)
	if err != nil {
		return err
	}

	nextRun := schedule.Next(time.Now())

	if existingSchedule != nil {
		// Update existing schedule
		existingSchedule.Enabled = true
		existingSchedule.CronSchedule = req.CronSchedule
		existingSchedule.RetentionDays = req.RetentionDays
		existingSchedule.NextRunTime = &nextRun
		existingSchedule.UpdatedAt = time.Now()

		if err := s.backupRepo.UpdateBackupSchedule(existingSchedule); err != nil {
			return fmt.Errorf("failed to update backup schedule: %v", err)
		}

		// Update cron job
		scheduleID := existingSchedule.ID.String()
		if oldEntryID, exists := s.cronEntries[scheduleID]; exists {
			s.cronManager.Remove(oldEntryID)
		}

		entryID, err := s.cronManager.AddFunc(req.CronSchedule, func() {
			s.executeCronBackup(existingSchedule)
		})
		if err != nil {
			return fmt.Errorf("failed to schedule backup: %v", err)
		}

		s.cronEntries[scheduleID] = entryID
		return nil
	}

	// Create new schedule if none exists
	backupSchedule := &BackupSchedule{
		ID:            uuid.New(),
		ConnectionID:  req.ConnectionID,
		Enabled:       true,
		CronSchedule:  req.CronSchedule,
		RetentionDays: req.RetentionDays,
		NextRunTime:   &nextRun,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.backupRepo.CreateBackupSchedule(backupSchedule); err != nil {
		return fmt.Errorf("failed to save backup schedule: %v", err)
	}

	scheduleID := backupSchedule.ID.String()
	entryID, err := s.cronManager.AddFunc(req.CronSchedule, func() {
		s.executeCronBackup(backupSchedule)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule backup: %v", err)
	}

	s.cronEntries[scheduleID] = entryID
	return nil
}

func (s *BackupService) executeCronBackup(schedule *BackupSchedule) {
	// if schedule.CronSchedule == "0 */1 * * * *" {
	// 	err := fmt.Errorf("test failure: this is a simulated backup failure for SMTP testing")
	// 	if notifyErr := s.createFailureNotification(schedule.ConnectionID, err); notifyErr != nil {
	// 		fmt.Printf("Error creating failure notification: %v\n", notifyErr)
	// 	}
	// 	return
	// }

	backup, err := s.CreateBackup(context.Background(), schedule.ConnectionID)
	if err != nil {
		if notifyErr := s.createFailureNotification(schedule.ConnectionID, err); notifyErr != nil {
			fmt.Printf("Error creating failure notification: %v\n", notifyErr)
		}
	} else {
		scheduleIDStr := schedule.ID.String()
		if err := s.backupRepo.UpdateBackupStatusAndSchedule(backup.ID.String(), backup.Status, scheduleIDStr); err != nil {
			fmt.Printf("Error updating backup status and schedule: %v\n", err)
		}
	}

	// Update schedule's next run time and last backup time
	nextRun, _ := nextRunTime(schedule.CronSchedule, time.Now())
	schedule.NextRunTime = &nextRun
	now := time.Now()
	schedule.LastBackupTime = &now
	schedule.UpdatedAt = now

	if err := s.backupRepo.UpdateBackupSchedule(schedule); err != nil {
		fmt.Printf("Error updating backup schedule: %v\n", err)
	}

	if schedule.RetentionDays > 0 {
		s.cleanupOldBackups(schedule.ConnectionID, schedule.RetentionDays)
	}
}

func (s *BackupService) cleanupOldBackups(connectionID string, retentionDays int) {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	oldBackups, err := s.backupRepo.GetBackupsOlderThan(connectionID, cutoffTime)
	if err != nil {
		fmt.Printf("Error fetching old backups for cleanup: %v\n", err)
		return
	}

	if len(oldBackups) == 0 {
		return
	}

	// Get connection to retrieve user settings for S3
	conn, err := s.connStorage.GetConnection(connectionID)
	if err != nil {
		fmt.Printf("Error getting connection for cleanup: %v\n", err)
		return
	}

	// Initialize S3 client if S3 is enabled and configured
	var s3Storage *S3Storage
	if storage, err := s.s3StorageForUser(conn.UserID); err == nil {
		s3Storage = storage
	} else if !errors.Is(err, ErrS3NotConfigured) {
		fmt.Printf("Warning: Failed to set up S3 client for cleanup: %v\n", err)
	}

	// Clean up old backups
	ctx := context.Background()
	for _, backup := range oldBackups {
		backupID := backup.ID.String()

		// Delete from S3 if object key exists, S3 is configured, and connection has S3 cleanup enabled
		if backup.S3ObjectKey != nil && *backup.S3ObjectKey != "" && s3Storage != nil && conn.S3CleanupOnRetention {
			if err := s3Storage.DeleteFile(ctx, *backup.S3ObjectKey); err != nil {
				fmt.Printf("Warning: Failed to delete S3 object %s for backup %s: %v\n",
					*backup.S3ObjectKey, backupID, err)
			} else {
				fmt.Printf("Deleted S3 object %s for backup %s (retention cleanup)\n",
					*backup.S3ObjectKey, backupID)
			}
		}

		// Delete local file if it exists
		if _, err := os.Stat(backup.Path); err == nil {
			if err := os.Remove(backup.Path); err != nil {
				fmt.Printf("Warning: Failed to delete local file %s for backup %s: %v\n",
					backup.Path, backupID, err)
			} else {
				fmt.Printf("Deleted local file %s for backup %s (retention cleanup)\n",
					backup.Path, backupID)
			}
		}

		// Delete backup record from database
		if err := s.backupRepo.DeleteBackup(backupID); err != nil {
			fmt.Printf("Error deleting backup record %s: %v\n", backupID, err)
		} else {
			fmt.Printf("Deleted backup record %s (retention cleanup)\n", backupID)
		}
	}

	fmt.Printf("Retention cleanup completed: processed %d old backups for connection %s\n",
		len(oldBackups), connectionID)
}

func (s *BackupService) DisableBackupSchedule(connectionID string) error {
	schedule, err := s.backupRepo.GetBackupSchedule(connectionID)
	if err != nil {
		return err
	}

	scheduleID := schedule.ID.String()
	if entryID, exists := s.cronEntries[scheduleID]; exists {
		s.cronManager.Remove(entryID)
		delete(s.cronEntries, scheduleID)
	}

	schedule.Enabled = false
	schedule.UpdatedAt = time.Now()
	if err := s.backupRepo.UpdateBackupSchedule(schedule); err != nil {
		return err
	}

	return nil
}

func (s *BackupService) UpdateBackupSchedule(connectionID string, req *UpdateScheduleRequest) error {
	schedule, err := s.backupRepo.GetBackupSchedule(connectionID)
	if err != nil {
		return err
	}

	if _, err := parseCronSchedule(req.CronSchedule); err != nil {
		return err
	}

	schedule.CronSchedule = req.CronSchedule
	schedule.RetentionDays = req.RetentionDays
	err = s.backupRepo.UpdateBackupSchedule(schedule)
	if err != nil {
		return err
	}

	// Remove old cron job
	if entryID, ok := s.cronEntries[schedule.ID.String()]; ok {
		s.cronManager.Remove(entryID)
		delete(s.cronEntries, schedule.ID.String())
	}

	// Add new cron job
	entryID, err := s.cronManager.AddFunc(schedule.CronSchedule, func() {
		s.executeCronBackup(schedule)
	})
	if err != nil {
		return fmt.Errorf("failed to register cron job: %v", err)
	}

	// Store the new entry ID
	s.cronEntries[schedule.ID.String()] = entryID

	return nil
}
