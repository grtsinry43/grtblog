#!/usr/bin/env bash
# GrtBlog v2 — One-Click Deployment Script
# https://github.com/grtsinry43/grtblog
#
# Usage:
#   bash <(curl -fsSL https://raw.githubusercontent.com/grtsinry43/grtblog/main/deploy/install.sh)
#   # China:
#   bash <(curl -fsSL https://cnb.cool/grtsinry43/grtblog/-/git/raw/main/deploy/install.sh)
#
# Non-interactive:
#   GRTBLOG_NONINTERACTIVE=1 APP_VERSION=2.0.2 \
#     IMAGE_REPO_PREFIX=docker.cnb.cool/grtsinry43/grtblog/ \
#     bash <(curl -fsSL ...)
set -euo pipefail
trap 'printf "\n  \033[1;31m✗\033[0m Script failed at line %d (exit code %d).\n    Please report this issue: https://github.com/grtsinry43/grtblog/issues\n" "$LINENO" "$?" >&2' ERR

# ---------------------------------------------------------------------------
# Output helpers (matches scripts/release.sh conventions)
# ---------------------------------------------------------------------------
USE_COLOR="false"
if [[ -t 1 ]] && command -v tput >/dev/null 2>&1 && [[ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]]; then
  USE_COLOR="true"
fi

_c() {
  if [[ "$USE_COLOR" == "true" ]]; then
    printf '\033[%sm' "$1"
  fi
}

section() {
  printf '\n%s==>%s %s\n' "$(_c '1;36')" "$(_c '0')" "$1"
}

info() {
  printf '  %s-%s %s\n' "$(_c '0;32')" "$(_c '0')" "$1"
}

warn() {
  printf '  %s!%s %s\n' "$(_c '1;33')" "$(_c '0')" "$1" >&2
}

err() {
  printf '  %s✗%s %s\n' "$(_c '1;31')" "$(_c '0')" "$1" >&2
}

ok() {
  printf '  %s✓%s %s\n' "$(_c '1;32')" "$(_c '0')" "$1"
}

prompt_value() {
  printf '  %s>%s %s' "$(_c '1;35')" "$(_c '0')" "$1"
}

# ---------------------------------------------------------------------------
# Utilities
# ---------------------------------------------------------------------------
NONINTERACTIVE="${GRTBLOG_NONINTERACTIVE:-0}"

ask() {
  # ask "prompt" default_value variable_name
  local prompt="$1" default="$2" varname="$3"
  if [[ "$NONINTERACTIVE" == "1" ]]; then
    eval "$varname=\"$default\""
    return
  fi
  local answer
  prompt_value "$prompt [${default}]: "
  read -r answer
  if [[ -z "$answer" ]]; then
    eval "$varname=\"$default\""
  else
    eval "$varname=\"$answer\""
  fi
}

ask_yn() {
  # ask_yn "question" default(y/n) — returns 0 for yes, 1 for no
  local prompt="$1" default="$2"
  if [[ "$NONINTERACTIVE" == "1" ]]; then
    [[ "$default" == "y" ]] && return 0 || return 1
  fi
  local answer
  prompt_value "$prompt [${default}]: "
  read -r answer
  answer="${answer:-$default}"
  [[ "$answer" =~ ^[Yy] ]] && return 0 || return 1
}

choose() {
  # choose "title" option1 option2 ... — sets CHOICE (1-based index)
  local title="$1"; shift
  local options=("$@")
  if [[ "$NONINTERACTIVE" == "1" ]]; then
    CHOICE=1
    return
  fi
  printf '\n'
  info "$title"
  local i=1
  for opt in "${options[@]}"; do
    printf '    %s%d)%s %s\n' "$(_c '1;35')" "$i" "$(_c '0')" "$opt"
    i=$((i + 1))
  done
  local answer
  prompt_value "$(__ ENTER_CHOICE) [1]: "
  read -r answer
  answer="${answer:-1}"
  if [[ "$answer" =~ ^[0-9]+$ ]] && [[ "$answer" -ge 1 ]] && [[ "$answer" -le "${#options[@]}" ]]; then
    CHOICE="$answer"
  else
    CHOICE=1
  fi
}

http_get() {
  # http_get URL output_file — returns 0 on success
  local url="$1" output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --connect-timeout 10 --max-time 60 "$url" -o "$output" 2>/dev/null
  elif command -v wget >/dev/null 2>&1; then
    wget -q --timeout=10 -O "$output" "$url" 2>/dev/null
  else
    return 1
  fi
}

http_get_stdout() {
  # http_get_stdout URL — prints content to stdout
  local url="$1"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --connect-timeout 10 --max-time 60 "$url" 2>/dev/null
  elif command -v wget >/dev/null 2>&1; then
    wget -q --timeout=10 -O - "$url" 2>/dev/null
  else
    return 1
  fi
}

