-- +goose Up
-- +goose StatementBegin
SELECT 'Adding SSH tunnel support to connections';

ALTER TABLE connections ADD COLUMN ssh_enabled INTEGER DEFAULT 0;
ALTER TABLE connections ADD COLUMN ssh_host TEXT;
ALTER TABLE connections ADD COLUMN ssh_port INTEGER DEFAULT 22;
ALTER TABLE connections ADD COLUMN ssh_username TEXT;
ALTER TABLE connections ADD COLUMN ssh_password TEXT;
ALTER TABLE connections ADD COLUMN ssh_private_key TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing SSH tunnel support from connections';

ALTER TABLE connections DROP COLUMN ssh_enabled;
ALTER TABLE connections DROP COLUMN ssh_host;
ALTER TABLE connections DROP COLUMN ssh_port;
ALTER TABLE connections DROP COLUMN ssh_username;
ALTER TABLE connections DROP COLUMN ssh_password;
ALTER TABLE connections DROP COLUMN ssh_private_key;

-- +goose StatementEnd
