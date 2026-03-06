# 架构总览

## 为什么从 v1 到 v2？

GrtBlog v1 基于 **Spring Boot (Java 17) + Next.js 15 + Umi.js 4** 构建，另有 Python 推荐服务、Elasticsearch、MongoDB 等多个组件。在一年多的迭代中，系统积累了以下问题：

### v1 的痛点

| 问题 | 具体表现 |
|------|---------|
| **架构复杂** | Java 后端 + Next.js 前端 + Umi.js 后台 + Python 推荐服务，四个独立技术栈 |
| **数据库过多** | MySQL + MongoDB + Redis + Elasticsearch + MeiliSearch，五个存储引擎各司其职但运维成本极高 |
| **部署门槛高** | Docker Compose 需要 6+ 个容器，配置繁琐，甚至阻碍了作者自己后续维护 |
| **仓库膨胀** | Git 历史混入大量二进制资源，仓库体积快速膨胀 |
| **边界模糊** | 设计系统、内容模型与插件机制（PF4J）的职责逐渐交叉 |
| **BFF 废弃** | 规划的 BFF 层未能落地，停留在空目录 |

### v2 的设计决策

v2 不是简单重写，而是基于 v1 所有经验的架构重新设计：

| 决策 | v1 做法 | v2 做法 | 理由 |
|------|---------|---------|------|
| 后端语言 | Java (Spring Boot) | **Go (Fiber)** | 编译为单二进制，内存占用从数百 MB 降至数十 MB |
| 前端框架 | Next.js (React) | **SvelteKit (Svelte 5)** | 更小的 bundle、更少的运行时开销、Runes 语法更直觉 |
| 管理后台 | Umi.js (React) | **Vue 3 (Naive UI)** | 轻量且与前台技术栈解耦，避免 React 全家桶 |
| 数据库 | MySQL + MongoDB | **PostgreSQL 一个搞定** | JSONB 覆盖文档型需求，减少运维复杂度 |
| 搜索 | Elasticsearch + MeiliSearch | **后端内建** | 博客体量下内建搜索足够，去掉两个重型依赖 |
| 推荐系统 | 独立 Python 微服务 | **Go 内建** | 减少跨语言通信和部署复杂度 |
| 静态生成 | Next.js ISR (框架内建) | **自研 ISR (Go 驱动)** | Go 后端直接调度渲染、原子写入，完全可控 |
| 实时通信 | Socket.io + Netty | **原生 WebSocket** | 去掉 Socket.io 协议层开销 |
| 部署 | 6+ 容器 | **3 容器** (Go + SvelteKit + Nginx + DB) | 大幅降低部署门槛 |

### 核心原则

1. **更简单** — 单体结构、少依赖、少魔法。核心功能先跑通，复杂能力按阶段演进
2. **更轻量** — 默认 SSG，按需 SSR / API。资产与静态资源彻底解耦，仓库长期保持干净
3. **更可持续** — 用工程化方式给自己一个可长期维护的内容平台

## 注水静态架构 (Rehydrated Static Architecture)

GrtBlog v2 的核心架构在静态站点的极致性能和动态应用的实时交互之间取得平衡：

1. **静态先行 (Static First)** - 所有公开页面默认为纯静态 HTML，由 Nginx 直接分发，实现极致首屏速度与 0 CPU 占用
2. **增量生成 (Incremental Generation)** - 仅在内容变更时，由 Go 控制平面驱动 SvelteKit 渲染器生成受影响的页面
3. **实时注水 (Realtime Rehydration)** - 客户端通过 WebSocket 实现评论、点赞及内容的毫秒级热更新
4. **联合社交 (Union Social)** - 自有联合协议为核心，兼容 ActivityPub 协议

## 系统拓扑

