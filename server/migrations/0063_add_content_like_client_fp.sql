-- +goose Up
ALTER TABLE content_like
    ADD COLUMN IF NOT EXISTS client_fp VARCHAR(64) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_content_like_client_fp
    ON content_like (target_type, target_id, client_fp)
    WHERE client_fp <> '';

-- +goose Down
DROP INDEX IF EXISTS uq_content_like_client_fp;

ALTER TABLE content_like
    DROP COLUMN IF EXISTS client_fp;
