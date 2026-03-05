-- +goose Up
ALTER TABLE moment
    ADD COLUMN IF NOT EXISTS activitypub_object_id VARCHAR(500);

ALTER TABLE moment
    ADD COLUMN IF NOT EXISTS activitypub_last_published_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uq_moment_activitypub_object_id
    ON moment (activitypub_object_id)
    WHERE activitypub_object_id IS NOT NULL;

ALTER TABLE thinking
    ADD COLUMN IF NOT EXISTS activitypub_object_id VARCHAR(500);

ALTER TABLE thinking
    ADD COLUMN IF NOT EXISTS activitypub_last_published_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uq_thinking_activitypub_object_id
    ON thinking (activitypub_object_id)
    WHERE activitypub_object_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_thinking_activitypub_object_id;

ALTER TABLE thinking
    DROP COLUMN IF EXISTS activitypub_last_published_at;

ALTER TABLE thinking
    DROP COLUMN IF EXISTS activitypub_object_id;

DROP INDEX IF EXISTS uq_moment_activitypub_object_id;

ALTER TABLE moment
    DROP COLUMN IF EXISTS activitypub_last_published_at;

ALTER TABLE moment
    DROP COLUMN IF EXISTS activitypub_object_id;
