-- +goose Up
-- +goose StatementBegin
SELECT 'Adding S3 storage settings to user_settings';

ALTER TABLE user_settings ADD COLUMN s3_enabled INTEGER DEFAULT 0;
ALTER TABLE user_settings ADD COLUMN s3_endpoint TEXT;
ALTER TABLE user_settings ADD COLUMN s3_region TEXT;
ALTER TABLE user_settings ADD COLUMN s3_bucket TEXT;
ALTER TABLE user_settings ADD COLUMN s3_access_key TEXT;
ALTER TABLE user_settings ADD COLUMN s3_secret_key TEXT;
ALTER TABLE user_settings ADD COLUMN s3_use_ssl INTEGER DEFAULT 1;
ALTER TABLE user_settings ADD COLUMN s3_path_prefix TEXT;

-- Add s3_object_key to backups table to track S3 storage location
ALTER TABLE backups ADD COLUMN s3_object_key TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing S3 storage settings from user_settings';

ALTER TABLE user_settings DROP COLUMN s3_enabled;
ALTER TABLE user_settings DROP COLUMN s3_endpoint;
ALTER TABLE user_settings DROP COLUMN s3_region;
ALTER TABLE user_settings DROP COLUMN s3_bucket;
ALTER TABLE user_settings DROP COLUMN s3_access_key;
ALTER TABLE user_settings DROP COLUMN s3_secret_key;
ALTER TABLE user_settings DROP COLUMN s3_use_ssl;
ALTER TABLE user_settings DROP COLUMN s3_path_prefix;

ALTER TABLE backups DROP COLUMN s3_object_key;

-- +goose StatementEnd
