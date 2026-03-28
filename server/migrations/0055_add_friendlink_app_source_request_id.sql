-- +goose Up
ALTER TABLE friend_link_applications
    ADD COLUMN IF NOT EXISTS source_request_id VARCHAR(64);

-- +goose Down
ALTER TABLE friend_link_applications
    DROP COLUMN IF EXISTS source_request_id;
