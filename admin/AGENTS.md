# AGENTS.md — GrtBlog v2 Admin（Vue 3 + Vite + Pinia）工程规范

> 目标：让 `admin/` 的代码组织、组件边界和验证流程可持续演进，减少上帝组件、横向耦合和散装业务逻辑。

## 0. 项目定位

- 后台管理端
- Vue 3 + Composition API + `<script setup lang="ts">`
- 页面编排优先，业务下沉，逐步向 DDD 风味靠拢

## 1. 当前仓库的核心原则

- `src/views/**`：路由页面与页面编排层。允许保留页面入口，但不要继续把完整业务堆进页面文件。
- `src/layout/**`：应用壳、导航、标签页、布局容器。不要混入具体业务域逻辑。
- `src/stores/**`：跨 feature 的全局状态，如用户、偏好、标签页、实时连接。禁止把单页表单状态塞进 Pinia。
- `src/components/**`：跨域复用的纯 UI/交互组件。只有在两个以上业务域共用时才允许放这里。
- `src/composables/**`：跨域共享能力。只有跨 feature 的逻辑才允许进入这里。
- `src/services/**`：现有基础设施目录。视为“历史兼容层”，不要继续把新业务逻辑无脑堆到这里。

## 2. 目标架构（DDD 风味，渐进迁移）

- `src/app/**`
  - 应用启动、路由注册、layout provider、主题注入
- `src/domains/<domain>/**`
  - 纯领域模型、类型、映射、领域规则
  - 不依赖 Vue 组件
- `src/features/<feature>/**`
  - 业务用例主场
  - 推荐结构：
    - `pages/`：feature 入口页
    - `components/`：feature 局部组件
    - `composables/`：feature 局部逻辑
    - `api/`：请求封装与协议适配
    - `model/`：feature 内部类型、selectors、转换函数
- `src/shared/**`
  - 真正跨 feature 的能力
  - 推荐结构：
    - `ui/`
    - `composables/`
    - `lib/`
    - `types/`
    - `config/`

## 3. 过渡期约束（非常重要）

- 在正式迁到 `src/features/**` 之前，新增代码优先按 feature 就近放在 `src/views/<feature>/**` 下。
- 页面文件只做组合，不做大段业务实现。
- `src/services/**` 不再作为默认落点：
  - 新增业务请求优先放到 feature 自己的 `api/`
  - 只有被多个 feature 共享的基础请求适配，才放到共享层
- `src/components/**` 不再放 feature 私有组件：
  - feature 私有组件放 `src/views/<feature>/components/` 或未来的 `src/features/<feature>/components/`
- `src/composables/**` 不再放 feature 私有逻辑：
  - feature 私有逻辑放 `src/views/<feature>/composables/` 或未来的 `src/features/<feature>/composables/`

## 4. 页面只做编排

- 页面组件负责：
  - 读取路由参数
  - 组装 feature 组件
  - 连接少量页面级 UI 状态
- 页面组件不负责：
  - API 协议细节
  - 多段业务规则拼接
  - 复杂副作用编排
  - 多块抽屉/弹窗/列表/表单同时内嵌实现

出现以下任一情况就应拆分：

- 一个页面同时处理表单、预览、弹窗、列表、统计、网络副作用
- 一个文件内存在 3 个以上相对独立的 UI 区块
- 两个页面之间出现明显复制结构
- 一个 composable 或组件需要“顺手知道”太多上层上下文

## 5. Vue 组件边界

- 默认使用 Composition API，不使用 Options API。
- 默认使用 `<script setup lang="ts">`。
- 组件通信遵循：
  - Props Down
  - Events Up
- 子组件禁止直接修改父状态，除非通过 `defineModel` 建立明确双向契约。
- 跨 3 层以上的共享上下文，才考虑 `provide/inject`。
- 子组件应尽量是：
  - 一个明确职责
  - 一个清晰输入输出面
  - 一个稳定的复用范围

## 6. Reactivity 规则

