# AGENTS.md — GrtBlog Telemetry Collector

## Project Overview
Lightweight standalone service that receives anonymous telemetry snapshots from GrtBlog instances, stores them in PostgreSQL, and exposes data via Grafana and admin API.

## Structure
- All Go source files are in the package root (`package main`) — intentionally flat for a small service.
- `deploy/` — docker-compose + Grafana provisioning.

## Tech Stack
- Go 1.24+, Fiber v2, pgx v5 (no ORM — raw SQL)
- PostgreSQL 17+ (only supported database)
- Grafana for dashboards (pre-provisioned)

## Key Design Decisions
- **No ORM**: pgx direct queries — the schema is small enough that GORM overhead is not justified.
- **Flat package**: Single `main` package — this is a ~600 line service, not a DDD application.
- **Auto-migrate on startup**: Schema DDL runs via `CREATE TABLE IF NOT EXISTS` — no Goose.
- **Passkey auth**: Admin endpoints hidden (404) until first credential is registered. Setup requires `SETUP_TOKEN` env var.
- **Rate limiting**: In-memory per-instance-id, 1 request/hour. No Redis needed at this scale.

## Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9090` | HTTP listen port |
| `DATABASE_URL` | `postgres://...localhost.../grtblog_telemetry` | PostgreSQL DSN |
| `SETUP_TOKEN` | (empty) | Required for first Passkey registration |
| `WEBAUTHN_RP_ID` | `localhost` | WebAuthn Relying Party ID (domain, no scheme) |
| `WEBAUTHN_RP_ORIGIN` | `http://localhost:9090` | Allowed origin for WebAuthn ceremonies |
| `GRAFANA_URL` | `http://localhost:3000` | Internal Grafana URL for reverse proxy |

## Coding Conventions
- Follow standard Go formatting (`gofmt`).
- PostgreSQL-specific SQL is fine (FILTER, jsonb, DISTINCT ON).
- Keep it simple — resist adding abstractions until the service outgrows its flat structure.
