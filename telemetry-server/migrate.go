package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate creates the telemetry tables if they don't exist.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	ddl := `
CREATE TABLE IF NOT EXISTS telemetry_report (
    id            BIGSERIAL PRIMARY KEY,
    instance_id   TEXT        NOT NULL,
    version       TEXT        NOT NULL DEFAULT '',
    go_version    TEXT        NOT NULL DEFAULT '',
    os            TEXT        NOT NULL DEFAULT '',
    arch          TEXT        NOT NULL DEFAULT '',
    deploy_mode   TEXT        NOT NULL DEFAULT '',
    uptime_sec    BIGINT      NOT NULL DEFAULT 0,
    features      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    metrics       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    error_count   INT         NOT NULL DEFAULT 0,
    panic_count   INT         NOT NULL DEFAULT 0,
    payload       JSONB       NOT NULL,
    received_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_report_instance_day
    ON telemetry_report (instance_id, ((received_at AT TIME ZONE 'UTC')::date));

CREATE INDEX IF NOT EXISTS idx_report_version
    ON telemetry_report (version);

CREATE INDEX IF NOT EXISTS idx_report_received
    ON telemetry_report (received_at DESC);

CREATE TABLE IF NOT EXISTS telemetry_error_digest (
    id             BIGSERIAL PRIMARY KEY,
    report_id      BIGINT  NOT NULL REFERENCES telemetry_report(id) ON DELETE CASCADE,
    fingerprint    TEXT    NOT NULL,
    kind           TEXT    NOT NULL DEFAULT '',
    biz_code       TEXT    NOT NULL DEFAULT '',
    location       TEXT    NOT NULL DEFAULT '',
    sample_message TEXT    NOT NULL DEFAULT '',
    count          BIGINT  NOT NULL DEFAULT 0,
    first_seen     TIMESTAMPTZ,
    last_seen      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_digest_fingerprint
    ON telemetry_error_digest (fingerprint);

CREATE INDEX IF NOT EXISTS idx_digest_report
    ON telemetry_error_digest (report_id);

-- WebAuthn credential storage (serialized as JSON).
CREATE TABLE IF NOT EXISTS passkey_credential (
    id              TEXT PRIMARY KEY,
    credential_json JSONB       NOT NULL,
    display_name    TEXT        NOT NULL DEFAULT 'admin',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- WebAuthn ceremony session data (short-lived, between Begin/Finish).
CREATE TABLE IF NOT EXISTS webauthn_session (
    id           TEXT PRIMARY KEY,
    session_json JSONB       NOT NULL,
    session_type TEXT        NOT NULL DEFAULT '',
    expires_at   TIMESTAMPTZ NOT NULL
);

-- Application sessions (long-lived, after successful login).
CREATE TABLE IF NOT EXISTS admin_session (
    token       TEXT PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL
);
`
	_, err := pool.Exec(ctx, ddl)
	return err
}
