package settings

import (
	"github.com/dendianugerah/velld/internal/common"
	"github.com/google/uuid"
)

type SettingsService struct {
	repo          *SettingsRepository
	cryptoService *common.EncryptionService
}

func NewSettingsService(repo *SettingsRepository, crypto *common.EncryptionService) *SettingsService {
	return &SettingsService{
		repo:          repo,
		cryptoService: crypto,
	}
}

func (s *SettingsService) GetUserSettings(userID uuid.UUID) (*UserSettings, error) {
	settings, err := s.repo.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data before returning
	settings.SMTPPassword = nil
	settings.S3SecretKey = nil
	return settings, nil
}

func (s *SettingsService) UpdateUserSettings(userID uuid.UUID, req *UpdateSettingsRequest) (*UserSettings, error) {
	settings, err := s.repo.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.NotifyDashboard != nil {
		settings.NotifyDashboard = *req.NotifyDashboard
	}
	if req.NotifyEmail != nil {
		settings.NotifyEmail = *req.NotifyEmail
	}
	if req.NotifyWebhook != nil {
		settings.NotifyWebhook = *req.NotifyWebhook
	}
	if req.WebhookURL != nil {
		settings.WebhookURL = req.WebhookURL
	}
	if req.Email != nil {
		settings.Email = req.Email
	}
	if req.SMTPHost != nil {
		settings.SMTPHost = req.SMTPHost
	}
	if req.SMTPPort != nil {
		settings.SMTPPort = req.SMTPPort
	}
	if req.SMTPUsername != nil {
		settings.SMTPUsername = req.SMTPUsername
	}
	if req.SMTPPassword != nil {
		// Encrypt SMTP password before storing
		encryptedPass, err := s.cryptoService.Encrypt(*req.SMTPPassword)
		if err != nil {
			return nil, err
		}
		settings.SMTPPassword = &encryptedPass
	}

	// Update S3 settings
	if req.S3Enabled != nil {
		settings.S3Enabled = *req.S3Enabled
	}
	if req.S3Endpoint != nil {
		settings.S3Endpoint = req.S3Endpoint
	}
	if req.S3Region != nil {
		settings.S3Region = req.S3Region
	}
	if req.S3Bucket != nil {
		settings.S3Bucket = req.S3Bucket
	}
	if req.S3AccessKey != nil {
		settings.S3AccessKey = req.S3AccessKey
	}
	if req.S3SecretKey != nil {
		// Encrypt S3 secret key before storing
		encryptedKey, err := s.cryptoService.Encrypt(*req.S3SecretKey)
		if err != nil {
			return nil, err
		}
		settings.S3SecretKey = &encryptedKey
	}
	if req.S3UseSSL != nil {
		settings.S3UseSSL = *req.S3UseSSL
	}
	if req.S3PathPrefix != nil {
		settings.S3PathPrefix = req.S3PathPrefix
	}

	if err := s.repo.UpdateUserSettings(settings); err != nil {
		return nil, err
	}

	// Remove sensitive data before returning
	settings.SMTPPassword = nil
	settings.S3SecretKey = nil
	return settings, nil
}
