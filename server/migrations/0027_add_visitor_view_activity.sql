-- +goose Up
CREATE TABLE IF NOT EXISTS analytics_visitor_view
(
    visitor_id      VARCHAR(255) NOT NULL,
    content_type    VARCHAR(20)  NOT NULL,
    content_id      BIGINT       NOT NULL,
    platform        VARCHAR(45),
    browser         VARCHAR(45),
    location        VARCHAR(255),
    first_view_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    last_view_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    view_count      BIGINT       NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT pk_analytics_visitor_view PRIMARY KEY (visitor_id, content_type, content_id)
);

CREATE INDEX IF NOT EXISTS idx_analytics_visitor_view_visitor_last
    ON analytics_visitor_view (visitor_id, last_view_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_visitor_view_last
    ON analytics_visitor_view (last_view_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_analytics_visitor_view_last;
DROP INDEX IF EXISTS idx_analytics_visitor_view_visitor_last;
DROP TABLE IF EXISTS analytics_visitor_view;