http_check() {
  # http_check URL timeout — returns 0 if reachable
  local url="$1" timeout="${2:-3}"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --connect-timeout "$timeout" --max-time "$timeout" "$url" -o /dev/null 2>/dev/null
  elif command -v wget >/dev/null 2>&1; then
    wget -q --timeout="$timeout" --tries=1 -O /dev/null "$url" 2>/dev/null
  else
    return 1
  fi
}

random_hex() {
  local length="${1:-32}"
  local result=""
  if command -v openssl >/dev/null 2>&1; then
    result="$(openssl rand -hex "$length" 2>/dev/null)" || true
  fi
  if [[ -z "$result" ]]; then
    # Fallback to /dev/urandom
    result="$(head -c "$length" /dev/urandom 2>/dev/null | od -An -tx1 | tr -d ' \n' | head -c "$((length * 2))")" || true
  fi
  printf '%s\n' "$result"
}

# ---------------------------------------------------------------------------
# i18n — Chinese / English bilingual support
# ---------------------------------------------------------------------------
LANG_CODE="${GRTBLOG_LANG:-}"

__() {
  # __ KEY — prints the localized string for KEY
  local key="$1"
  local varname="MSG_${LANG_CODE}_${key}"
  printf '%s' "${!varname:-$key}"
}

# --- English strings ---
MSG_en_ENTER_CHOICE="Enter choice"
MSG_en_LANG_SELECT="Select language / 选择语言:"
MSG_en_LANG_EN="English"
MSG_en_LANG_ZH="简体中文"
MSG_en_STEP1="Step 1/11: Environment Check"
MSG_en_NO_CURL_WGET="Neither curl nor wget found. Please install one of them."
MSG_en_NO_DOCKER="Docker is not installed. Please install Docker first."
MSG_en_NO_COMPOSE="Docker Compose is not installed."
MSG_en_NO_DOCKER_PERM="Cannot connect to Docker daemon. Is it running? Do you have permission?"
MSG_en_DOCKER_PERM_TIP="Try: sudo usermod -aG docker \$USER && newgrp docker"
MSG_en_DOCKER_OK="Docker daemon is accessible"
MSG_en_STEP2="Step 2/11: Detect Existing Installation"
MSG_en_EXISTING_FOUND="Existing installation detected in current directory."
MSG_en_EXISTING_MENU="What would you like to do?"
MSG_en_EXISTING_UPGRADE="Upgrade (keep .env, update config files and version)"
MSG_en_EXISTING_REINSTALL="Reinstall (backup .env, fresh install)"
MSG_en_EXISTING_EXIT="Exit"
MSG_en_MODE_UPGRADE="Upgrade mode selected"
MSG_en_ENV_BACKED_UP="Existing .env backed up to"
MSG_en_EXITING="Exiting."
MSG_en_NO_EXISTING="No existing installation found. Proceeding with fresh install."
MSG_en_STEP3="Step 3/11: Create Directories"
MSG_en_STEP4="Step 4/11: Select Update Channel"
MSG_en_CHANNEL_MENU="Select update channel:"
MSG_en_CHANNEL_STABLE="stable (recommended — production releases only)"
MSG_en_CHANNEL_PREVIEW="preview (includes alpha / beta / rc releases)"
MSG_en_CHANNEL_SET="Update channel"
MSG_en_STEP5="Step 5/11: Network Detection & Source Selection"
MSG_en_NET_TESTING="Testing network connectivity..."
MSG_en_NET_CHINA="Google not reachable — assuming China mainland network"
MSG_en_NET_INTL="Google reachable — assuming international network"
MSG_en_SOURCE_MENU="Select image source:"
MSG_en_SOURCE_CNB_REC="CNB — docker.cnb.cool (recommended for China)"
MSG_en_SOURCE_DOCKERHUB="Docker Hub — docker.io"
MSG_en_SOURCE_DOCKERHUB_REC="Docker Hub — docker.io (recommended)"
MSG_en_SOURCE_GHCR="GHCR — ghcr.io"
MSG_en_SOURCE_GHCR_TEST="GHCR — ghcr.io (for preview/testing)"
MSG_en_SOURCE_CNB="CNB — docker.cnb.cool (for China)"
MSG_en_IMAGE_SOURCE="Image source"
MSG_en_CONFIG_SOURCE="Config source"
MSG_en_STEP6="Step 6/11: Fetch Latest Version"
MSG_en_FETCH_STABLE="Fetching latest stable release from GitHub..."
MSG_en_FETCH_PREVIEW="Fetching latest preview tag from GitHub..."
MSG_en_FETCH_FALLBACK="No preview tags found, falling back to latest stable release"
MSG_en_DETECTED_VER="Detected version"
MSG_en_USE_VERSION="Use this version?"
MSG_en_FETCH_FAIL="Could not detect latest version from GitHub API."
MSG_en_ENTER_VERSION="Enter version manually (e.g. 2.0.2)"
MSG_en_VERSION="Version"
MSG_en_STEP7="Step 7/11: Generate Credentials"
MSG_en_KEEP_CREDS="Keeping existing credentials from .env"
MSG_en_GEN_PGPASS="Generated POSTGRES_PASSWORD"
MSG_en_GEN_SECRET="Generated AUTH_SECRET"
MSG_en_PORT_IN_USE="Port 80 appears to be in use."
MSG_en_ENTER_PORT="Enter a different port"
MSG_en_REVIEW_CREDS="Review/modify credentials before continuing?"
MSG_en_CREDS_READY="Credentials ready"
MSG_en_STEP8="Step 8/11: Download Config Files"
MSG_en_DOWNLOADING="Downloading"
MSG_en_DOWNLOADED="downloaded"
MSG_en_DOWNLOAD_FALLBACK="Primary source failed, trying fallback..."
MSG_en_DOWNLOADED_FB="downloaded (fallback)"
MSG_en_DOWNLOAD_FAIL="Failed to download from all sources"
MSG_en_STEP9="Step 9/11: Generate .env"
MSG_en_UPGRADE_ENV="Upgrade mode: updating version and image source in .env"
MSG_en_ENV_UPDATED=".env updated (upgrade)"
MSG_en_ENV_CREATED=".env created"
MSG_en_STEP10="Step 10/11: Pull & Start"
MSG_en_PULLING="Pulling images..."
MSG_en_PULL_FAIL="Failed to pull images."
MSG_en_PULL_CHECK="Check your IMAGE_REPO_PREFIX and APP_VERSION settings in .env"
MSG_en_PULLED="Images pulled"
MSG_en_STARTING="Starting services..."
MSG_en_START_FAIL="Failed to start services."
MSG_en_LOGS_TITLE="Container Logs"
MSG_en_STARTED="Services started"
MSG_en_STEP11="Step 11/11: Health Check"
MSG_en_HEALTH_WAIT="Waiting for health check at"
MSG_en_HEALTH_OK="Health check passed!"
MSG_en_HEALTH_TIMEOUT="Health check timed out after"
MSG_en_HEALTH_TIP="The services may still be starting. Check logs with:"
MSG_en_DEPLOY_DONE="Deployment Complete"
MSG_en_BLOG="Blog"
MSG_en_ADMIN="Admin"
MSG_en_CREDS_SAVED="Credentials (saved in .env):"
MSG_en_USEFUL_CMDS="Useful commands:"
MSG_en_VIEW_LOGS="View logs"
MSG_en_CHECK_STATUS="Check status"
MSG_en_STOP_SERVICES="Stop services"
MSG_en_UPGRADE_LATER="To upgrade later:"
MSG_en_DOCS="Documentation"
MSG_en_DONE="Done!"
MSG_en_WAITING="waiting"

