package settings

import (
	"database/sql"
	"time"

	"github.com/dendianugerah/velld/internal/common"
	"github.com/google/uuid"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetUserSettings(userID uuid.UUID) (*UserSettings, error) {
	settings := &UserSettings{}
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRow(`
        SELECT id, user_id, notify_dashboard, notify_email, notify_webhook,
               webhook_url, email, smtp_host, smtp_port, smtp_username, 
               smtp_password, s3_enabled, s3_endpoint, s3_region, s3_bucket,
               s3_access_key, s3_secret_key, s3_use_ssl, s3_path_prefix, s3_purge_local,
               created_at, updated_at
        FROM user_settings
        WHERE user_id = $1`, userID).Scan(
		&settings.ID, &settings.UserID, &settings.NotifyDashboard,
		&settings.NotifyEmail, &settings.NotifyWebhook, &settings.WebhookURL,
		&settings.Email, &settings.SMTPHost, &settings.SMTPPort,
		&settings.SMTPUsername, &settings.SMTPPassword,
		&settings.S3Enabled, &settings.S3Endpoint, &settings.S3Region, &settings.S3Bucket,
		&settings.S3AccessKey, &settings.S3SecretKey, &settings.S3UseSSL, &settings.S3PathPrefix,
		&settings.S3PurgeLocal,
		&createdAtStr, &updatedAtStr)

	if err == sql.ErrNoRows {
		// Create default settings if none exist
		now := time.Now()
		settings = &UserSettings{
			ID:              uuid.New(),
			UserID:          userID,
			NotifyDashboard: true,
			S3UseSSL:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		return settings, r.CreateUserSettings(settings)
	}

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	settings.CreatedAt, err = common.ParseTime(createdAtStr)
	if err != nil {
		return nil, err
	}

	settings.UpdatedAt, err = common.ParseTime(updatedAtStr)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (r *SettingsRepository) CreateUserSettings(settings *UserSettings) error {
	_, err := r.db.Exec(`
        INSERT INTO user_settings (
            id, user_id, notify_dashboard, notify_email, notify_webhook,
            webhook_url, email, smtp_host, smtp_port, smtp_username, 
            smtp_password, s3_enabled, s3_endpoint, s3_region, s3_bucket,
            s3_access_key, s3_secret_key, s3_use_ssl, s3_path_prefix, s3_purge_local,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`,
		settings.ID, settings.UserID, settings.NotifyDashboard,
		settings.NotifyEmail, settings.NotifyWebhook, settings.WebhookURL,
		settings.Email, settings.SMTPHost, settings.SMTPPort,
		settings.SMTPUsername, settings.SMTPPassword,
		settings.S3Enabled, settings.S3Endpoint, settings.S3Region, settings.S3Bucket,
		settings.S3AccessKey, settings.S3SecretKey, settings.S3UseSSL, settings.S3PathPrefix,
		settings.S3PurgeLocal,
		settings.CreatedAt, settings.UpdatedAt)
	return err
}

func (r *SettingsRepository) UpdateUserSettings(settings *UserSettings) error {
	settings.UpdatedAt = time.Now()
	_, err := r.db.Exec(`
        UPDATE user_settings SET
            notify_dashboard = $1, notify_email = $2, notify_webhook = $3,
            webhook_url = $4, email = $5, smtp_host = $6, smtp_port = $7,
            smtp_username = $8, smtp_password = $9, s3_enabled = $10,
            s3_endpoint = $11, s3_region = $12, s3_bucket = $13,
            s3_access_key = $14, s3_secret_key = $15, s3_use_ssl = $16,
            s3_path_prefix = $17, s3_purge_local = $18, updated_at = $19
        WHERE user_id = $20`,
		settings.NotifyDashboard, settings.NotifyEmail, settings.NotifyWebhook,
		settings.WebhookURL, settings.Email, settings.SMTPHost, settings.SMTPPort,
		settings.SMTPUsername, settings.SMTPPassword,
		settings.S3Enabled, settings.S3Endpoint, settings.S3Region, settings.S3Bucket,
		settings.S3AccessKey, settings.S3SecretKey, settings.S3UseSSL, settings.S3PathPrefix,
		settings.S3PurgeLocal,
		settings.UpdatedAt, settings.UserID)
	return err
}

// func (r *SettingsRepository) GetDatabaseBinaryPath(dbType string, userID uuid.UUID) (string, error) {
// 	var binPath sql.NullString
// 	err := r.db.QueryRow(`
//         SELECT CASE
//             WHEN $1 = 'postgresql' THEN postgresql_bin_path
//             WHEN $1 IN ('mysql', 'mariadb') THEN mysql_bin_path
//             WHEN $1 = 'mongodb' THEN mongodb_bin_path
//         END
//         FROM user_settings
//         WHERE user_id = $2`,
// 		dbType, userID).Scan(&binPath)

// 	if err == sql.ErrNoRows {
// 		return "", nil
// 	}

// 	if err != nil {
// 		return "", err
// 	}

// 	if !binPath.Valid {
// 		return "", nil
// 	}

// 	return binPath.String, nil
// }
