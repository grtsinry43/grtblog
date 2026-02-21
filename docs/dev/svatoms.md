# svatoms — 模型数据流

`svatoms` 是本项目作者编写的 Svelte 模型数据流库，提供 "一次提供，处处消费" 的数据分发模式。API 理念来源于 React 生态的 Jotai "provider + selector"，但使用 Svelte 原生的 `context + stores` 实现。

- npm: `svatoms`
- 版本: `^0.2.0`

## 解决什么问题

SvelteKit 项目中常见的痛点：

- `load()` 数据需要在多个层级使用，prop drilling 快速膨胀
- 许多组件只需要模型的极少字段，但传递整个对象会触发多余更新
- 业务数据分散在各处，缺少集中管理

`svatoms` 将这些问题收敛为三步：

1. **单一入口**：在 `+page.svelte` 或 `+layout.svelte` 调用 `mountModelData` 挂载数据
2. **最小切片**：子组件通过 `selectModelData` 只订阅需要的字段
3. **集中更新**：通过 `useModelActions` 获取更新方法，一处修改全局响应

## 核心 API

```ts
import { createModelDataContext } from 'svatoms'
```

### `createModelDataContext<Model>(options?)`

创建一个指定模型类型的 context 管理器。

| 参数 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | 内部 symbol 标签（调试用） |
| `key` | `symbol` | 自定义 context key（高级） |
| `initial` | `Model \| null` | 初始值 |
| `defaultScope` | `'local' \| 'global'` | 默认作用域 |

返回的对象包含以下方法：

| 方法 | 说明 |
|------|------|
| `mountModelData(dataOrSource, opts?)` | 挂载数据，组件销毁时自动重置为 `null`。支持静态值、getter、`Readable` store |
| `selectModelData(selector, opts?)` | 创建派生 store，只订阅 selector 返回的切片 |
| `useModelActions()` | 在组件初始化时绑定 `setModelData`、`updateModelData`、`getModelData`（事件回调中安全使用） |
| `setModelData(value)` | 直接设置模型数据 |
| `updateModelData(fn)` | 函数式更新模型数据 |
| `getModelData()` | 获取当前模型数据快照 |

## 项目中的用法

### 1. 定义 Context（`context.ts`）

每个业务域模块在 `context.ts` 中创建自己的 context：

```ts
// src/lib/features/post/context.ts
import { createModelDataContext } from 'svatoms';
import type { PostDetail } from '$lib/features/post/types';

export const postDetailCtx = createModelDataContext<PostDetail | null>({
  name: 'postDetailCtx',
  initial: null
});
```

项目中已有的 context 实例：

| Context | 模块 | 模型类型 |
|---------|------|----------|
| `postListCtx` | `features/post` | 文章列表 + 分页 |
| `postDetailCtx` | `features/post` | 文章详情 |
| `commentAreaCtx` | `features/comment` | 评论区状态 |
| `AuthCtx` | `features/auth` | 认证状态 |
| `momentListCtx` | `features/moment` | 手记列表 |
| `momentDetailCtx` | `features/moment` | 手记详情 |
| `thinkingListCtx` | `features/thinking` | 思考列表 |
| `websiteInfoCtx` | `features/website-info` | 站点信息 |
| `detailPanelCtx` | `shared/detail-panel` | 侧边详情面板 |
| `imageExtInfoCtx` | `shared/markdown` | 图片扩展信息 |

### 2. 页面挂载数据（`+page.svelte`）

在页面组件顶层调用 `mountModelData`，传入 getter 使数据随 SvelteKit 导航自动同步：

```svelte
<!-- src/routes/posts/[slug]/+page.svelte -->
<script lang="ts">
  import { postDetailCtx } from '$lib/features/post/context';
  let { data } = $props();

  // 传 getter 确保客户端导航时 data 变化后 context 自动同步
  postDetailCtx.mountModelData(() => data.post ?? null);
</script>
```

布局级挂载同理（如站点信息在根布局挂载）：

```svelte
<!-- src/routes/+layout.svelte -->
<script lang="ts">
  websiteInfoCtx.mountModelData(() => data.websiteInfo ?? null);
</script>
```

### 3. 子组件消费切片（`selectModelData`）

任意子组件通过 selector 只订阅需要的字段：

```svelte
<script lang="ts">
  import { websiteInfoCtx } from '$lib/features/website-info/context';

  const websiteName = websiteInfoCtx.selectModelData(
    (data) => data?.website_name || 'grtBlog'
  );
</script>

<h1>{$websiteName}</h1>
```

### 4. 更新数据（`useModelActions`）

在事件回调中更新数据必须先通过 `useModelActions()` 绑定：

```svelte
<script lang="ts">
  import { commentAreaCtx } from '$lib/features/comment/context';

  // 组件初始化时绑定（必须在顶层）
  const { updateModelData } = commentAreaCtx.useModelActions();

  const setReply = (comment: CommentNode) => {
    updateModelData((prev) => ({ ...prev, replyingTo: comment }));
  };
</script>
```

> **注意**：`getContext()` 只能在组件初始化时调用，因此不能在事件回调中直接使用 `setModelData`，必须提前通过 `useModelActions()` 绑定。

## 选择器等价判断

当 selector 返回对象/数组时，默认的 `Object.is` 比较会导致每次 model 变动都触发重渲染。此时需提供自定义 `equals`：

```ts
const store = detailPanelCtx.selectModelData(
  (model) => ({ toc: model?.toc ?? [], title: model?.title ?? '' }),
  {
    equals: (a, b) => a.title === b.title && a.toc === b.toc
  }
);
```

**规则**：selector 返回原始值时可省略 `equals`，返回新对象/数组时必须提供。

## 作用域

- **local**（默认）：数据作用域为调用 `mountModelData` 的组件树，SSR 安全
- **global**：整个应用共享，类似单例 store

```ts
// 显式指定 global 作用域
websiteInfoCtx.mountModelData(() => data, { scope: 'global' });
```