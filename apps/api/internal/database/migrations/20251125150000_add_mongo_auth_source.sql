-- +goose Up
ALTER TABLE connections ADD COLUMN mongo_auth_source TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE connections DROP COLUMN mongo_auth_source;
