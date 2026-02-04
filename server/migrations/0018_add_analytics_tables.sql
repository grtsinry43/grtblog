-- +goose Up
CREATE TABLE IF NOT EXISTS analytics_content_hourly
(
    content_type VARCHAR(20) NOT NULL,
    content_id   BIGINT      NOT NULL,
    hour_bucket  TIMESTAMPTZ NOT NULL,
    pv           BIGINT      NOT NULL DEFAULT 0,
    uv           BIGINT      NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_analytics_content_hourly PRIMARY KEY (content_type, content_id, hour_bucket)
);

CREATE INDEX IF NOT EXISTS idx_analytics_content_hourly_hour ON analytics_content_hourly (hour_bucket DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_content_hourly_type_hour ON analytics_content_hourly (content_type, hour_bucket DESC);

CREATE TABLE IF NOT EXISTS analytics_online_hourly
(
    hour_bucket  TIMESTAMPTZ NOT NULL PRIMARY KEY,
    peak_online  BIGINT      NOT NULL DEFAULT 0,
    sample_total BIGINT      NOT NULL DEFAULT 0,
    sample_count BIGINT      NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_analytics_online_hourly_hour ON analytics_online_hourly (hour_bucket DESC);

-- +goose Down
DROP TABLE IF EXISTS analytics_online_hourly;
DROP TABLE IF EXISTS analytics_content_hourly;