- 最小化源状态，优先 `computed` 派生。
- `watch` 只做副作用，不做纯数据派生。
- 避免模板里写复杂过滤、排序、拼装逻辑。
- 对外暴露的 composable，优先显式 action，而不是把整个可变状态裸露出去。
- 不要从 `reactive()` 直接解构丢失响应式。

## 7. Composable 规则

- composable 只做一类能力，不做“大杂烩控制器”。
- 跨 feature 才进入共享 composable。
- feature 私有逻辑必须留在 feature 内。
- 纯函数工具放 `utils/lib`，不要伪装成 composable。
- composable 入参超过 3 个可选项时，优先改成 options object。

## 8. 组件目录规则

- 共享 UI：`src/components/**`
- feature 局部 UI：`src/views/<feature>/components/**`
- feature 局部逻辑：`src/views/<feature>/composables/**`
- 共享编辑器/表格/上传等横切能力：
  - 若已被多个 feature 使用，可放 `src/views/shared/**`
  - 后续迁移时优先进入 `src/shared/**` 或 `src/features/<capability>/**`

## 9. 服务与领域边界

- `services/*.ts` 只负责：
  - HTTP 请求
  - WebSocket 接入
  - 第三方接口适配
- `services/*.ts` 不应负责：
  - 页面 UI 状态
  - Vue 生命周期
  - 大量视图专属转换
- 领域相关类型、映射和 selector，优先收敛到 feature/model 或 future domain 层。

## 10. Store 使用规则

- Pinia 只放跨 feature 状态：
  - 认证
  - 偏好
  - tabs
  - realtime
  - health
- 单个编辑页的表单、抽屉、预览、局部过滤器等状态必须留在页面/feature composable。
- 不允许为了省 props/emits 而把局部状态抬进全局 store。

## 11. 共享能力准入标准

一个能力只有满足以下条件之一，才允许进入共享层：

- 已被两个以上 feature 使用
- 明确属于应用级基础设施
- 明确属于设计系统级 UI

否则一律留在 feature 内部。

## 12. 代码风格与 SFC 约束

- SFC 顺序固定：
  - `<script setup>`
  - `<template>`
  - `<style scoped>`
- 样式默认 `scoped`
- 优先类选择器，不写泄漏式 element selector
- 使用 PascalCase 组件名
- 避免在模板中内联大块对象和复杂表达式

## 13. 后台端建议的演进路线

- 第一阶段：
  - 继续保留 `src/views/**`
  - 但所有新增逻辑必须按 feature 就近收纳
- 第二阶段：
  - 新 feature 直接落 `src/features/**`
  - `src/views/**` 退化为路由壳
- 第三阶段：
  - 收缩 `src/services/**`、`src/components/**`、`src/composables/**`
  - 让业务逻辑主要停留在 `features + domains + shared`

## 14. 验证要求

在 `admin/` 内完成改动后，默认至少执行：

- `pnpm lint:check`
- `pnpm type-check`
- `pnpm build`

如果未来引入 `oxlint`：

- 新增 `pnpm lint:ox`
- 默认同时跑 `pnpm lint:check` 与 `pnpm lint:ox`

未执行验证时，必须明确说明。

## 15. 当前硬性禁止事项

- 不继续制造上帝组件
- 不把 feature 私有组件放进全局 `components`
- 不把 feature 私有逻辑放进全局 `composables`
- 不让页面直接承担完整业务子系统
- 不把局部表单状态塞进 Pinia
- 不在 `watch` 里维护本应由 `computed` 派生的数据
- 不在模板里写复杂业务逻辑
- 不凭感觉下结论，必须基于代码事实

## 16. 给 Agent 的直接指令

- 修改前先判断代码应该属于：
  - app
  - domain
  - feature
  - shared
- 如果新增文件，优先考虑 feature 就近放置，而不是全局目录。
- 如果改动页面文件时发现页面已经过重，优先拆子组件或 composable，再继续功能改动。
- 如果发现两个页面存在 60% 以上结构重复，优先抽共享能力。
- 回答或提交变更时，说明：
  - 放在这个目录的原因
  - 是否符合 feature/shared 边界
  - 运行了哪些验证
