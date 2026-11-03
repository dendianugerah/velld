-- +goose Up
-- +goose StatementBegin
CREATE TABLE backup_verifications (
    id TEXT PRIMARY KEY,
    backup_id TEXT NOT NULL REFERENCES backups(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    method TEXT NOT NULL,
    checksum_matches INTEGER NOT NULL DEFAULT 0,
    details TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

CREATE INDEX idx_backup_verifications_backup_id ON backup_verifications(backup_id);
CREATE INDEX idx_backup_verifications_status ON backup_verifications(status);

-- +goose Down
-- +goose StatementBegin
DROP TABLE backup_verifications;
-- +goose StatementEnd
