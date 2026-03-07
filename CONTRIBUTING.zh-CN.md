[English](./CONTRIBUTING.md)

# 贡献指南

感谢你对 **grtblog-v2** 的关注。本文档涵盖开发环境搭建、架构概览、关键约定与贡献流程。

---

## 技术栈

本项目以 Monorepo 形式组织，包含以下三个主要模块：

| 模块 | 路径 | 技术栈 | 说明 |
|------|------|--------|------|
| **Server** | `/server` | Go 1.24+, Fiber, GORM, PostgreSQL | 后端 API，采用 DDD 架构；负责数据管理与静态页面生成。 |
| **Admin** | `/admin` | Vue 3, Naive UI, Vite | 内容管理后台，提供 Markdown 编辑能力。 |
| **Web** | `/web` | SvelteKit, Tailwind CSS | 面向读者的前端。生产环境使用 SSR，静态快照由后端生成。 |

---

## 开发环境搭建

### 前置要求

- Go 1.24+
- Node.js 20+
- pnpm 9+
- PostgreSQL

### 1. 启动后端 (Server)

后端是整个系统的数据源，运行在端口 **:8080**。

```bash
cd server
cp .env.example .env   # 配置数据库连接信息
make migrate-up         # 执行数据库迁移
make run
```

### 2. 启动后台管理 (Admin)

后台管理面板运行在端口 **:5799**。

```bash
cd admin
pnpm install && pnpm dev
```

### 3. 启动前台 (Web)

- **开发模式**（端口 **:5173**）：支持热更新。
- **生产 SSR 模式**（端口 **:3000**）：用于测试 ISR 静态生成流程。

```bash
cd web
pnpm install && pnpm dev
```

---

## ISR（增量静态生成）测试

本项目实现了一套自定义 ISR 机制：Go 后端抓取 SSR 服务（端口 :3000）的页面，并将渲染结果持久化为静态 HTML 文件。

如果你修改了 **SvelteKit 路由逻辑**或**后端生成代码**，务必验证 ISR 流程。

### 运行 ISR 预览

ISR 预览脚本需要 Admin Token 进行 API 认证。请先在后台管理面板创建令牌（**设置 > 管理员令牌**），然后执行：

```bash
# 确保后端已在 :8080 运行
export PREVIEW_ISR_TOKEN="gt_你的管理员令牌"
make preview-isr
```

脚本将依次执行以下操作：

1. 清空 HTML 存储目录。
2. 构建 Web 前端。
3. 将静态资源复制到 `server/storage/html/`。
4. 在端口 :3000 启动 SSR 服务。
5. 调用后端 API 生成静态 HTML 快照。
6. 关闭 SSR 服务，并在端口 **:5555** 启动静态文件服务器。

> 如果 :5555 上的所有页面都能正常渲染，则说明 ISR 流程运行正常。

---

## 开发规范

### 后端：同步与异步生成

- **API 处理器 (`RefreshPostsHTML`)**：必须为**同步**执行。调用方阻塞直到生成完成并收到 200 响应，以此通知脚本何时关闭 SSR 服务。
- **事件订阅器 (`ArticleUpdated`)**：必须为**异步**执行。文章保存时通过 EventBus 触发，不得阻塞保存接口。

### 前端：URL 尾部斜杠

SvelteKit 配置了 `trailingSlash: 'always'`。所有生成的 URL 必须以 `/` 结尾（如 `/posts/hello/`），以匹配文件系统中 `posts/hello/index.html` 的目录结构。否则静态部署后刷新页面会返回 404。

### 共享 Markdown 组件

自定义 Markdown 组件（相册、卡片等）定义在 `shared/markdown/components.ts` 中，Admin 和 Web 共用此代码。新增或修改组件时，需确保两个模块均已适配。

---

## 提交规范

本项目遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

| 前缀 | 用途 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复缺陷 |
| `docs` | 文档变更 |
| `refactor` | 代码重构（不改变行为） |
| `chore` | 构建、CI、依赖更新等 |

示例：`feat(server): 增加文章字数统计功能`

---

## 版本发布

项目采用**统一版本号**，所有模块共享同一版本。

- **格式**：`vMAJOR.MINOR.PATCH`（如 `v2.1.0`）
  - `MAJOR`：不兼容的变更（API 或数据库迁移破坏兼容性）
  - `MINOR`：向后兼容的新功能
  - `PATCH`：向后兼容的修复与小幅优化
- **预发布版本**：使用 SemVer 预发布标记（如 `v2.1.0-beta.1`、`v2.1.0-rc.1`）
- **中间构建版本**：使用 commit hash 标识（如 `8f3c1a2b9d4e`）

### 创建发布

```bash
# 本地创建发布标签
./scripts/release.sh v2.1.0

# 创建并推送发布标签
./scripts/release.sh v2.1.0 --push
```

发布脚本将自动执行：

1. 校验版本号格式。
2. 在 `docs/releases/<version>.md` 生成发布说明草稿。
3. 创建带注释的 Git 标签。

### 镜像标签

Docker 镜像标签必须与发布版本号一致：

- `grtblog-server:v2.1.0`
- `grtblog-renderer:v2.1.0`

部署时通过 `deploy/.env` 中的 `APP_VERSION` 变量统一控制镜像版本。
