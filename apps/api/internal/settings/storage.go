package settings

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dendianugerah/velld/internal/objectstorage"
	"github.com/google/uuid"
)

var (
	errInvalidStorageSettings = errors.New("invalid storage settings")
	r2AccountIDPattern        = regexp.MustCompile("^[a-fA-F0-9]{32}$")
)

func (settings *UserSettings) storageConfig() (objectstorage.S3Config, error) {
	value := func(field *string) string {
		if field == nil {
			return ""
		}
		return *field
	}
	config := objectstorage.S3Config{
		Endpoint: value(settings.S3Endpoint), Region: value(settings.S3Region),
		Bucket: value(settings.S3Bucket), AccessKey: value(settings.S3AccessKey),
		UseSSL: settings.S3UseSSL, PathPrefix: value(settings.S3PathPrefix),
	}
	switch settings.S3Provider {
	case "", "s3":
		if config.Endpoint == "" {
			return config, fmt.Errorf("%w: enter an S3 endpoint", errInvalidStorageSettings)
		}
		if config.Region == "" {
			config.Region = "us-east-1"
		}
	case "r2":
		accountID := value(settings.R2AccountID)
		if !r2AccountIDPattern.MatchString(accountID) {
			return config, fmt.Errorf("%w: enter a 32-character hexadecimal Cloudflare account ID", errInvalidStorageSettings)
		}
		config.IsR2 = true
		config.Endpoint = strings.ToLower(accountID) + ".r2.cloudflarestorage.com"
		config.Region = "auto"
		config.UseSSL = true
	default:
		return config, fmt.Errorf("%w: choose S3 or Cloudflare R2", errInvalidStorageSettings)
	}
	if config.Bucket == "" || config.AccessKey == "" || value(settings.S3SecretKey) == "" {
		return config, fmt.Errorf("%w: enter a bucket name, access key ID, and secret access key", errInvalidStorageSettings)
	}
	return config, nil
}

func (s *SettingsService) OpenStorage(ctx context.Context, settings *UserSettings) (*objectstorage.S3Storage, error) {
	return s.openStorage(ctx, settings, true)
}

func (s *SettingsService) OpenConfiguredStorage(ctx context.Context, settings *UserSettings) (*objectstorage.S3Storage, error) {
	if !settings.S3Enabled {
		return nil, nil
	}
	storage, err := s.OpenStorage(ctx, settings)
	if errors.Is(err, errInvalidStorageSettings) {
		return nil, nil
	}
	return storage, err
}

func (s *SettingsService) openStorage(ctx context.Context, settings *UserSettings, createBucket bool) (*objectstorage.S3Storage, error) {
	config, err := settings.storageConfig()
	if err != nil {
		return nil, err
	}
	config.SecretKey, err = s.cryptoService.Decrypt(*settings.S3SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt storage secret key")
	}
	return objectstorage.NewS3Storage(ctx, config, createBucket)
}

func (s *SettingsService) TestStorage(ctx context.Context, userID uuid.UUID, req *UpdateSettingsRequest) error {
	settings, err := s.GetUserSettingsInternal(userID)
	if err != nil {
		return err
	}
	if err := s.applyStorageSettings(settings, req); err != nil {
		return err
	}
	storage, err := s.openStorage(ctx, settings, false)
	if err != nil {
		return err
	}
	return storage.TestConnection(ctx)
}
