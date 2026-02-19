# Deploy (Docker Compose)

## 1) Prepare env

```bash
cd deploy
cp .env.example .env
```

Update at least these values in `.env`:

- `POSTGRES_PASSWORD`
- `AUTH_SECRET`

## 2) Start stack

```bash
docker compose up -d --build
```

`migrate` service will run Goose `up` before `server` starts.
`server` startup entrypoint will fix volume ownership for
`storage/html`, `storage/uploads`, and `storage/geoip` before booting app process.

## 2.1) Migration commands

Check status:

```bash
docker compose run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" status'
```

Current version:

```bash
docker compose run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" version'
```

Rollback one step:

```bash
docker compose run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" down'
```

## 3) Verify

```bash
curl -f http://localhost:${NGINX_PORT:-80}/healthz
curl -f http://localhost:${NGINX_PORT:-80}/health/liveness
```

Admin panel URL: `http://localhost:${NGINX_PORT:-80}/admin/`

## 4) Data & static volumes

- `postgres_data`: PostgreSQL data
- `redis_data`: Redis AOF data
- `html_data`: ISR/HTML snapshot static outputs (`storage/html`)
- `uploads_data`: uploaded files
- `geoip_data`: GeoIP db cache

## Routing behavior

- `/api/*` and `/api/v2/ws/*` -> `server`
- `/uploads/*` -> `server`
- `/admin/*` -> `admin` (Vue SPA static build)
- other paths -> `nginx try_files` static-first, fallback to `renderer` (adapter-node)

## Notes

- Current repo keeps `adapter-node`; this Compose setup is based on that runtime.
- Internal service network: `grtblog-internal` (server/renderer/postgres/redis/nginx all attach to it).
- `server` renders snapshot pages from `HTMLSNAPSHOT_BASE_URL=http://renderer:3000`.
- `renderer` SSR calls API via `INTERNAL_API_BASE_URL=http://server:8080/api/v2`.
- `admin` is built with `VITE_APP_BASE=/admin/`, then served behind gateway path `/admin/`.