# --- Chinese strings ---
MSG_zh_ENTER_CHOICE="请输入选项"
MSG_zh_LANG_SELECT="Select language / 选择语言:"
MSG_zh_LANG_EN="English"
MSG_zh_LANG_ZH="简体中文"
MSG_zh_STEP1="步骤 1/11: 环境检查"
MSG_zh_NO_CURL_WGET="未找到 curl 或 wget，请先安装其中一个。"
MSG_zh_NO_DOCKER="未安装 Docker，请先安装 Docker。"
MSG_zh_NO_COMPOSE="未安装 Docker Compose。"
MSG_zh_NO_DOCKER_PERM="无法连接 Docker 守护进程。是否已启动？当前用户是否有权限？"
MSG_zh_DOCKER_PERM_TIP="尝试: sudo usermod -aG docker \$USER && newgrp docker"
MSG_zh_DOCKER_OK="Docker 守护进程可访问"
MSG_zh_STEP2="步骤 2/11: 检测已有安装"
MSG_zh_EXISTING_FOUND="在当前目录检测到已有安装。"
MSG_zh_EXISTING_MENU="请选择操作:"
MSG_zh_EXISTING_UPGRADE="升级（保留 .env，更新配置文件和版本）"
MSG_zh_EXISTING_REINSTALL="重新安装（备份 .env，全新安装）"
MSG_zh_EXISTING_EXIT="退出"
MSG_zh_MODE_UPGRADE="已选择升级模式"
MSG_zh_ENV_BACKED_UP="已有 .env 已备份至"
MSG_zh_EXITING="退出。"
MSG_zh_NO_EXISTING="未检测到已有安装，将进行全新安装。"
MSG_zh_STEP3="步骤 3/11: 创建目录"
MSG_zh_STEP4="步骤 4/11: 选择更新通道"
MSG_zh_CHANNEL_MENU="选择更新通道:"
MSG_zh_CHANNEL_STABLE="stable（推荐 — 仅正式版本）"
MSG_zh_CHANNEL_PREVIEW="preview（包含 alpha / beta / rc 版本）"
MSG_zh_CHANNEL_SET="更新通道"
MSG_zh_STEP5="步骤 5/11: 网络检测与源选择"
MSG_zh_NET_TESTING="正在检测网络连通性..."
MSG_zh_NET_CHINA="无法访问 Google — 判定为国内网络"
MSG_zh_NET_INTL="可以访问 Google — 判定为国际网络"
MSG_zh_SOURCE_MENU="选择镜像源:"
MSG_zh_SOURCE_CNB_REC="CNB — docker.cnb.cool（国内推荐）"
MSG_zh_SOURCE_DOCKERHUB="Docker Hub — docker.io"
MSG_zh_SOURCE_DOCKERHUB_REC="Docker Hub — docker.io（推荐）"
MSG_zh_SOURCE_GHCR="GHCR — ghcr.io"
MSG_zh_SOURCE_GHCR_TEST="GHCR — ghcr.io（用于预发布/测试）"
MSG_zh_SOURCE_CNB="CNB — docker.cnb.cool（国内用户）"
MSG_zh_IMAGE_SOURCE="镜像源"
MSG_zh_CONFIG_SOURCE="配置文件源"
MSG_zh_STEP6="步骤 6/11: 获取最新版本"
MSG_zh_FETCH_STABLE="正在从 GitHub 获取最新正式版本..."
MSG_zh_FETCH_PREVIEW="正在从 GitHub 获取最新预发布标签..."
MSG_zh_FETCH_FALLBACK="未找到预发布标签，回退到最新正式版本"
MSG_zh_DETECTED_VER="检测到版本"
MSG_zh_USE_VERSION="使用此版本？"
MSG_zh_FETCH_FAIL="无法从 GitHub API 获取最新版本。"
MSG_zh_ENTER_VERSION="请手动输入版本号（例如 2.0.2）"
MSG_zh_VERSION="版本"
MSG_zh_STEP7="步骤 7/11: 生成凭据"
MSG_zh_KEEP_CREDS="保留 .env 中已有的凭据"
MSG_zh_GEN_PGPASS="已生成 POSTGRES_PASSWORD"
MSG_zh_GEN_SECRET="已生成 AUTH_SECRET"
MSG_zh_PORT_IN_USE="端口 80 似乎已被占用。"
MSG_zh_ENTER_PORT="请输入其他端口"
MSG_zh_REVIEW_CREDS="是否在继续前查看/修改凭据？"
MSG_zh_CREDS_READY="凭据准备就绪"
MSG_zh_STEP8="步骤 8/11: 下载配置文件"
MSG_zh_DOWNLOADING="正在下载"
MSG_zh_DOWNLOADED="下载完成"
MSG_zh_DOWNLOAD_FALLBACK="主源下载失败，尝试备用源..."
MSG_zh_DOWNLOADED_FB="下载完成（备用源）"
MSG_zh_DOWNLOAD_FAIL="所有源均下载失败"
MSG_zh_STEP9="步骤 9/11: 生成 .env"
MSG_zh_UPGRADE_ENV="升级模式：更新 .env 中的版本和镜像源"
MSG_zh_ENV_UPDATED=".env 已更新（升级）"
MSG_zh_ENV_CREATED=".env 已创建"
MSG_zh_STEP10="步骤 10/11: 拉取镜像并启动"
MSG_zh_PULLING="正在拉取镜像..."
MSG_zh_PULL_FAIL="镜像拉取失败。"
MSG_zh_PULL_CHECK="请检查 .env 中的 IMAGE_REPO_PREFIX 和 APP_VERSION 设置"
MSG_zh_PULLED="镜像拉取完成"
MSG_zh_STARTING="正在启动服务..."
MSG_zh_START_FAIL="服务启动失败。"
MSG_zh_LOGS_TITLE="容器日志"
MSG_zh_STARTED="服务已启动"
MSG_zh_STEP11="步骤 11/11: 健康检查"
MSG_zh_HEALTH_WAIT="正在等待健康检查"
MSG_zh_HEALTH_OK="健康检查通过！"
MSG_zh_HEALTH_TIMEOUT="健康检查超时，已等待"
MSG_zh_HEALTH_TIP="服务可能仍在启动中，请查看日志："
MSG_zh_DEPLOY_DONE="部署完成"
MSG_zh_BLOG="博客首页"
MSG_zh_ADMIN="管理后台"
MSG_zh_CREDS_SAVED="凭据（已保存在 .env 中）:"
MSG_zh_USEFUL_CMDS="常用命令:"
MSG_zh_VIEW_LOGS="查看日志"
MSG_zh_CHECK_STATUS="查看状态"
MSG_zh_STOP_SERVICES="停止服务"
MSG_zh_UPGRADE_LATER="后续升级命令:"
MSG_zh_DOCS="文档"
MSG_zh_DONE="完成！"
MSG_zh_WAITING="等待中"

