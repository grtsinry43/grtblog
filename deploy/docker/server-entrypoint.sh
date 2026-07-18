#!/bin/sh
set -eu

# If command args are provided (e.g. goose/status/debug), run them directly.
# This allows reusing the same image for ad-hoc commands.
if [ "$#" -gt 0 ]; then
	exec "$@"
fi

mkdir -p /app/storage/html /app/storage/meta/isr /app/storage/uploads /app/storage/backups /app/storage/geoip
chown -R app:app /app/storage

if [ -f /app/storage/backups/.restore-request.json ] || [ -f /app/storage/backups/.restore-running.json ]; then
	echo "[entrypoint] Pending full-site restore found; restoring before startup..."
	if ! su-exec app /app/grtblog-restore; then
		echo "[entrypoint] Restore failed. The previous site was preserved where possible; starting normally for inspection."
	fi
fi

# Run database migrations before starting server
echo "[entrypoint] Running database migrations..."
goose -table public.goose_db_version -dir /app/migrations postgres "$DB_DSN" up
echo "[entrypoint] Migrations complete."

exec su-exec app /app/grtblog-server
