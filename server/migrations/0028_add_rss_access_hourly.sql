-- +goose Up
CREATE TABLE IF NOT EXISTS analytics_rss_access_hourly
(
    hour_bucket  TIMESTAMPTZ NOT NULL,
    request_path VARCHAR(64) NOT NULL,
    ip           VARCHAR(64) NOT NULL,
    client_name  VARCHAR(128) NOT NULL,
    client_hint  VARCHAR(128),
    user_agent   VARCHAR(512),
    platform     VARCHAR(45),
    browser      VARCHAR(45),
    location     VARCHAR(255),
    requests     BIGINT      NOT NULL DEFAULT 1,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pk_analytics_rss_access_hourly PRIMARY KEY (hour_bucket, request_path, ip, client_name)
);

CREATE INDEX IF NOT EXISTS idx_analytics_rss_access_hourly_hour
    ON analytics_rss_access_hourly (hour_bucket DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_rss_access_hourly_client
    ON analytics_rss_access_hourly (client_name, hour_bucket DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_rss_access_hourly_ip
    ON analytics_rss_access_hourly (ip, hour_bucket DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_analytics_rss_access_hourly_ip;
DROP INDEX IF EXISTS idx_analytics_rss_access_hourly_client;
DROP INDEX IF EXISTS idx_analytics_rss_access_hourly_hour;
DROP TABLE IF EXISTS analytics_rss_access_hourly;
