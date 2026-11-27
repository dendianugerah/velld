package settings

import (
	"errors"
	"testing"
)

func TestStorageConfigProviders(t *testing.T) {
	endpoint, bucket, accessKey, encryptedKey := "s3.example.com", "backups", "key", "encrypted"
	settings := UserSettings{
		S3Endpoint: &endpoint, S3Bucket: &bucket, S3AccessKey: &accessKey, S3SecretKey: &encryptedKey,
	}
	s3, err := settings.storageConfig()
	if err != nil || s3.IsR2 || s3.Endpoint != endpoint || s3.Region != "us-east-1" || s3.UseSSL {
		t.Fatalf("legacy S3 config: %+v, %v", s3, err)
	}

	accountID := "0123456789ABCDEF0123456789ABCDEF"
	settings.S3Provider, settings.R2AccountID = "r2", &accountID
	r2, err := settings.storageConfig()
	if err != nil || !r2.IsR2 || r2.Endpoint != "0123456789abcdef0123456789abcdef.r2.cloudflarestorage.com" ||
		r2.Region != "auto" || !r2.UseSSL || r2.Bucket != bucket || r2.AccessKey != accessKey {
		t.Fatalf("R2 config: %+v, %v", r2, err)
	}

	accountID = "example.com:443"
	if _, err := settings.storageConfig(); !errors.Is(err, errInvalidStorageSettings) {
		t.Fatalf("invalid account ID error: %v", err)
	}
}