# ---------------------------------------------------------------------------
# Language selection (before anything else)
# ---------------------------------------------------------------------------
if [[ -z "$LANG_CODE" ]]; then
  if [[ "$NONINTERACTIVE" == "1" ]]; then
    LANG_CODE="en"
  else
    printf '\n%s==>%s Select language / 选择语言:\n' "$(_c '1;36')" "$(_c '0')"
    printf '    %s1)%s English\n' "$(_c '1;35')" "$(_c '0')"
    printf '    %s2)%s 简体中文\n' "$(_c '1;35')" "$(_c '0')"
    printf '  %s>%s Choose / 请选择 [1]: ' "$(_c '1;35')" "$(_c '0')"
    read -r _lang_choice
    _lang_choice="${_lang_choice:-1}"
    case "$_lang_choice" in
      2) LANG_CODE="zh" ;;
      *) LANG_CODE="en" ;;
    esac
  fi
fi

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
INSTALL_MODE="fresh"       # fresh | upgrade
COMPOSE_CMD=""
APP_VERSION="${APP_VERSION:-}"
APP_UPDATE_CHANNEL="${APP_UPDATE_CHANNEL:-}"
IMAGE_REPO_PREFIX="${IMAGE_REPO_PREFIX:-}"
NGINX_PORT="${NGINX_PORT:-80}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"
AUTH_SECRET="${AUTH_SECRET:-}"

