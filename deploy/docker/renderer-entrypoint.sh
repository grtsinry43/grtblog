#!/bin/sh
set -eu

# Sync client assets to the shared volume so nginx can serve them
# directly as static files (survives renderer restarts/crashes).
if [ -d /assets ]; then
	echo "[entrypoint] Syncing client assets..."
	rm -rf /assets/_app
	cp -a /app/build/client/. /assets/
	echo "[entrypoint] Client assets synced."
fi

exec node /app/build/index.js
