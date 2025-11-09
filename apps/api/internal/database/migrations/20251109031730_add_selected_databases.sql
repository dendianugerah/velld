-- +goose Up
-- +goose StatementBegin

ALTER TABLE connections ADD COLUMN selected_databases TEXT DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connections DROP COLUMN selected_databases;
-- +goose StatementEnd
