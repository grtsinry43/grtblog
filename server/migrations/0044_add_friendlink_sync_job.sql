-- +goose Up
CREATE TABLE IF NOT EXISTS friend_link_sync_job (
    id             BIGSERIAL PRIMARY KEY,
    target_type    VARCHAR(30)  NOT NULL,
    sync_method    VARCHAR(20)  NOT NULL DEFAULT 'rss',
    friend_link_id BIGINT,
    instance_id    BIGINT,
    target_url     VARCHAR(500) NOT NULL,
    feed_url       VARCHAR(500),
    status         VARCHAR(20)  NOT NULL DEFAULT 'queued',
    attempt_count  INT          NOT NULL DEFAULT 0,
    max_attempts   INT          NOT NULL DEFAULT 1,
    next_retry_at  TIMESTAMPTZ,
    started_at     TIMESTAMPTZ,
    finished_at    TIMESTAMPTZ,
    duration_ms    BIGINT,
    pulled_count   INT          NOT NULL DEFAULT 0,
    error_message  TEXT,
    trigger_source VARCHAR(40)  NOT NULL DEFAULT 'scheduler',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_friend_link_sync_job_target_type CHECK (target_type IN ('friend_link','federation_instance')),
    CONSTRAINT chk_friend_link_sync_job_method CHECK (sync_method IN ('timeline','rss')),
    CONSTRAINT chk_friend_link_sync_job_status CHECK (status IN ('queued','running','success','failed'))
);

CREATE INDEX IF NOT EXISTS idx_friend_link_sync_job_status_created
    ON friend_link_sync_job (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_friend_link_sync_job_next_retry
    ON friend_link_sync_job (next_retry_at);
CREATE INDEX IF NOT EXISTS idx_friend_link_sync_job_friend_link_id
    ON friend_link_sync_job (friend_link_id);
CREATE INDEX IF NOT EXISTS idx_friend_link_sync_job_instance_id
    ON friend_link_sync_job (instance_id);

ALTER TABLE friend_link_sync_job
    ADD CONSTRAINT fk_friend_link_sync_job_friend_link
        FOREIGN KEY (friend_link_id) REFERENCES friend_link (id);

ALTER TABLE friend_link_sync_job
    ADD CONSTRAINT fk_friend_link_sync_job_instance
        FOREIGN KEY (instance_id) REFERENCES federation_instance (id);

-- +goose Down
ALTER TABLE friend_link_sync_job
    DROP CONSTRAINT IF EXISTS fk_friend_link_sync_job_instance;

ALTER TABLE friend_link_sync_job
    DROP CONSTRAINT IF EXISTS fk_friend_link_sync_job_friend_link;

DROP TABLE IF EXISTS friend_link_sync_job;
