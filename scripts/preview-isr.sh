#!/bin/bash
set -euo pipefail

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BOOTSTRAP_API="${PREVIEW_ISR_BOOTSTRAP_API:-http://localhost:8080/api/v2/public/html/posts/refresh}"
SSR_PID=""

stop_ssr_server() {
    if [[ -n "${SSR_PID}" ]] && kill -0 "${SSR_PID}" 2>/dev/null; then
        echo "   Stopping SSR Server..."
        kill "${SSR_PID}" >/dev/null 2>&1 || true
    fi
    SSR_PID=""
}

trap stop_ssr_server EXIT

echo -e "${BLUE}🔍 [0/6] Checking Backend Status...${NC}"
if ! nc -z localhost 8080; then
    echo -e "${RED}Error: Backend server is NOT running on port 8080!${NC}"
    echo "Please start the backend server in a separate terminal:"
    echo "  cd server && go run cmd/api/main.go"
    exit 1
fi
echo -e "${GREEN}   Backend is running!${NC}"

echo -e "${BLUE}🚀 [1/6] Cleaning HTML storage...${NC}"
# 1. 清空目录 (保留 .keep 或 server.js 如果有的话，这里直接清空 html 子目录)
rm -rf server/storage/html/*
mkdir -p server/storage/html

echo -e "${BLUE}📦 [2/6] Building Web Frontend...${NC}"
# 2. 构建前端
cd web
pnpm build

echo -e "${BLUE}📁 [3/6] Copying assets to storage...${NC}"
# 3. 复制构建产物 (client 目录下的所有静态资源)
# SvelteKit 构建输出在 build/client
cp -r build/client/* ../server/storage/html/
cd ..

echo -e "${BLUE}🔌 [4/6] Starting SSR Server (for scraping)...${NC}"
# 4. 后台启动 pnpm serve (SSR)
cd web
# 这里的 pnpm serve 对应生产环境运行 (port 3000)
pnpm serve > /dev/null 2>&1 &
SSR_PID=$!
echo "   SSR Server running with PID: $SSR_PID"

# 等待端口 3000 就绪
echo "   Waiting for port 3000..."
while ! nc -z localhost 3000; do
  sleep 0.5
done
echo -e "${GREEN}   SSR Server is ready!${NC}"
cd ..

echo -e "${BLUE}🔄 [5/6] Triggering Backend ISR Bootstrap...${NC}"
# 5. 调用后端 API 执行一次 ISR Bootstrap
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BOOTSTRAP_API}")

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}   ISR bootstrap completed successfully!${NC}"
else
    echo -e "${RED}   Failed to run ISR bootstrap. Status: $HTTP_STATUS${NC}"
    exit 1
fi

# 任务完成，关闭 SSR 服务器
stop_ssr_server

echo -e "${BLUE}🌍 [6/6] Starting Static Server & Opening Browser...${NC}"
# 6. 启动静态服务器并打开浏览器
# 注意：这里我们用 wait 或者直接 exec 切换进程
# 先打开浏览器 (延迟 1 秒等待 server 启动)
if [[ "$OSTYPE" == "darwin"* ]]; then
    (sleep 1 && open http://localhost:5555) &
else
    (sleep 1 && xdg-open http://localhost:5555) &
fi

# 启动静态服务器 (server.js)
node server/storage/server.js