GITHUB_RAW_BASE="https://raw.githubusercontent.com/grtsinry43/grtblog/main"
CNB_RAW_BASE="https://cnb.cool/grtsinry43/grtblog/-/git/raw/main"
CONFIG_BASE_URL=""

REPO_DOCKERHUB="grtsinry43/"
REPO_GHCR="ghcr.io/grtsinry43/"
REPO_CNB="docker.cnb.cool/grtsinry43/grtblog/"

# =========================================================================
# Step 1: Environment Check
# =========================================================================
section "$(__ STEP1)"

# curl or wget
if command -v curl >/dev/null 2>&1; then
  info "curl: $(curl --version | head -n1)"
elif command -v wget >/dev/null 2>&1; then
  info "wget: $(wget --version | head -n1)"
else
  err "$(__ NO_CURL_WGET)"
  exit 1
fi

# Docker
if ! command -v docker >/dev/null 2>&1; then
  err "$(__ NO_DOCKER)"
  info "https://docs.docker.com/engine/install/"
  exit 1
fi
info "docker: $(docker --version)"

# Docker Compose
if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD="docker compose"
  info "compose: $(docker compose version)"
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD="docker-compose"
  info "compose: $(docker-compose --version)"
else
  err "$(__ NO_COMPOSE)"
  info "https://docs.docker.com/compose/install/"
  exit 1
fi

# Docker permission
if ! docker info >/dev/null 2>&1; then
  err "$(__ NO_DOCKER_PERM)"
  info "$(__ DOCKER_PERM_TIP)"
  exit 1
fi
ok "$(__ DOCKER_OK)"

# =========================================================================
# Step 2: Detect Existing Installation
# =========================================================================
section "$(__ STEP2)"

if [[ -f "docker-compose.yml" ]] && [[ -f ".env" ]]; then
  warn "$(__ EXISTING_FOUND)"
  choose "$(__ EXISTING_MENU)" \
    "$(__ EXISTING_UPGRADE)" \
    "$(__ EXISTING_REINSTALL)" \
    "$(__ EXISTING_EXIT)"

  case "$CHOICE" in
    1)
      INSTALL_MODE="upgrade"
      info "$(__ MODE_UPGRADE)"
      ;;
    2)
      INSTALL_MODE="fresh"
      BACKUP_FILE=".env.backup.$(date +%Y%m%d%H%M%S)"
      cp .env "$BACKUP_FILE"
      info "$(__ ENV_BACKED_UP) ${BACKUP_FILE}"
      ;;
    3)
      info "$(__ EXITING)"
      exit 0
      ;;
  esac
else
  info "$(__ NO_EXISTING)"
fi

# =========================================================================
# Step 3: Create Directories
# =========================================================================
section "$(__ STEP3)"

mkdir -p storage/html storage/uploads storage/geoip nginx
ok "storage/html"
ok "storage/uploads"
ok "storage/geoip"
ok "nginx/"

# =========================================================================
# Step 4: Update Channel
# =========================================================================
section "$(__ STEP4)"

if [[ -z "$APP_UPDATE_CHANNEL" ]]; then
  choose "$(__ CHANNEL_MENU)" \
    "$(__ CHANNEL_STABLE)" \
    "$(__ CHANNEL_PREVIEW)"
  case "$CHOICE" in
    1) APP_UPDATE_CHANNEL="stable" ;;
    2) APP_UPDATE_CHANNEL="preview" ;;
  esac
