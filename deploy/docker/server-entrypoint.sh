#!/bin/sh
set -eu

# If command args are provided (e.g. goose/status/debug), run them directly.
# This allows reusing the same image for ad-hoc commands.
if [ "$#" -gt 0 ]; then
	exec "$@"
fi

mkdir -p /app/storage/html /app/storage/uploads /app/storage/geoip
chown -R app:app /app/storage

# Inject runtime feature flags into admin SPA
sed -i "s/%%__DANGEROUSLY_ENABLE_BETA_FEDERATION_SETTINGS__%%/${__DANGEROUSLY_ENABLE_BETA_FEDERATION_SETTINGS__:-false}/g" /app/admin/index.html

# Run database migrations before starting server
echo "[entrypoint] Running database migrations..."
goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" up
echo "[entrypoint] Migrations complete."

exec su-exec app /app/grtblog-server
