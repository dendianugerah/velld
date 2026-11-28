-- +goose Up
ALTER TABLE connections ADD COLUMN d1_account_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connections ADD COLUMN d1_database_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connections ADD COLUMN d1_api_token TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE connections DROP COLUMN d1_api_token;
ALTER TABLE connections DROP COLUMN d1_database_id;
ALTER TABLE connections DROP COLUMN d1_account_id;