fi
info "$(__ CHANNEL_SET): ${APP_UPDATE_CHANNEL}"

# =========================================================================
# Step 5: Network Detection & Source Selection
# =========================================================================
section "$(__ STEP5)"

IS_CHINA="false"
info "$(__ NET_TESTING)"

if ! http_check "https://www.google.com" 3; then
  IS_CHINA="true"
  info "$(__ NET_CHINA)"
else
  info "$(__ NET_INTL)"
fi

if [[ -z "$IMAGE_REPO_PREFIX" ]]; then
  if [[ "$IS_CHINA" == "true" ]]; then
    choose "$(__ SOURCE_MENU)" \
      "$(__ SOURCE_CNB_REC)" \
      "$(__ SOURCE_DOCKERHUB)" \
      "$(__ SOURCE_GHCR_TEST)"
  else
    choose "$(__ SOURCE_MENU)" \
      "$(__ SOURCE_DOCKERHUB_REC)" \
      "$(__ SOURCE_GHCR)" \
      "$(__ SOURCE_CNB)"
  fi

  if [[ "$IS_CHINA" == "true" ]]; then
    case "$CHOICE" in
      1) IMAGE_REPO_PREFIX="$REPO_CNB";      CONFIG_BASE_URL="$CNB_RAW_BASE" ;;
      2) IMAGE_REPO_PREFIX="$REPO_DOCKERHUB"; CONFIG_BASE_URL="$GITHUB_RAW_BASE" ;;
      3) IMAGE_REPO_PREFIX="$REPO_GHCR";      CONFIG_BASE_URL="$GITHUB_RAW_BASE" ;;
    esac
  else
    case "$CHOICE" in
      1) IMAGE_REPO_PREFIX="$REPO_DOCKERHUB"; CONFIG_BASE_URL="$GITHUB_RAW_BASE" ;;
      2) IMAGE_REPO_PREFIX="$REPO_GHCR";      CONFIG_BASE_URL="$GITHUB_RAW_BASE" ;;
      3) IMAGE_REPO_PREFIX="$REPO_CNB";      CONFIG_BASE_URL="$CNB_RAW_BASE" ;;
    esac
  fi
else
  # Determine config base URL from image repo prefix
  if [[ "$IMAGE_REPO_PREFIX" == *"cnb"* ]]; then
    CONFIG_BASE_URL="$CNB_RAW_BASE"
  else
    CONFIG_BASE_URL="$GITHUB_RAW_BASE"
  fi
fi

info "$(__ IMAGE_SOURCE): ${IMAGE_REPO_PREFIX}"
info "$(__ CONFIG_SOURCE): ${CONFIG_BASE_URL}"

# =========================================================================
# Step 6: Fetch Latest Version
# =========================================================================
section "$(__ STEP6)"

if [[ -z "$APP_VERSION" ]]; then
  FETCHED_VERSION=""

  if [[ "$APP_UPDATE_CHANNEL" == "stable" ]]; then
    info "$(__ FETCH_STABLE)"
    API_RESPONSE="$(http_get_stdout "https://api.github.com/repos/grtsinry43/grtblog/releases/latest" || true)"
    if [[ -n "$API_RESPONSE" ]]; then
      FETCHED_VERSION="$(printf '%s' "$API_RESPONSE" | grep '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' | head -n1)"
      # Strip leading 'v' if present
      FETCHED_VERSION="${FETCHED_VERSION#v}"
    fi
  else
    info "$(__ FETCH_PREVIEW)"
    API_RESPONSE="$(http_get_stdout "https://api.github.com/repos/grtsinry43/grtblog/git/refs/tags" || true)"
    if [[ -n "$API_RESPONSE" ]]; then
      # Find all tags with pre-release suffixes, pick the last one
      FETCHED_VERSION="$(printf '%s' "$API_RESPONSE" | grep '"ref"' | grep -E '(alpha|beta|rc)' | sed 's/.*refs\/tags\/v\{0,1\}\([^"]*\)".*/\1/' | tail -n1)"
      FETCHED_VERSION="${FETCHED_VERSION#v}"
    fi
    # If no preview found, fall back to latest stable
    if [[ -z "$FETCHED_VERSION" ]]; then
      warn "$(__ FETCH_FALLBACK)"
      API_RESPONSE="$(http_get_stdout "https://api.github.com/repos/grtsinry43/grtblog/releases/latest" || true)"
      if [[ -n "$API_RESPONSE" ]]; then
        FETCHED_VERSION="$(printf '%s' "$API_RESPONSE" | grep '"tag_name"' | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' | head -n1)"
        FETCHED_VERSION="${FETCHED_VERSION#v}"
      fi
    fi
  fi

  if [[ -n "$FETCHED_VERSION" ]]; then
    info "$(__ DETECTED_VER): ${FETCHED_VERSION}"
    ask "$(__ USE_VERSION)" "$FETCHED_VERSION" APP_VERSION
  else
    warn "$(__ FETCH_FAIL)"
    ask "$(__ ENTER_VERSION)" "2.0.2" APP_VERSION
  fi
