#!/bin/sh
set -eu

APP_VERSION="${APP_VERSION:-dev}"
BUILD_COMMIT="${BUILD_COMMIT:-unknown}"
PORT="${PORT:-3000}"
APP_ENV="${APP_ENV:-${NODE_ENV:-production}}"
HOST="${HOST:-0.0.0.0}"

cat <<EOF
================================================================
> grtblog renderer ${APP_VERSION} (${BUILD_COMMIT})
> 不仅是博客，也是全新的内容基础设施。

by @grtsinry43 · github.com/grtsinry43
“代码是写给人看的，顺便在机器上运行的。”

- 渲染模式: SvelteKit adapter-node (SSR)
- 监听地址: ${HOST}
- 监听端口: :${PORT}
- 运行环境: ${APP_ENV}
================================================================
EOF

# Sync client assets to the shared volume so nginx can serve them
# directly as static files (survives renderer restarts/crashes).
if [ -d /assets ]; then
	echo "[entrypoint] Syncing client assets..."
	rm -rf /assets/_app
	cp -a /app/build/client/. /assets/
	echo "[entrypoint] Client assets synced."
fi

exec node /app/build/index.js
