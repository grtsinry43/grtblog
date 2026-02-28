-- +goose Up
ALTER TABLE analytics_visitor_view
    ADD COLUMN IF NOT EXISTS last_ip VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_analytics_visitor_view_last_ip
    ON analytics_visitor_view (last_ip, last_view_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_analytics_visitor_view_last_ip;

ALTER TABLE analytics_visitor_view
    DROP COLUMN IF EXISTS last_ip;