fi

APP_VERSION="${APP_VERSION#v}"
info "$(__ VERSION): ${APP_VERSION}"

# =========================================================================
# Step 7: Generate Credentials
# =========================================================================
section "$(__ STEP7)"

if [[ "$INSTALL_MODE" == "upgrade" ]]; then
  # Read existing credentials from .env
  if [[ -f ".env" ]]; then
    # shellcheck disable=SC1091
    set +u
    source .env 2>/dev/null || true
    set -u
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"
    AUTH_SECRET="${AUTH_SECRET:-}"
    NGINX_PORT="${NGINX_PORT:-80}"
  fi
  info "$(__ KEEP_CREDS)"
else
  # Generate new credentials if not provided via env
  if [[ -z "$POSTGRES_PASSWORD" ]]; then
    POSTGRES_PASSWORD="$(random_hex 32)"
  fi
  if [[ -z "$AUTH_SECRET" ]]; then
    AUTH_SECRET="$(random_hex 32)"
  fi

  info "$(__ GEN_PGPASS): ${POSTGRES_PASSWORD:0:8}..."
  info "$(__ GEN_SECRET): ${AUTH_SECRET:0:8}..."
  info "NGINX_PORT: ${NGINX_PORT}"

  # Check if port 80 is in use
  if [[ "$NGINX_PORT" == "80" ]]; then
    if command -v ss >/dev/null 2>&1; then
      if ss -tlnp 2>/dev/null | grep -q ":80 "; then
        warn "$(__ PORT_IN_USE)"
        ask "$(__ ENTER_PORT)" "8080" NGINX_PORT
      fi
    elif command -v netstat >/dev/null 2>&1; then
      if netstat -tlnp 2>/dev/null | grep -q ":80 "; then
        warn "$(__ PORT_IN_USE)"
        ask "$(__ ENTER_PORT)" "8080" NGINX_PORT
      fi
    fi
  fi

  if [[ "$NONINTERACTIVE" != "1" ]]; then
    if ask_yn "$(__ REVIEW_CREDS)" "n"; then
      ask "POSTGRES_PASSWORD" "$POSTGRES_PASSWORD" POSTGRES_PASSWORD
      ask "AUTH_SECRET" "$AUTH_SECRET" AUTH_SECRET
      ask "NGINX_PORT" "$NGINX_PORT" NGINX_PORT
    fi
  fi
fi

ok "$(__ CREDS_READY)"

# =========================================================================
# Step 8: Download Config Files
# =========================================================================
section "$(__ STEP8)"

download_with_fallback() {
  local filename="$1"
  local output="$2"
  local primary_url="${CONFIG_BASE_URL}/deploy/${filename}"
  local fallback_url=""

  # Set fallback
  if [[ "$CONFIG_BASE_URL" == "$GITHUB_RAW_BASE" ]]; then
    fallback_url="${CNB_RAW_BASE}/deploy/${filename}"
  else
    fallback_url="${GITHUB_RAW_BASE}/deploy/${filename}"
  fi

  info "$(__ DOWNLOADING) ${filename}..."
  if http_get "$primary_url" "$output"; then
    ok "${filename} $(__ DOWNLOADED)"
    return 0
  fi

  warn "$(__ DOWNLOAD_FALLBACK)"
  if http_get "$fallback_url" "$output"; then
    ok "${filename} $(__ DOWNLOADED_FB)"
    return 0
  fi

  err "$(__ DOWNLOAD_FAIL): ${filename}"
  return 1
}

download_with_fallback "docker-compose.yml" "docker-compose.yml" || exit 1
download_with_fallback "nginx/nginx.conf" "nginx/nginx.conf" || exit 1

# =========================================================================
# Step 9: Generate .env
# =========================================================================
section "$(__ STEP9)"

if [[ "$INSTALL_MODE" == "upgrade" ]]; then
  # Upgrade: only update APP_VERSION, IMAGE_REPO_PREFIX, APP_UPDATE_CHANNEL
  info "$(__ UPGRADE_ENV)"

  # Use sed to update specific keys in-place
  if grep -q '^APP_VERSION=' .env; then
    sed -i.bak "s|^APP_VERSION=.*|APP_VERSION=${APP_VERSION}|" .env
  else
    printf '\nAPP_VERSION=%s\n' "$APP_VERSION" >> .env
  fi

  if grep -q '^IMAGE_REPO_PREFIX=' .env; then
    sed -i.bak "s|^IMAGE_REPO_PREFIX=.*|IMAGE_REPO_PREFIX=${IMAGE_REPO_PREFIX}|" .env
  else
    printf '\nIMAGE_REPO_PREFIX=%s\n' "$IMAGE_REPO_PREFIX" >> .env
  fi

  if grep -q '^APP_UPDATE_CHANNEL=' .env; then
    sed -i.bak "s|^APP_UPDATE_CHANNEL=.*|APP_UPDATE_CHANNEL=${APP_UPDATE_CHANNEL}|" .env
  else
    printf '\nAPP_UPDATE_CHANNEL=%s\n' "$APP_UPDATE_CHANNEL" >> .env
  fi

  rm -f .env.bak
  ok "$(__ ENV_UPDATED)"
