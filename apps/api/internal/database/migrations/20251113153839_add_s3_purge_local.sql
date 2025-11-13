-- +goose Up
-- +goose StatementBegin
SELECT 'Adding s3_purge_local setting to user_settings';

ALTER TABLE user_settings ADD COLUMN s3_purge_local INTEGER DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing s3_purge_local setting from user_settings';

ALTER TABLE user_settings DROP COLUMN s3_purge_local;

-- +goose StatementEnd
