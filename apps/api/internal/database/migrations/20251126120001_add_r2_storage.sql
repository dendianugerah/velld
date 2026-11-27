-- +goose Up
ALTER TABLE user_settings ADD COLUMN s3_provider TEXT NOT NULL DEFAULT 's3';
ALTER TABLE user_settings ADD COLUMN r2_account_id TEXT;

-- +goose Down
ALTER TABLE user_settings DROP COLUMN r2_account_id;
ALTER TABLE user_settings DROP COLUMN s3_provider;