else
  # Fresh install: write complete .env
  cat > .env <<EOF
APP_VERSION=${APP_VERSION}
IMAGE_REPO_PREFIX=${IMAGE_REPO_PREFIX}

NGINX_PORT=${NGINX_PORT}

POSTGRES_DB=grtblog
POSTGRES_USER=postgres
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

REDIS_PASSWORD=
REDIS_PREFIX=grtblog:

AUTH_SECRET=${AUTH_SECRET}

APP_UPDATE_CHECK_ENABLED=true
APP_UPDATE_CHECK_REPO=grtsinry43/grtblog-v2
APP_UPDATE_CHANNEL=${APP_UPDATE_CHANNEL}

# Admin panel (build-time, baked into JS bundle)
VITE_APP_BASE=/admin/
VITE_APP_NAME=Grtblog Admin
VITE_APP_TITLE=管理后台
VITE_WATERMARK_CONTENT=
VITE_API_BASE_URL=/api/v2
EOF
  ok "$(__ ENV_CREATED)"
fi

# =========================================================================
# Step 10: Pull & Start
# =========================================================================
section "$(__ STEP10)"

info "$(__ PULLING)"
if ! $COMPOSE_CMD pull; then
  err "$(__ PULL_FAIL)"
  warn "$(__ PULL_CHECK)"
  warn "IMAGE_REPO_PREFIX=${IMAGE_REPO_PREFIX} APP_VERSION=${APP_VERSION}"
  exit 1
fi
ok "$(__ PULLED)"

info "$(__ STARTING)"
if ! $COMPOSE_CMD up -d; then
  err "$(__ START_FAIL)"
  section "$(__ LOGS_TITLE)"
  $COMPOSE_CMD logs --tail=50 2>&1 || true
  exit 1
fi
ok "$(__ STARTED)"

# =========================================================================
# Step 11: Health Check & Result
# =========================================================================
section "$(__ STEP11)"

HEALTH_URL="http://127.0.0.1:${NGINX_PORT}/healthz"
MAX_WAIT=120
INTERVAL=5
ELAPSED=0
HEALTHY="false"

info "$(__ HEALTH_WAIT) ${HEALTH_URL}..."

while [[ "$ELAPSED" -lt "$MAX_WAIT" ]]; do
  if http_check "$HEALTH_URL" 5; then
    HEALTHY="true"
    break
  fi
  sleep "$INTERVAL"
  ELAPSED=$((ELAPSED + INTERVAL))
  printf '  . %s (%ds/%ds)\r' "$(__ WAITING)" "$ELAPSED" "$MAX_WAIT"
done
printf '\n'

if [[ "$HEALTHY" == "true" ]]; then
  ok "$(__ HEALTH_OK)"
else
  warn "$(__ HEALTH_TIMEOUT) ${MAX_WAIT}s."
  warn "$(__ HEALTH_TIP)"
  info "  ${COMPOSE_CMD} logs -f"
fi

# ---------------------------------------------------------------------------
# Final Summary
# ---------------------------------------------------------------------------
section "$(__ DEPLOY_DONE)"

SERVER_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
if [[ -z "$SERVER_IP" ]]; then
  SERVER_IP="your-server-ip"
fi

ACCESS_URL="http://${SERVER_IP}"
if [[ "$NGINX_PORT" != "80" ]]; then
  ACCESS_URL="http://${SERVER_IP}:${NGINX_PORT}"
fi

printf '\n'
info "$(__ BLOG):  ${ACCESS_URL}"
info "$(__ ADMIN): ${ACCESS_URL}/admin/"
printf '\n'
info "$(__ CREDS_SAVED)"
info "  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}"
info "  AUTH_SECRET:       ${AUTH_SECRET}"
printf '\n'
info "$(__ USEFUL_CMDS)"
info "  ${COMPOSE_CMD} logs -f           # $(__ VIEW_LOGS)"
info "  ${COMPOSE_CMD} ps                # $(__ CHECK_STATUS)"
info "  ${COMPOSE_CMD} down              # $(__ STOP_SERVICES)"
printf '\n'
info "$(__ UPGRADE_LATER)"
info "  bash <(curl -fsSL ${CONFIG_BASE_URL}/deploy/install.sh)"
printf '\n'
info "$(__ DOCS): https://github.com/grtsinry43/grtblog"
ok "$(__ DONE)"
