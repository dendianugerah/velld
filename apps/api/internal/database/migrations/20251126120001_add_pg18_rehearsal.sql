-- +goose Up
ALTER TABLE backups ADD COLUMN pg18_status TEXT;
ALTER TABLE backups ADD COLUMN pg18_report TEXT;

-- +goose Down
ALTER TABLE backups DROP COLUMN pg18_report;
ALTER TABLE backups DROP COLUMN pg18_status;
