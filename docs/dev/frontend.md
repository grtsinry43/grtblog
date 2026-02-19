# 前端架构 (SvelteKit Web)

## 技术栈

- **SvelteKit** + **Svelte 5** (Runes 语法)
- **Tailwind CSS v4** (@theme 配置)
- **TanStack Query** (客户端异步数据)
- **adapter-node** (SSR 模式，供 Go 后端爬取)

## 目录结构

```
web/src/
├── routes/             # 页面路由 (SvelteKit 文件路由)
│   ├── +layout.svelte  # 根布局
│   ├── +page.svelte    # 首页
│   ├── posts/          # 文章
│   ├── thinkings/      # 思考
│   ├── moments/        # 手记
│   ├── friends/        # 友链
│   ├── timeline/       # 时间线
│   ├── tags/           # 标签
│   ├── auth/           # 认证
│   └── statistics/     # 统计
├── lib/
│   ├── features/       # 业务域模块
│   │   └── <feature>/
│   │       ├── api.ts
│   │       ├── types.ts
│   │       ├── context.ts
│   │       └── components/
│   ├── shared/         # 共享能力
│   │   ├── clients/    # API 客户端
│   │   ├── markdown/   # Markdown 渲染
│   │   ├── theme/      # 主题系统
│   │   └── actions/    # Svelte actions
│   ├── ui/             # 通用 UI 组件
│   └── assets/         # 静态资源
└── hooks.server.ts     # 服务端钩子
```

## 核心原则

### 页面只做编排

页面组件 (`+page.svelte`) 只负责"拼装"，业务逻辑下沉到 `lib/features/` 和 `lib/shared/`：

```
routes/posts/[slug]/+page.svelte   <- 编排组件
lib/features/post/api.ts            <- 数据请求
lib/features/post/types.ts          <- 类型定义
lib/features/post/components/       <- 业务组件
```

### 数据获取分层

**服务端数据（首屏关键内容）**：

- 通过 `+page.server.ts` 的 `load` 函数获取
- 必须透传 `fetch`：`getArticle(fetch, slug)`
- 确保内容可被 SSR/SSG，SEO 友好

**客户端数据（实时交互内容）**：

- 评论、点赞、实时更新等走 TanStack Query
- 通过 `QueryRoot` 组件统一挂载
- 客户端 fetch 传 `undefined`：`getComments(undefined, postId)`

### Svelte 5 Runes

项目使用 Svelte 5 Runes 语法：

- `$state` - 本地可变状态
- `$derived` - 派生状态（替代 reactive 声明）
- `$effect` - 副作用（DOM / 网络 / 订阅）
- `$props` - 组件 props

### URL 尾部斜杠

项目配置了 `trailingSlash: 'always'`，所有路由 URL 必须以 `/` 结尾：

```
正确: /posts/hello/
错误: /posts/hello
```

这确保静态文件结构 `posts/hello/index.html` 能正确匹配。

## 性能策略

- **文章 / 列表 / 友链 / 归档**：SSR 输出完整 HTML
- **评论 / 点赞 / TOC 高亮**：client-only islands
- **重型库**：仅在 `onMount` / 交互后动态 import
- **浏览器 API**：封装为 Svelte actions，统一处理 cleanup 和 SSR 安全

## 设计系统

全局主题定义在 `src/routes/layout.css`（Tailwind v4 @theme）：

- **色彩**：温暖灰 (Warm Gray) + Jade 主色
- **字体**：Sans / Serif / Mono 三套字体
- **样式优先使用 Tailwind 类名**；复用场景可用 `@apply`

## 共享 Markdown 组件

`shared/markdown/components.ts` 定义了自定义 Markdown 组件（相册、卡片等），**Admin 和 Web 共用**。修改时两端都需要适配。
