-- +goose Up
-- +goose StatementBegin
SELECT 'Adding s3_cleanup_on_retention to connections table';

ALTER TABLE connections ADD COLUMN s3_cleanup_on_retention INTEGER DEFAULT 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing s3_cleanup_on_retention from connections table';

ALTER TABLE connections DROP COLUMN s3_cleanup_on_retention;

-- +goose StatementEnd
