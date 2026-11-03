package backup

import (
	"errors"
	"fmt"

	"github.com/dendianugerah/velld/internal/settings"
	"github.com/google/uuid"
)

// ErrS3NotConfigured is returned when the user has not enabled S3 or has not
// filled in every required field. Callers treat it as "skip S3" rather than a
// failure.
var ErrS3NotConfigured = errors.New("s3 storage is not configured")

// s3StorageForUser validates the user's S3 settings and returns a ready client.
// It replaces the validation and client construction that was previously copied
// into every S3 operation.
func (s *BackupService) s3StorageForUser(userID uuid.UUID) (*S3Storage, error) {
	userSettings, err := s.settingsService.GetUserSettingsInternal(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	if !userSettings.S3Enabled {
		return nil, ErrS3NotConfigured
	}

	if err := validateS3Settings(userSettings); err != nil {
		return nil, err
	}

	secretKey, err := s.cryptoService.Decrypt(*userSettings.S3SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt S3 secret key: %w", err)
	}

	return NewS3Storage(S3Config{
		Endpoint:   *userSettings.S3Endpoint,
		Region:     s3RegionOrDefault(userSettings.S3Region),
		Bucket:     *userSettings.S3Bucket,
		AccessKey:  *userSettings.S3AccessKey,
		SecretKey:  secretKey,
		UseSSL:     userSettings.S3UseSSL,
		PathPrefix: s3StringValue(userSettings.S3PathPrefix),
	})
}

func validateS3Settings(userSettings *settings.UserSettings) error {
	required := map[string]*string{
		"endpoint":   userSettings.S3Endpoint,
		"bucket":     userSettings.S3Bucket,
		"access key": userSettings.S3AccessKey,
		"secret key": userSettings.S3SecretKey,
	}

	for name, value := range required {
		if value == nil || *value == "" {
			return fmt.Errorf("%w: %s not set", ErrS3NotConfigured, name)
		}
	}

	return nil
}

func s3RegionOrDefault(region *string) string {
	if region != nil && *region != "" {
		return *region
	}
	return "us-east-1"
}

func s3StringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
