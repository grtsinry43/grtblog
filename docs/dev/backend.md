# 后端架构 (Go Server)

## 技术栈

- **Go 1.24+** / Fiber v2 (HTTP 框架)
- **GORM** (ORM)
- **Goose** (数据库迁移)
- **Casbin** (RBAC 权限控制)
- **JWT** (认证)

## DDD 分层架构

```
server/internal/
├── domain/         # 领域层：实体、值对象、仓库接口
├── app/            # 应用层：用例编排、业务逻辑
├── http/           # 接口层：Handler、Router、Middleware
├── infra/          # 基础设施层：持久化、事件、外部服务
├── security/       # 安全：JWT、RBAC、Turnstile
└── ws/             # WebSocket Hub
```

### 领域模块

| 模块 | 说明 |
|------|------|
| `identity` | 用户与认证 |
| `content` | 文章内容 |
| `comment` | 评论系统 |
| `navigation` | 导航菜单 |
| `media` | 媒体文件 |
| `social` | 社交功能 |
| `like` | 点赞 |
| `federation` | 联邦 (ActivityPub 兼容层) |
| `thinking` | 思考 |

### 应用层补充模块

domain 层之外，app 层还有以下独立模块：

| 模块 | 说明 |
|------|------|
| `moment` | 手记 (Moments) |
| `htmlsnapshot` | HTML 快照生成 |
| `friendtimeline` | 友链时间线 |
| `hitokoto` | 一言 |
| `home` | 首页聚合 |
| `globalnotification` | 全局通知 |
| `adminnotification` | 管理员通知 |
| `adminstats` | 管理后台统计 |
| `setupstate` | 初始化状态 |
| `federationconfig` | 联邦配置 (ActivityPub) |
| `email` | 邮件通知 |
| `search` | 全文搜索 |
| `webhook` | Webhook 推送 |
| `config` | 系统配置 |

### 应用服务

应用层中值得关注的模块：

- **`isr/`** - ISR 核心逻辑：脏路径计算、渲染队列、原子写入
- **`event/`** - 事件总线，驱动异步任务（如文章更新后触发 ISR）
- **`analytics/`** - 访客统计与分析
- **`moment/`** - 手记业务逻辑
- **`federation/`** + **`federationconfig/`** - 联邦 (ActivityPub 兼容层) 实现与配置

## 关键设计

### AtomicWriter

静态文件写入采用 TempFile + Rename 机制，避免并发读写导致的白屏：

```
写入流程: HTML String -> TempFile -> os.Rename -> 完成
```

`os.Rename` 在同一文件系统上是原子操作，保证读者要么读到旧文件，要么读到新文件，不会读到半截内容。

### DirtyPathCalculator

更新内容时，智能计算所有需要重新生成的路径：

```
更新 Post(id=1, tag="Go")
  -> /posts/1       (文章详情页)
  -> /              (首页)
  -> /tags/Go       (标签页)
  -> /feed.xml      (RSS)
```

### 同步 vs 异步

这是一个容易混淆的点：

- **API Handler（如 `RefreshPostsHTML`）**：**同步**执行。脚本调用后等待生成完成再返回 200，确保调用方知道何时可以停止 SSR 进程
- **事件监听（如 `ArticleUpdated`）**：**异步**执行。通过 EventBus 触发，不阻塞保存接口

### WebSocket Hub

基于房间订阅模式：

- 客户端按文章 ID 订阅特定房间
- 支持的事件类型：`post_created`、`post_update`、`new_comment`、`new_mention` 等

## 数据库迁移

使用 Goose 管理 SQL 迁移文件，位于 `server/migrations/`。

```bash
cd server
make migrate-up      # 执行迁移
make migrate-down    # 回滚一步
make migrate-status  # 查看状态
```

## API 文档

后端集成了 Swagger/Scalar，启动后访问 `/docs` 查看完整的 API 文档。

API 遵循 RESTful 规范，统一错误码格式 `APP-4xxx`。
