-- +goose Up
ALTER TABLE backups ADD COLUMN verification_status TEXT;
ALTER TABLE backups ADD COLUMN verification_message TEXT;
ALTER TABLE backups ADD COLUMN verified_at TEXT;

-- +goose Down
ALTER TABLE backups DROP COLUMN verified_at;
ALTER TABLE backups DROP COLUMN verification_message;
ALTER TABLE backups DROP COLUMN verification_status;
