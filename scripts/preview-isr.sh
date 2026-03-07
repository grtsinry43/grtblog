#!/bin/bash
set -euo pipefail

# Color definitions
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

BOOTSTRAP_API="${PREVIEW_ISR_BOOTSTRAP_API:-http://localhost:8080/api/v2/admin/html/posts/refresh}"
SSR_PID=""

# ---------- Auth ----------
TOKEN="${PREVIEW_ISR_TOKEN:-}"
if [[ -z "${TOKEN}" ]]; then
    echo -e "${RED}Error: PREVIEW_ISR_TOKEN is not set.${NC}"
    echo ""
    echo "This script requires an admin token to call the ISR bootstrap API."
    echo "Create one in the Admin dashboard (Settings > Admin Tokens), then run:"
    echo ""
    echo "  export PREVIEW_ISR_TOKEN=\"gt_your_token_here\""
    echo "  make preview-isr"
    exit 1
fi

# ---------- Helpers ----------
stop_ssr_server() {
    if [[ -n "${SSR_PID}" ]] && kill -0 "${SSR_PID}" 2>/dev/null; then
        echo "   Stopping SSR Server..."
        kill "${SSR_PID}" >/dev/null 2>&1 || true
    fi
    SSR_PID=""
}

trap stop_ssr_server EXIT

# ---------- Pipeline ----------
echo -e "${BLUE}[0/6] Checking Backend Status...${NC}"
if ! nc -z localhost 8080; then
    echo -e "${RED}Error: Backend server is NOT running on port 8080!${NC}"
    echo "Please start the backend server in a separate terminal:"
    echo "  cd server && make run"
    exit 1
fi
echo -e "${GREEN}   Backend is running.${NC}"

echo -e "${BLUE}[1/6] Cleaning HTML storage...${NC}"
rm -rf server/storage/html/*
mkdir -p server/storage/html

echo -e "${BLUE}[2/6] Building Web Frontend...${NC}"
cd web
pnpm build

echo -e "${BLUE}[3/6] Copying assets to storage...${NC}"
cp -r build/client/* ../server/storage/html/
cd ..

echo -e "${BLUE}[4/6] Starting SSR Server (for scraping)...${NC}"
cd web
pnpm serve > /dev/null 2>&1 &
SSR_PID=$!
echo "   SSR Server running with PID: $SSR_PID"

echo "   Waiting for port 3000..."
while ! nc -z localhost 3000; do
  sleep 0.5
done
echo -e "${GREEN}   SSR Server is ready.${NC}"
cd ..

echo -e "${BLUE}[5/6] Triggering Backend ISR Bootstrap...${NC}"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST \
    -H "Authorization: ${TOKEN}" \
    "${BOOTSTRAP_API}")

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}   ISR bootstrap completed successfully.${NC}"
else
    echo -e "${RED}   Failed to run ISR bootstrap. HTTP status: $HTTP_STATUS${NC}"
    exit 1
fi

stop_ssr_server

echo -e "${BLUE}[6/6] Starting Static Server & Opening Browser...${NC}"
if [[ "$OSTYPE" == "darwin"* ]]; then
    (sleep 1 && open http://localhost:5555) &
else
    (sleep 1 && xdg-open http://localhost:5555) &
fi

node server/storage/server.js
