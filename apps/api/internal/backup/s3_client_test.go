package backup

import (
	"errors"
	"testing"

	"github.com/dendianugerah/velld/internal/settings"
)

func strPtr(s string) *string { return &s }

func TestValidateS3Settings(t *testing.T) {
	full := &settings.UserSettings{
		S3Enabled:   true,
		S3Endpoint:  strPtr("https://s3.example.com"),
		S3Bucket:    strPtr("backups"),
		S3AccessKey: strPtr("key"),
		S3SecretKey: strPtr("secret"),
	}

	t.Run("complete settings pass", func(t *testing.T) {
		if err := validateS3Settings(full); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing endpoint is rejected", func(t *testing.T) {
		broken := *full
		broken.S3Endpoint = nil
		err := validateS3Settings(&broken)
		if !errors.Is(err, ErrS3NotConfigured) {
			t.Fatalf("expected ErrS3NotConfigured, got %v", err)
		}
	})

	t.Run("empty secret is rejected", func(t *testing.T) {
		broken := *full
		broken.S3SecretKey = strPtr("")
		if err := validateS3Settings(&broken); !errors.Is(err, ErrS3NotConfigured) {
			t.Fatalf("expected ErrS3NotConfigured, got %v", err)
		}
	})
}

func TestS3RegionOrDefault(t *testing.T) {
	if got := s3RegionOrDefault(nil); got != "us-east-1" {
		t.Fatalf("region = %q, want us-east-1", got)
	}
	if got := s3RegionOrDefault(strPtr("")); got != "us-east-1" {
		t.Fatalf("region = %q, want us-east-1", got)
	}
	if got := s3RegionOrDefault(strPtr("eu-west-1")); got != "eu-west-1" {
		t.Fatalf("region = %q, want eu-west-1", got)
	}
}

func TestS3StringValue(t *testing.T) {
	if got := s3StringValue(nil); got != "" {
		t.Fatalf("value = %q, want empty", got)
	}
	if got := s3StringValue(strPtr("prefix")); got != "prefix" {
		t.Fatalf("value = %q, want prefix", got)
	}
}
