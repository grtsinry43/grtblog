# GrtBlog v2

**现代化博客系统** — 静态先行、实时注水、联合社交

[![Docker Publish](https://github.com/grtsinry43/grtblog-v2/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/grtsinry43/grtblog-v2/actions/workflows/docker-publish.yml)

GrtBlog 是一个面向创作者的博客系统，以纯静态 HTML 分发实现极致首屏速度，通过 WebSocket 实现毫秒级实时更新，并内置联合社交协议让博客不再是孤岛。

## 特性

- **极速加载** — 页面以纯静态 HTML 分发，首屏 < 0.5s，Go 后端宕机时 Nginx 仍可降级只读服务
- **实时更新** — WebSocket 驱动的内容热更新，修改文章后读者无需刷新即可看到最新内容
- **联合社交** — 自有联合协议 + ActivityPub 兼容，Mastodon 等 Fediverse 平台可关注你的博客
- **丰富内容** — 文章、手记 (Moments)、思考 (Thinking)、友链、时间线，满足多种表达需求
- **管理后台** — 美观且功能完备的 Vue 3 后台，Markdown 实时预览、评论管理、数据统计
- **一键部署** — Docker Compose 一键启动，多架构镜像 (amd64/arm64) 自动发布到 GHCR

## 架构

```
                      ┌──────────┐
                      │  用户/CDN │
                      └────┬─────┘
                           │
                      ┌────▼─────┐
                      │  Nginx   │  静态文件优先，回退到 SSR
                      └────┬─────┘
                           │
             ┌─────────────┼─────────────┐
             │             │             │
       ┌─────▼─────┐ ┌────▼────┐ ┌──────▼──────┐
       │ 静态 HTML  │ │ Go API  │ │  Admin SPA  │
       │           │ │  :8080  │ │  (Vue 3)    │
       └───────────┘ └────┬────┘ └─────────────┘
                          │
             ┌────────────┼────────────┐
             │            │            │
       ┌─────▼─────┐ ┌───▼───┐ ┌──────▼──────┐
       │ PostgreSQL │ │ Redis │ │  SvelteKit  │
       │            │ │       │ │  Renderer   │
       └────────────┘ └───────┘ └─────────────┘
```

**三个平面：**

| 平面 | 组件 | 职责 |
|------|------|------|
| 控制平面 | Go (Fiber) | API、ISR 调度、WebSocket Hub、联合协议、认证鉴权 |
| 渲染平面 | SvelteKit | SSR 渲染工厂，由 Go 后端驱动生成静态 HTML |
| 数据平面 | Nginx | 静态文件分发、反向代理、降级只读网关 |

**ISR (Incremental Static Regeneration)：** 内容变更时，Go 后端计算受影响路径 → 请求 SvelteKit 渲染 → 原子写入静态文件 → WebSocket 广播实时更新。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.24+, Fiber, GORM, Goose, Casbin, JWT |
| 前台 | SvelteKit, Svelte 5 (Runes), Tailwind CSS v4, TanStack Query |
| 后台 | Vue 3.5, Naive UI, Tailwind CSS, Pinia, Vite |
| 数据库 | PostgreSQL 17 |
| 缓存 | Redis 7 |
| 部署 | Docker Compose, Nginx, GitHub Actions, GHCR |

## 快速开始

### 使用预构建镜像部署（推荐）

```bash
# 创建部署目录
mkdir -p grtblog && cd grtblog

# 下载部署配置
BASE_URL="https://raw.githubusercontent.com/grtsinry43/grtblog-v2/main"
curl -fsSL "$BASE_URL/deploy/docker-compose.yml" -o docker-compose.yml
curl -fsSL "$BASE_URL/deploy/.env.example"       -o .env
mkdir -p nginx
curl -fsSL "$BASE_URL/deploy/nginx/nginx.conf"    -o nginx/nginx.conf

# 编辑 .env：设置 IMAGE_REPO_PREFIX、APP_VERSION、密码和密钥
#   IMAGE_REPO_PREFIX=ghcr.io/grtsinry43/
#   APP_VERSION=1.0.0              # 查看 Releases 页面获取最新版本
#   POSTGRES_PASSWORD=<强密码>
#   AUTH_SECRET=<openssl rand -hex 32>

# 启动
mkdir -p storage/html storage/uploads storage/geoip
docker compose up -d
```

首次启动会自动拉取镜像、运行数据库迁移。

- 博客首页: `http://your-server-ip`
- 管理后台: `http://your-server-ip/admin/`

### 本地构建部署

```bash
git clone https://github.com/grtsinry43/grtblog-v2.git
cd grtblog-v2/deploy
cp .env.example .env
# 编辑 .env：设置密码和密钥（IMAGE_REPO_PREFIX 留空）

mkdir -p storage/html storage/uploads storage/geoip
docker compose up -d --build
```

详细部署说明见 [部署文档](docs/guide/deployment.md)。

## 升级

```bash
# 修改 .env 中的 APP_VERSION，然后：
docker compose pull server renderer
docker compose up -d server renderer
```

Nginx 无需重启，自动发现新容器。

## 本地开发

```bash
# 1. 后端
cd server && cp .env.example .env && make migrate-up && make run   # :8080

# 2. 管理后台
cd admin && pnpm i && pnpm dev   # :5799

# 3. 前台
cd web && pnpm i && pnpm dev     # :5173
```

详细开发说明见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 项目结构

```
grtblog-v2/
├── server/         # Go 后端（控制平面）
├── web/            # SvelteKit 前台（渲染平面）
├── admin/          # Vue 3 管理后台
├── shared/         # 前端共享代码（Markdown 组件等）
├── deploy/         # Docker Compose 部署配置
├── scripts/        # 工具脚本（发布、迁移等）
└── docs/           # 文档（VitePress）
```

## 文档

| 文档 | 说明 |
|------|------|
| [项目介绍](docs/guide/introduction.md) | 核心特性与定位 |
| [快速部署](docs/guide/deployment.md) | 部署步骤与配置 |
| [写作指南](docs/guide/writing.md) | 内容创作与管理 |
| [个性化配置](docs/guide/configuration.md) | 站点设置 |
| [架构总览](docs/dev/architecture.md) | 系统设计与 ISR 机制 |
| [后端架构](docs/dev/backend.md) | Go 服务端 DDD 架构 |
| [前端架构](docs/dev/frontend.md) | SvelteKit 前台设计 |
| [管理后台](docs/dev/admin.md) | Vue 3 Admin 开发 |

## 数据迁移（v1 -> v2）

已提供 API 迁移脚本：`scripts/migrate-v1-to-v2.mjs`
使用说明见：`scripts/migrate-v1-to-v2.md`

## 致谢

本项目包含第三方开源软件，详见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

## License

[MIT](LICENSE)
