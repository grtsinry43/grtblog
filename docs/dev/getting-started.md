# 本地开发

## 项目结构

```
grtblog-v2/
├── server/         # Go Fiber 后端 (控制平面)
├── web/            # SvelteKit 前端 (渲染平面)
├── admin/          # Vue 3 管理后台
├── shared/         # 前端共享代码 (Markdown 组件等)
├── deploy/         # Docker Compose 部署配置
├── scripts/        # 工具脚本
└── docs/           # 文档 (VitePress)
```

## 环境准备

- **Go** 1.24+
- **Node.js** 20+ (推荐使用 pnpm 作为包管理器)
- **PostgreSQL** 15+ (开发时也可使用 SQLite)
- **Redis** (可选，开发时非必需)

## 启动开发环境

三个服务独立运行，互不依赖构建，但数据流向为：Server -> Web -> 静态文件。

### 1. 启动后端

```bash
cd server
cp .env.example .env   # 首次需要配置
make migrate-up         # 首次需要执行数据库迁移
make run                # 启动 :8080
```

### 2. 启动管理后台

```bash
cd admin
pnpm i && pnpm dev      # 启动 :5799
```

### 3. 启动前台

```bash
cd web
pnpm i && pnpm dev      # 启动 :5173 (开发热更新)
```

## 测试 ISR 流程

ISR 是本项目最核心的机制。如果你改了 SvelteKit 路由逻辑或后端生成代码，务必测试：

```bash
# 确保后端 (:8080) 在运行，然后在根目录执行
make preview-isr
```

这个命令会自动：
1. 编译 Web 前端
2. 后台启动 SSR 服务 (:3000)
3. 调用后端 API 执行一次 ISR bootstrap
4. 启动静态文件服务器 (:5555) 并打开浏览器

::: warning 关键判断标准
`:5555` 端口的页面正常显示 = ISR 机制正常工作。
:::

## 调试小贴士

### 后端 API 文档

启动后端后，访问 Swagger/Scalar 文档：
```
http://localhost:8080/docs
```

### 日志

后端日志中搜索 `isr`、`render`、`html` 关键字可快速定位静态生成相关问题。

### 内部通信

- Web (renderer) 通过 `INTERNAL_API_BASE_URL` 访问 Go API（生产环境为内网地址）
- Go 通过 `HTMLSNAPSHOT_BASE_URL` 请求 SvelteKit 渲染页面
