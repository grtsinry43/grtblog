# GrtBlog v2 版本、通道与更新策略

本文档定义 GrtBlog v2 的版本规则、发布通道、更新检查策略、数据库迁移兼容策略和回滚策略。

## 1. 版本体系

采用 **SemVer**：`MAJOR.MINOR.PATCH`。

- `MAJOR`：不兼容变更（API 破坏、协议破坏、关键行为破坏）。
- `MINOR`：向后兼容的新功能（新 API、新后台能力、新联邦能力）。
- `PATCH`：向后兼容的修复与性能优化。

预发布标识：

- Alpha：`2.0.0-alpha.N`
- Beta：`2.0.0-beta.N`
- RC：`2.0.0-rc.N`

构建元信息（可选）：

- `2.0.0-alpha.4+20260216.shaabcdef0`

## 2. 发布通道

对外只暴露两个通道：

- `stable`
- `preview`

内部阶段仍然保留：

- `alpha`
- `beta`
- `rc`

规则：

- `stable`：发 GitHub Release，同时推送 GHCR 和 Docker Hub
- `preview`：只推送 GHCR，只打 Git tag，不创建 GitHub Release
- `preview` 内部可以包含 `alpha/beta/rc`，但用户配置层面只选择 `stable` 或 `preview`

镜像标签策略：

- Stable 精确标签：`2.1.3`
- Stable 滚动标签：`stable`、`latest`、`2.1`
- Preview 精确标签：`2.2.0-beta.2`、`2.2.0-rc.1`
- Preview 滚动标签：`preview`
- Preview 阶段标签：`beta`、`rc`（未来需要时再加 `alpha`）

生产环境默认建议固定到不可变标签，不要把滚动标签作为默认部署方式。

## 3. 多组件版本约定

仓库中包含 `server`、`renderer(web)`、`admin` 三个组件，采用“**单仓统一发布号 + 镜像分组件标签**”策略：

- Git Tag（统一）：`v2.1.0-beta.2`
- Server 镜像：`ghcr.io/grtsinry43/grtblog-server:2.1.0-beta.2`
- Renderer 镜像：`ghcr.io/grtsinry43/grtblog-renderer:2.1.0-beta.2`
- Admin：随 `server` 镜像一起发布，不单独打包成生产部署单元

## 4. API 与协议兼容策略

后端公共 API 当前前缀为 `/api/v2`。

- 在 `v2.x` 生命周期中，`/api/v2` 保持兼容。
- 新字段：只追加，不复用旧字段语义。
- 字段废弃流程：`deprecate -> soft-remove(隐藏文档) -> major-remove`。

联邦协议（ActivityPub）建议维护独立 `protocol_version`（已在接口中暴露），其提升规则：

- 兼容扩展：`protocol minor` 递增。
- 不兼容变更：`protocol major` 递增，并至少跨一个 `MINOR` 版本提供兼容窗口。

## 5. 数据库迁移策略（关键）

采用“**Expand / Migrate / Contract**”三阶段，避免发布时硬中断：

1. Expand：先加新表/新列/新索引，不删除旧结构。
2. Migrate：应用新版本读写双轨（兼容新旧结构）。
3. Contract：确认所有实例完成升级后，再删旧列/旧约束（仅在下一次 MINOR 或 MAJOR）。

强约束：

- `PATCH` 发布不做破坏性 DDL。
- 破坏性 DDL 仅允许在 `MINOR`（且必须提前公告）或 `MAJOR`。
- 任何迁移都必须可重复执行（幂等）并提供回滚脚本或降级步骤说明。

## 6. 更新检查策略

更新检查的唯一有效通道配置应放在服务端部署环境：

- `APP_UPDATE_CHANNEL=stable|preview`

约束：

- `stable` 通道：后端读取 GitHub Releases，目标版本为最新稳定版
- `preview` 通道：后端读取 Git tags，目标版本为“当前 major 内最新预发布版本”
- `preview` 默认不跨 major 自动提示，例如 `2.x` 实例不会自动提示 `3.0.0-alpha.1`
- Admin 不直接访问 GitHub，只展示后端返回的更新结果

更新体验：

- 后台展示当前版本、目标版本、通道、来源和变更说明
- 默认给出“修改 `.env` -> `docker compose pull` -> `docker compose up -d`”的升级方式
- 自托管场景不追求一键更新，优先保证可理解、可回滚

## 7. 发布节奏（建议）

当前阶段建议节奏：

- Alpha：每周 1-2 次，允许功能快速迭代。
- Beta：每 2 周 1 次，收敛接口与行为。
- Stable：每月 1 次 `MINOR`，按需发布 `PATCH`。

进入下一阶段门槛：

- `Alpha -> Beta`：核心链路稳定（发布、ISR、实时 WS、备份恢复演练通过）。
- `Beta -> Stable`：回归测试稳定 + 至少 2 个发布周期无 P1 事故。

## 8. 回滚策略

回滚单位分两层：

- 应用层：`server/renderer` 镜像版本回退到上一个可用标签。
- 数据层：优先前滚修复，不建议直接回滚破坏性迁移。

建议执行顺序：

1. 停止写入敏感操作（维护模式或后台限流）。
2. 回滚 `server/renderer` 到上一个稳定镜像标签。
3. 若涉及 DB 不兼容：执行预先准备的兼容脚本，而不是直接 `down migration`。
4. 验证 `/health/liveness`、关键 API、首页/文章页、WS 通道。

## 9. Compose 环境版本锁定

`deploy/docker-compose.yml` 建议：

- Postgres 锁定大版本（例如 `17-alpine`）。
- Redis 锁定大版本（例如 `7-alpine`）。
- 业务镜像固定精确版本标签（例如 `2.0.0-beta.2`）。

只有在验证通过后，才推进 `alpha/beta/stable` 滚动标签。

## 10. GHCR 清理策略

GHCR 需要定期清理，但清理范围应只针对 preview 版本：

- `stable` 版本默认长期保留
- `preview` 版本按发布线保留最近 N 个（例如每条 `2.2.0-beta.*` 保留最近 5 个）
- 永远保留当前滚动标签引用的版本：`preview`、`beta`、`rc`

建议每周运行一次定时清理任务。

## 11. Git 与发布清单

建议 release checklist：

1. 合并到发布分支并打 tag：`vX.Y.Z[-pre.N]`
2. 生成并推送 `server/renderer` 镜像（精确标签 + 通道标签）
3. 执行 migration（Expand/Migrate）
4. 发布 Compose（或 Helm）配置
5. 运行 smoke test
6. 发布 changelog（包含 Breaking Changes / Migration Notes）

## 12. 当前推荐默认值

对绝大多数自托管实例，推荐默认值：

- `APP_VERSION=<精确版本号>`
- `APP_UPDATE_CHANNEL=stable`

只有测试环境或愿意承担更高变更频率的实例，再使用：

- `APP_VERSION=<preview 精确版本号>` 或 `APP_VERSION=preview`
- `APP_UPDATE_CHANNEL=preview`
