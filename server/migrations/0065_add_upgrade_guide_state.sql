-- +goose Up
CREATE TABLE IF NOT EXISTS upgrade_guide_state (
    task_id VARCHAR(128) PRIMARY KEY,
    revision INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(16) NOT NULL CHECK (status IN ('completed', 'dismissed')),
    selection JSONB NOT NULL DEFAULT '{}'::jsonb,
    decided_by BIGINT NULL,
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DELETE FROM sys_config WHERE config_key LIKE 'system.upgrade_guide.%.completed';

-- +goose Down
DROP TABLE IF EXISTS upgrade_guide_state;
