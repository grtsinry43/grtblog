-- +goose Up
CREATE TABLE IF NOT EXISTS backup_ops.schedule_config
(
    id              SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    interval_hours  INTEGER NOT NULL DEFAULT 24 CHECK (interval_hours BETWEEN 1 AND 8760),
    retention_count INTEGER NOT NULL DEFAULT 7 CHECK (retention_count BETWEEN 1 AND 100),
    next_run_at     TIMESTAMPTZ,
    last_run_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO backup_ops.schedule_config (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS backup_ops.schedule_config;
