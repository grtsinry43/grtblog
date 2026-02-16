#!/bin/sh
set -eu

# If command args are provided (e.g. goose/status/debug), run them directly.
# This allows reusing the same image for migration jobs.
if [ "$#" -gt 0 ]; then
	exec "$@"
fi

mkdir -p /app/storage/html /app/storage/uploads /app/storage/geoip
chown -R app:app /app/storage

exec su-exec app /app/grtblog-server
