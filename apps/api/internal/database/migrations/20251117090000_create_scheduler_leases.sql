-- +goose Up
-- +goose StatementBegin
CREATE TABLE scheduler_leases (
    name TEXT PRIMARY KEY,
    holder TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE scheduler_leases;
-- +goose StatementEnd