```
                          ┌──────────┐
                          │  用户/CDN │
                          └────┬─────┘
                               │
                          ┌────▼─────┐
                          │  Nginx   │
                          │ Gateway  │
                          └────┬─────┘
                               │
                 ┌─────────────┼─────────────┐
                 │             │             │
           ┌─────▼─────┐ ┌────▼────┐ ┌──────▼──────┐
           │ 静态文件   │ │ Go API  │ │ Admin SPA   │
           │ /var/www/  │ │  :8080  │ │ (Vue 3)     │
           │   html     │ └────┬────┘ └─────────────┘
           └────────────┘      │
                 ┌─────────────┼─────────────┐
                 │             │             │
           ┌─────▼─────┐ ┌────▼────┐ ┌──────▼──────┐
           │ PostgreSQL │ │ Redis   │ │ SvelteKit   │
           │            │ │         │ │ Renderer    │
           └────────────┘ └─────────┘ │   :3000     │
                                      └─────────────┘
```

## 三个平面

### 控制平面 (Control Plane) - Go Backend

系统的中枢神经，负责：

- 数据持久化与 API 服务
- ISR 调度：计算脏路径、触发页面再生成
- WebSocket Hub：维护实时连接与房间订阅
- **联合协议**：自有联合协议实现，内容分发与订阅
- **ActivityPub 兼容层**：HTTP Signatures 验签、Inbox 消息处理
- 认证与安全 (JWT)

**技术栈**: Go 1.24+, Fiber, GORM, DDD 分层架构

### 渲染平面 (Render Plane) - SvelteKit

一个"无头"的渲染工厂，仅对内服务：

- Go 后端通过内部 HTTP 调用请求渲染
- SvelteKit 执行 `load` 函数，走内网访问 Go API 获取数据
- 返回完整 HTML 字符串，由 Go 原子写入磁盘

**技术栈**: SvelteKit, Bun/Node, Tailwind CSS, Svelte 5 Runes

### 数据平面 (Data Plane) - Nginx Gateway

面向用户的网关层：

```nginx
location / {
    # 静态文件优先 -> 目录索引 -> 回源后端
    try_files $uri $uri.html $uri/index.html @backend;
}
```

当 Go 后端宕机时，Nginx 仍可服务静态文件，实现「降级只读」。

## 路由分发

| 路径 | 目标 | 说明 |
|------|------|------|
| `/api/*` | Go Server | REST API |
| `/api/v2/ws/*` | Go Server | WebSocket |
| `/uploads/*` | Go Server | 上传文件 |
| `/admin/*` | Admin SPA | Vue 3 管理后台 |
| 其他 | Nginx try_files | 静态优先，回退到 SvelteKit SSR |

## ISR 工作流

ISR（Incremental Static Regeneration）是本项目的核心机制，类似 Next.js 的 ISR，但完全自研：

```
Admin 发布文章
  │
  ▼
Go 写入数据库
  │
  ▼
DirtyPathCalculator 计算受影响路径
  例: /posts/new, /index, /tags/Go, /feed.xml
  │
  ▼
RenderQueue 异步任务入队
  │
  ▼
Worker 请求 SvelteKit Renderer
  GET http://renderer:3000/posts/new
  │
  ▼
AtomicWriter 原子写入静态文件
  TempFile -> Rename (防并发读写白屏)
  │
  ▼
WebSocket Hub 广播 post_created 事件
  │
  ▼
在线用户收到实时通知
```

## 实时更新流

```
Admin 修改文章错别字
  │
  ▼
Go 更新 DB + 广播 WS post_update (带 payload)
  │
  ▼
在线阅读用户的 Svelte Store 收到 payload
  │
  ▼
无感替换 DOM 文本节点（无需刷新）
  │
  ▼
Go 异步触发静态文件重新生成（为后来者服务）
```

## 联合与社交协议

### 自有联合协议（核心）

GrtBlog 定义了自有的联合协议，用于 GrtBlog 实例之间以及与支持该协议的平台之间的内容分发、订阅和互动。这是 GrtBlog 社交能力的核心基础。

### ActivityPub 兼容（兼容层）

为了让博客能融入更广泛的 Fediverse 生态，GrtBlog 同时兼容 ActivityPub 协议。Mastodon、Misskey 等平台的用户可以通过 ActivityPub 关注博客、接收更新。

::: info 设计定位
联合协议是 GrtBlog 的「一等公民」，ActivityPub 是「兼容适配器」。两者在后端有独立的模块（`union` / `federation`），在管理后台也分别独立配置。
:::
