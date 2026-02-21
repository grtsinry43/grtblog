# svmarkdown — Markdown 渲染

`svmarkdown` 是本项目作者编写的 Svelte 运行时 Markdown 渲染库。与传统 "解析为 HTML 字符串再挂载" 不同，它将 Markdown 解析为自定义 AST，然后以声明式的 Svelte 组件树渲染。

- npm: `svmarkdown`
- 版本: `^0.1.3`
- 解析器: `markdown-it`

## 核心理念

```
Markdown 文本 → markdown-it 解析 → 自定义 AST (SvmdRoot) → Svelte 组件树渲染
```

每种 Markdown 元素（标题、段落、代码块、链接等）都可以映射到自定义 Svelte 组件，实现完全可控的渲染。

## 核心 API

```ts
import { Markdown, createParser, parseMarkdown, SvmdChildren } from 'svmarkdown'
```

| API | 说明 |
|-----|------|
| `parseMarkdown(markdown, options)` | 单次解析，返回 `SvmdRoot` AST |
| `createParser(options)` | 创建可复用的解析器（频繁更新场景） |
| `<Markdown />` | 一体化组件：接收 `content`，内部解析 + 渲染 |
| `<SvmdChildren />` | 渲染 AST 节点数组（配合 `parseMarkdown` 使用） |

## 自定义组件语法

### Container 语法 (`:::`)

```md
::: Alert type=warning title="注意"
这里是 **Markdown 内容**，会作为 `children` 传入组件。
:::
```

### Fence 语法 (`` ``` ``)

````md
```component:Chart {"title":"流量统计"}
month,visits
Jan,421
Feb,530
```
````

## 项目中的架构

### 整体结构

```
src/lib/
├── shared/markdown/
│   ├── svmarkdown.ts         # 解析/渲染配置中心
│   ├── MarkdownView.svelte   # 统一渲染入口组件
│   └── shared/components.ts  # 自定义组件块定义（Admin/Web 共用）
└── ui/markdown/
    ├── MarkdownHeading.svelte
    ├── MarkdownParagraph.svelte
    ├── MarkdownCodeBlock.svelte
    ├── MarkdownLink.svelte
    ├── MarkdownImage.svelte
    ├── MarkdownList.svelte
    ├── MarkdownListItem.svelte
    ├── MarkdownBlockquote.svelte
    ├── MarkdownTable.svelte
    ├── MarkdownHr.svelte
    ├── MarkdownFallback.svelte
    ├── YearCard.svelte
    ├── LinkCard.svelte
    └── FootnoteLinkCard.svelte
```

### 配置中心（`svmarkdown.ts`）

`src/lib/shared/markdown/svmarkdown.ts` 是整个 Markdown 渲染的配置核心，导出三个关键对象：

**组件映射表 `markdownComponents`**

将每种 Markdown 元素映射到对应的 Svelte 组件：

```ts
export const markdownComponents: SvmdComponentMap = {
  h1: MarkdownHeading,
  h2: MarkdownHeading,
  // ... h3-h6 同上
  p: MarkdownParagraph,
  ul: MarkdownList,
  ol: MarkdownList,
  li: MarkdownListItem,
  blockquote: MarkdownBlockquote,
  hr: MarkdownHr,
  table: MarkdownTable,
  // ... thead, tbody, tr, th, td
  a: MarkdownLink,
  img: MarkdownImage,
  code: MarkdownCodeBlock,
  // 自定义组件块
  'year-card': YearCard,
  'link-card': LinkCard,
  'footnote-link-card': FootnoteLinkCard,
  // 未实现的组件使用 Fallback
  gallery: MarkdownFallback,
  callout: MarkdownFallback,
  timeline: MarkdownFallback,
};
```

**解析配置 `markdownParseOptions`**

```ts
export const markdownParseOptions: SvmdParseOptions = {
  componentBlocks,          // 从 shared/components.ts 动态生成
  markdownItPlugins: [],
  markdownItOptions: {
    html: true,             // 允许 HTML 标签
    linkify: true,          // 自动识别链接
    typographer: true       // 排版优化（智能引号等）
  }
};
```

**渲染配置 `markdownRenderOptions`**

```ts
export const markdownRenderOptions: SvmdRenderOptions = {
  allowDangerousHtml: true  // 允许渲染原始 HTML
};
```

### 渲染入口（`MarkdownView.svelte`）

`MarkdownView` 是项目中所有 Markdown 内容的统一渲染入口：

```svelte
<script lang="ts">
  import { parseMarkdown, SvmdChildren } from 'svmarkdown';

  const { content, headingAnchors, components, parseOptions, renderOptions } = $props();

  const nodes = $derived.by(() => {
    const ast = parseMarkdown(content ?? '', parseOptions);
    return applyHeadingAnchors(ast.children, headingAnchors);
  });
</script>

<div class="markdown-preview">
  <SvmdChildren {nodes} {components} {renderOptions} />
</div>
```

它使用 `parseMarkdown` + `SvmdChildren` 的组合方式（而非 `<Markdown />` 组件），这样可以在渲染前对 AST 进行额外处理——如将服务端生成的 `headingAnchors` 注入到标题节点的 `id` 属性中，用于 TOC 锚点定位。

### 自定义组件编写

每个 Markdown UI 组件通过 `$props()` 接收节点信息。以 `MarkdownHeading` 为例：

```svelte
<script lang="ts">
  import type { SvmdElementNode } from 'svmarkdown';

  const { node, children } = $props<{
    node: SvmdElementNode;
    children?: Snippet;
  }>();

  // 根据 node.name (h1-h6) 决定样式
  const tag = node.name;
</script>

<svelte:element this={tag} id={node.attrs?.id} class="...">
  {@render children?.()}
</svelte:element>
```

可用的 props：

| Prop | 类型 | 说明 |
|------|------|------|
| `node` | `SvmdElementNode` / `SvmdComponentNode` | 当前 AST 节点，包含 `name`、`attrs`、`children` |
| `children` | `Snippet` | Svelte 5 子内容片段 |
| `syntax` | `string` | 组件块语法来源（`container` / `fence`） |
| `source` | `string` | 原始 Markdown 源码 |

### 自定义组件块

项目中通过 `shared/components.ts` 定义了三种自定义组件块：

| 组件名 | 用途 | Markdown 语法 |
|--------|------|---------------|
| `year-card` | 年份归档卡片 | `::: year-card year="2024"` |
| `link-card` | 链接卡片 | `::: link-card url="..." title="..."` |
| `footnote-link-card` | 脚注链接卡片 | `::: footnote-link-card` |

这些定义同时被 Web 前端和 Admin 后台共用。

## 扩展指南

### 添加新的 Markdown 组件

1. 在 `src/lib/ui/markdown/` 下创建新组件
2. 在 `src/lib/shared/markdown/svmarkdown.ts` 的 `markdownComponents` 中注册
3. 如果是自定义组件块，还需在 `shared/components.ts` 中添加定义

### 添加 markdown-it 插件

在 `markdownParseOptions.markdownItPlugins` 数组中添加：

```ts
import footnote from 'markdown-it-footnote';

export const markdownParseOptions: SvmdParseOptions = {
  // ...
  markdownItPlugins: [footnote],
};
```

### 渲染配置调整

```ts
// 禁用危险 HTML（更安全但限制功能）
export const markdownRenderOptions: SvmdRenderOptions = {
  allowDangerousHtml: false
};
```