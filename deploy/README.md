# Deploy (Docker Compose)

## 1) Prepare env

```bash
cd deploy
cp .env.example .env
```

Update at least these values in `.env`:

- `POSTGRES_PASSWORD`
- `AUTH_SECRET`
- `IMAGE_REPO_PREFIX` / `APP_VERSION` (see below)
- `APP_UPDATE_CHECK_ENABLED` / `APP_UPDATE_CHECK_REPO` (后台更新检查来源，默认 GitHub Release)

更新检查策略：Admin 面板打开时触发一次，服务端 30 分钟内复用缓存，不会频繁请求 GitHub API。

### Using prebuilt images from GHCR (recommended)

Every tagged release triggers a GitHub Actions workflow that builds multi-arch (`linux/amd64` + `linux/arm64`) images and pushes them to `ghcr.io/grtsinry43/`.

```ini
IMAGE_REPO_PREFIX=ghcr.io/grtsinry43/
APP_VERSION=1.2.3
```

Tag strategy:
- Stable `v1.2.3` → tags `1.2.3`, `1.2`, `latest`
- Prerelease `v2.0.0-alpha.1` → tag `2.0.0-alpha.1` only (no `latest`)

### Using local builds

Leave `IMAGE_REPO_PREFIX` empty and build from source:

```ini
IMAGE_REPO_PREFIX=
APP_VERSION=local
```

## 2) Start

```bash
mkdir -p storage/html storage/uploads storage/geoip

# Prebuilt images:
docker compose up -d

# Local build:
docker compose up -d --build
```

启动顺序（自动处理）：
1. `postgres` / `redis` 通过 healthcheck 就绪
2. `renderer` 启动，entrypoint 同步 `_app/*` 客户端资源到 `./storage/html`
3. `server` 启动，entrypoint 运行 Goose 数据库迁移后启动应用
4. `nginx` 反向代理所有流量，使用 Docker DNS resolver 自动感知容器 IP 变化

## 2.1) Migration commands

Check status:

```bash
docker compose exec server goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" status
```

Current version:

```bash
docker compose exec server goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" version
```

Rollback one step:

```bash
docker compose exec server goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" down
```

## 2.2) Update app services

```bash
# Update APP_VERSION in .env, then:
docker compose pull server renderer   # prebuilt images
docker compose up -d server renderer
# For local builds: docker compose up -d --build server renderer
```

Nginx 不会被重建。通过 `resolver 127.0.0.11 valid=10s` 自动发现新容器 IP，无需手动 reload。

## 3) Verify

```bash
curl -f http://localhost:${NGINX_PORT:-80}/healthz
curl -f http://localhost:${NGINX_PORT:-80}/health/liveness
```

Admin panel URL: `http://localhost:${NGINX_PORT:-80}/admin/`

## 4) Data layout

- `postgres_data` volume: PostgreSQL data
- `redis_data` volume: Redis AOF data
- `./storage/html`: ISR/HTML snapshots + renderer 客户端资源 (`_app/*`)
- `./storage/uploads`: uploaded files
- `./storage/geoip`: GeoIP db cache

## Routing behavior

- `/api/*` and `/api/v2/ws/*` -> `server`
- `/uploads/*` -> `server`
- `/admin/*` -> `server` (admin SPA 内置于 server 镜像，Fiber 直接 serve)
- `/docs` -> 不在生产 Nginx 代理；仅开发阶段直连后端使用
- other paths -> `nginx try_files` static-first, fallback to `renderer` (adapter-node)

## Notes

- Nginx 使用 Docker 内置 DNS (`resolver 127.0.0.11 valid=10s`) 代替 `upstream` 块，容器重建后最多 10s 自动恢复。
- `renderer` entrypoint 每次启动时清理旧 `_app/` 并拷贝新资源，解决版本堆积问题。
- `server` entrypoint 自动运行数据库迁移，无需单独的 migrate 服务。
- Internal service network: `grtblog-internal`.
- `server` renders snapshot pages from `HTMLSNAPSHOT_BASE_URL=http://renderer:3000`.
- `renderer` SSR calls API via `INTERNAL_API_BASE_URL=http://server:8080/api/v2`.
- Admin SPA 内置于 server 镜像 (`/app/admin/`)，由 Fiber 直接 serve，无需独立容器。
