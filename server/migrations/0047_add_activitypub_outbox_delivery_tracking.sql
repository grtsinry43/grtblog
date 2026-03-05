-- +goose Up
ALTER TABLE activitypub_outbox_item
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'queued',
    ADD COLUMN IF NOT EXISTS trigger_source VARCHAR(20) NOT NULL DEFAULT 'auto',
    ADD COLUMN IF NOT EXISTS total_targets INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS success_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS failure_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS deliveries JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS duration_ms BIGINT;

UPDATE activitypub_outbox_item
SET status = 'completed'
WHERE status = 'queued';

CREATE INDEX IF NOT EXISTS idx_activitypub_outbox_status_published
    ON activitypub_outbox_item (status, published_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_activitypub_outbox_status_published;

ALTER TABLE activitypub_outbox_item
    DROP COLUMN IF EXISTS duration_ms,
    DROP COLUMN IF EXISTS finished_at,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS deliveries,
    DROP COLUMN IF EXISTS failure_count,
    DROP COLUMN IF EXISTS success_count,
    DROP COLUMN IF EXISTS total_targets,
    DROP COLUMN IF EXISTS trigger_source,
    DROP COLUMN IF EXISTS status;
