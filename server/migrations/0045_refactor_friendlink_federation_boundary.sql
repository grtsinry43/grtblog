-- +goose Up
-- friend_link: switch from kind/sync_mode to type.
ALTER TABLE friend_link
    ADD COLUMN IF NOT EXISTS type VARCHAR(20);

UPDATE friend_link
SET type = CASE
    WHEN kind = 'federation' THEN 'federation'
    WHEN sync_mode = 'rss' THEN 'rss'
    ELSE 'norss'
END
WHERE type IS NULL OR btrim(type) = '';

ALTER TABLE friend_link
    ALTER COLUMN type SET NOT NULL;

ALTER TABLE friend_link
    DROP COLUMN IF EXISTS kind,
    DROP COLUMN IF EXISTS sync_mode;

ALTER TABLE friend_link
    DROP CONSTRAINT IF EXISTS chk_friend_link_type;

ALTER TABLE friend_link
    ADD CONSTRAINT chk_friend_link_type
        CHECK (type IN ('federation', 'rss', 'norss'));

ALTER TABLE friend_link
    DROP CONSTRAINT IF EXISTS chk_friend_link_type_instance;

ALTER TABLE friend_link
    ADD CONSTRAINT chk_friend_link_type_instance
        CHECK (
            (type = 'federation' AND instance_id IS NOT NULL)
            OR (type IN ('rss', 'norss') AND instance_id IS NULL)
        );

-- friend_link_sync_job: no longer uses federation_instance as a target.
ALTER TABLE friend_link_sync_job
    DROP CONSTRAINT IF EXISTS chk_friend_link_sync_job_target_type;

ALTER TABLE friend_link_sync_job
    ADD CONSTRAINT chk_friend_link_sync_job_target_type
        CHECK (target_type IN ('friend_link'));

-- federated_post_cache: move ownership to friend_link.
TRUNCATE TABLE federated_post_cache;

ALTER TABLE federated_post_cache
    ADD COLUMN IF NOT EXISTS friend_link_id BIGINT;

ALTER TABLE federated_post_cache
    ADD COLUMN IF NOT EXISTS source_method VARCHAR(20) NOT NULL DEFAULT 'rss';

ALTER TABLE federated_post_cache
    ALTER COLUMN friend_link_id SET NOT NULL;

ALTER TABLE federated_post_cache
    ALTER COLUMN instance_id DROP NOT NULL;

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS fk_federated_post_friend_link;

ALTER TABLE federated_post_cache
    ADD CONSTRAINT fk_federated_post_friend_link
        FOREIGN KEY (friend_link_id) REFERENCES friend_link (id);

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS uq_federated_post_url;

ALTER TABLE federated_post_cache
    ADD CONSTRAINT uq_federated_post_friend_link_url UNIQUE (friend_link_id, url);

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS chk_federated_post_source_method;

ALTER TABLE federated_post_cache
    ADD CONSTRAINT chk_federated_post_source_method
        CHECK (source_method IN ('timeline', 'rss', 'rss_fallback'));

CREATE INDEX IF NOT EXISTS idx_federated_post_friend_link_published
    ON federated_post_cache (friend_link_id, published_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_federated_post_friend_link_published;

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS chk_federated_post_source_method;

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS uq_federated_post_friend_link_url;

ALTER TABLE federated_post_cache
    ADD CONSTRAINT uq_federated_post_url UNIQUE (url);

ALTER TABLE federated_post_cache
    DROP CONSTRAINT IF EXISTS fk_federated_post_friend_link;

ALTER TABLE federated_post_cache
    ALTER COLUMN instance_id SET NOT NULL;

ALTER TABLE federated_post_cache
    DROP COLUMN IF EXISTS source_method,
    DROP COLUMN IF EXISTS friend_link_id;

ALTER TABLE friend_link_sync_job
    DROP CONSTRAINT IF EXISTS chk_friend_link_sync_job_target_type;

ALTER TABLE friend_link_sync_job
    ADD CONSTRAINT chk_friend_link_sync_job_target_type
        CHECK (target_type IN ('friend_link', 'federation_instance'));

ALTER TABLE friend_link
    DROP CONSTRAINT IF EXISTS chk_friend_link_type_instance;

ALTER TABLE friend_link
    DROP CONSTRAINT IF EXISTS chk_friend_link_type;

ALTER TABLE friend_link
    ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS sync_mode VARCHAR(20) NOT NULL DEFAULT 'none';

UPDATE friend_link
SET kind = CASE
        WHEN type = 'federation' THEN 'federation'
        ELSE 'manual'
    END,
    sync_mode = CASE
        WHEN type = 'federation' THEN 'federation'
        WHEN type = 'rss' THEN 'rss'
        ELSE 'none'
    END;

ALTER TABLE friend_link
    DROP COLUMN IF EXISTS type;
