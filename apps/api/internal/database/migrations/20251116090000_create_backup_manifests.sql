-- +goose Up
-- +goose StatementBegin
CREATE TABLE backup_manifests (
    backup_id TEXT PRIMARY KEY REFERENCES backups(id) ON DELETE CASCADE,
    checksum_algorithm TEXT NOT NULL,
    checksum TEXT NOT NULL,
    compression TEXT,
    uncompressed_size INTEGER NOT NULL DEFAULT 0,
    compressed_size INTEGER NOT NULL DEFAULT 0,
    server_version TEXT,
    tool_version TEXT,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

CREATE INDEX idx_backup_manifests_checksum ON backup_manifests(checksum);

-- +goose Down
-- +goose StatementBegin
DROP TABLE backup_manifests;
-- +goose StatementEnd
