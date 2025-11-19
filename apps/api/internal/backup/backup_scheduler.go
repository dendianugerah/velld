package backup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

func (s *BackupService) ScheduleBackup(req *ScheduleBackupRequest) error {
	// Check if a schedule already exists for this connection
	existingSchedule, err := s.backupRepo.GetBackupSchedule(req.ConnectionID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing schedule: %v", err)
	}

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.CronSchedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule: %v", err)
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

	backup, err := s.CreateBackup(schedule.ConnectionID)
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
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, _ := parser.Parse(schedule.CronSchedule)
	nextRun := cronSchedule.Next(time.Now())
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

	// Get user settings to check if S3 is enabled
	userSettings, err := s.settingsService.GetUserSettingsInternal(conn.UserID)
	if err != nil {
		fmt.Printf("Warning: Failed to get user settings for cleanup: %v\n", err)
		// Continue with local cleanup even if we can't get S3 settings
	}

	// Initialize S3 client if S3 is enabled and configured
	var s3Storage *S3Storage
	if userSettings != nil && userSettings.S3Enabled && 
	   userSettings.S3Endpoint != nil && *userSettings.S3Endpoint != "" &&
	   userSettings.S3Bucket != nil && *userSettings.S3Bucket != "" &&
	   userSettings.S3AccessKey != nil && *userSettings.S3AccessKey != "" &&
	   userSettings.S3SecretKey != nil && *userSettings.S3SecretKey != "" {
		
		secretKey, err := s.cryptoService.Decrypt(*userSettings.S3SecretKey)
		if err != nil {
			fmt.Printf("Warning: Failed to decrypt S3 secret key for cleanup: %v\n", err)
		} else {
			region := "us-east-1"
			if userSettings.S3Region != nil && *userSettings.S3Region != "" {
				region = *userSettings.S3Region
			}

			pathPrefix := ""
			if userSettings.S3PathPrefix != nil {
				pathPrefix = *userSettings.S3PathPrefix
			}

			s3Config := S3Config{
				Endpoint:   *userSettings.S3Endpoint,
				Region:     region,
				Bucket:     *userSettings.S3Bucket,
				AccessKey:  *userSettings.S3AccessKey,
				SecretKey:  secretKey,
				UseSSL:     userSettings.S3UseSSL,
				PathPrefix: pathPrefix,
			}

			s3Storage, err = NewS3Storage(s3Config)
			if err != nil {
				fmt.Printf("Warning: Failed to create S3 storage client for cleanup: %v\n", err)
				s3Storage = nil
			}
		}
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

	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err = parser.Parse(req.CronSchedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule: %v", err)
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
