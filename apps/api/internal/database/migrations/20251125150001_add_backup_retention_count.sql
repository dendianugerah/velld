-- +goose Up
ALTER TABLE backup_schedules ADD COLUMN retention_count INTEGER NOT NULL DEFAULT 0 CHECK (retention_count >= 0);

-- +goose Down
ALTER TABLE backup_schedules DROP COLUMN retention_count;
