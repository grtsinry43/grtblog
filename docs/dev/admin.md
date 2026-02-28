# 管理后台 (Vue 3 Admin)

## 技术栈

- **Vue 3.5** + TypeScript
- **Naive UI** (组件库)
- **Tailwind CSS**
- **Pinia** (状态管理)
- **Vite** (构建工具)

基于 **Lithe Admin** 模板开发，提供了响应式布局和丰富的基础组件。

## 目录结构

```
admin/src/
├── views/              # 页面视图
│   ├── articles/       # 文章管理
│   ├── notes/          # 手记管理 (Moments)
│   ├── thinking/       # 思考管理
│   ├── comments/       # 评论管理
│   ├── friend-links/   # 友链管理
│   ├── navigation/     # 导航管理
│   ├── uploads/        # 文件上传
│   ├── union/          # 联合 (自有协议)
│   ├── federation/     # 联邦 (ActivityPub)
│   ├── sysconfig/      # 系统配置
│   ├── monitoring/     # 监控
│   ├── visitors/       # 访客管理
│   ├── email/          # 邮件配置
│   ├── webhooks/       # Webhook 管理
│   └── ...
├── components/         # 公共组件
│   ├── markdown-editor/
│   ├── template-editor/
│   └── html-editor/
├── composables/        # Vue Composables
├── services/           # API 服务层
├── stores/             # Pinia Stores
├── router/             # 路由配置
├── layout/             # 布局组件
└── utils/              # 工具函数
```

## 开发

```bash
cd admin
pnpm i
pnpm dev        # 开发服务器 :5799
```

### 环境变量

先复制模板：

```bash
cp admin/.env.example admin/.env.development
cp admin/.env.example admin/.env.production
```

- `.env.development` - 开发环境配置
- `.env.production` - 生产环境配置（建议将 `VITE_APP_BASE` 改为 `/admin/`）

当前默认项如下：

- `VITE_APP_BASE` - 管理后台部署基础路径（开发默认 `/`，生产默认 `/admin/`）
- `VITE_APP_NAME` - 站点名称（用于页脚、Logo、页面标题）
- `VITE_APP_TITLE` - 初始 HTML 标题
- `VITE_WATERMARK_CONTENT` - 水印文本
- `VITE_API_BASE_URL` - API 基础路径（默认 `/api/v2`）
- `VITE_API_PROXY_TARGET` - 开发代理目标（默认 `http://localhost:8080`）

生产构建时 `VITE_APP_BASE=/admin/`，确保部署在 `/admin/` 路径下。

## 构建

```bash
pnpm build      # 输出到 dist/
```

构建产物为纯静态 SPA，在 Docker 部署中由 Nginx 直接服务。
