# 贡献指南

## Git 提交规范

遵循 Angular Commit Convention：

```
feat(scope): 新功能描述
fix(scope): Bug 修复描述
docs(scope): 文档变更
refactor(scope): 重构
chore(scope): 构建/依赖等杂项
```

示例：`feat(server): 增加文章字数统计功能`

scope 取值：`server`、`web`、`admin`、`docs`、`deploy`

## 开发注意事项

### 后端

- 新增 API 时遵循 DDD 分层：domain -> app -> http
- 数据库变更必须通过 Goose 迁移文件，不要手动改表
- API Handler 中不要放业务逻辑，下沉到 app service
- 统一错误码格式 `APP-4xxx`

### 前端 (Web)

- 页面组件只做编排，业务逻辑放 `lib/features/`
- 首屏数据走 `+page.server.ts`，客户端数据走 TanStack Query
- 不在 SSR 阶段访问 `window` / `document` / `navigator`
- 浏览器 API 封装为 Svelte actions
- URL 保持尾部斜杠（`trailingSlash: 'always'`）

### 管理后台 (Admin)

- 遵循 Lithe Admin 的组件和布局约定
- API 调用统一通过 `services/` 层

### 共享代码

- `shared/markdown/components.ts` 是 Admin 和 Web 共用的
- 修改时确保两端都能正常渲染

## 测试

改动涉及 ISR 流程时，务必运行：

```bash
make preview-isr
```

确认 `:5555` 端口页面正常显示。

## 代码审查

提交 PR 前请确认：

1. 代码通过 lint 检查
2. 不引入新的 `window` / `document` 直接访问
3. ISR 相关改动已通过 `preview-isr` 测试
4. 数据库变更有对应的迁移文件
