-- +goose Up
CREATE SCHEMA IF NOT EXISTS backup_ops;

CREATE TABLE IF NOT EXISTS backup_ops.backup_record
(
    id                VARCHAR(36) PRIMARY KEY,
    filename          VARCHAR(255) NOT NULL,
    status            VARCHAR(24) NOT NULL,
    stage             VARCHAR(48) NOT NULL DEFAULT 'queued',
    trigger_type      VARCHAR(24) NOT NULL DEFAULT 'manual',
    size_bytes        BIGINT NOT NULL DEFAULT 0,
    sha256            VARCHAR(64),
    app_version       VARCHAR(64),
    migration_version BIGINT NOT NULL DEFAULT 0,
    db_server_version VARCHAR(128),
    site_name         TEXT,
    site_url          TEXT,
    upload_file_count BIGINT NOT NULL DEFAULT 0,
    error_message     TEXT,
    pinned            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_backup_record_created_at
    ON backup_ops.backup_record (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_record_status
    ON backup_ops.backup_record (status);

CREATE TABLE IF NOT EXISTS backup_ops.download_ticket
(
    token_hash VARCHAR(64) PRIMARY KEY,
    backup_id  VARCHAR(36) NOT NULL REFERENCES backup_ops.backup_record (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_download_ticket_expires_at
    ON backup_ops.download_ticket (expires_at);

-- +goose Down
DROP SCHEMA IF EXISTS backup_ops CASCADE;
