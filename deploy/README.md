# Deploy (Docker Compose)

## 1) Prepare env

```bash
cd deploy
cp .env.example .env
```

Update at least these values in `.env`:

- `POSTGRES_PASSWORD`
- `AUTH_SECRET`
- `APP_VERSION` (tag 发布用 `v1.2.3`；非 tag 中间构建建议用 commit hash，如 `8f3c1a2b9d4e`)
- `APP_UPDATE_CHECK_ENABLED` / `APP_UPDATE_CHECK_REPO` (后台更新检查来源，默认 GitHub Release)

更新检查策略：Admin 面板打开时触发一次，服务端 30 分钟内复用缓存，不会频繁请求 GitHub API。

## 2) Start app stack (without nginx)

```bash
mkdir -p storage/html storage/uploads storage/geoip
docker compose -f docker-compose.yml up -d --build
```

`migrate` service will run Goose `up` before `server` starts.
`server` startup entrypoint will fix volume ownership for
`storage/html`, `storage/uploads`, and `storage/geoip` before booting app process.

## 2.1) Start gateway stack (nginx only)

```bash
docker network create grtblog-internal >/dev/null 2>&1 || true
docker compose -f docker-compose.gateway.yml up -d
```

`nginx` is now isolated in a separate Compose project.
Updating/restarting app services (`server` / `renderer` / `admin`) will not recreate `nginx`.

## 2.2) Migration commands

Check status:

```bash
docker compose -f docker-compose.yml run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" status'
```

Current version:

```bash
docker compose -f docker-compose.yml run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" version'
```

Rollback one step:

```bash
docker compose -f docker-compose.yml run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" down'
```

## 2.3) Update app services by version (keep nginx untouched)

```bash
# example: APP_VERSION=v1.2.3 in deploy/.env
docker compose -f docker-compose.yml pull server renderer admin
docker compose -f docker-compose.yml run --rm migrate sh -lc 'goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" up'
docker compose -f docker-compose.yml up -d server renderer admin
```

## 3) Verify

```bash
curl -f http://localhost:${NGINX_PORT:-80}/healthz
curl -f http://localhost:${NGINX_PORT:-80}/health/liveness
```

Admin panel URL: `http://localhost:${NGINX_PORT:-80}/admin/`

## 4) Data layout

- `postgres_data` volume: PostgreSQL data
- `redis_data` volume: Redis AOF data
- `./storage/html`: ISR/HTML snapshot static outputs
- `./storage/uploads`: uploaded files
- `./storage/geoip`: GeoIP db cache

## Routing behavior

- `/api/*` and `/api/v2/ws/*` -> `server`
- `/uploads/*` -> `server`
- `/admin/*` -> `admin` (Vue SPA static build)
- other paths -> `nginx try_files` static-first, fallback to `renderer` (adapter-node)

## Notes

- Current repo keeps `adapter-node`; this Compose setup is based on that runtime.
- Internal service network: `grtblog-internal` (app and gateway attach to it).
- `server` renders snapshot pages from `HTMLSNAPSHOT_BASE_URL=http://renderer:3000`.
- `renderer` SSR calls API via `INTERNAL_API_BASE_URL=http://server:8080/api/v2`.
- `admin` is built with `VITE_APP_BASE=/admin/`, then served behind gateway path `/admin/`.
