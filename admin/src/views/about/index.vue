<script setup lang="ts">
import { NScrollbar, NTag } from 'naive-ui'
import { onMounted, ref } from 'vue'

import packageJson from '@/../package.json'
import { ScrollContainer } from '@/components'
import { getSystemStatus } from '@/services/system'

defineOptions({
  name: 'About',
})

const APP_NAME = import.meta.env.VITE_APP_NAME
const version = ref('')
const commit = ref('')
const { dependencies, devDependencies } = packageJson

getSystemStatus().then((res) => {
  version.value = res.app.version
  commit.value = res.app.commit ?? ''
}).catch(() => {
  version.value = __APP_VERSION__
})

let codeToHtml: any
const dependenciesCodeHighlight = ref('')
const devDependenciesCodeHighlight = ref('')

const frontendTech = [
  { name: 'Vue 3', icon: 'ph--vue-logo', color: '#42b883', desc: '渐进式 JavaScript 框架' },
  { name: 'Naive UI', icon: null, color: '#75B93F', desc: '企业级 Vue 3 组件库' },
  { name: 'Vite', icon: 'ph--lightning', color: '#9499ff', desc: '下一代前端构建工具' },
  { name: 'TailwindCSS 4', icon: 'ph--wind', color: '#00bcff', desc: '原子化 CSS 框架' },
  { name: 'TypeScript', icon: 'ph--file-ts', color: '#3178C6', desc: '类型安全的 JavaScript' },
  { name: 'Pinia', icon: 'ph--tree-structure', color: '#FFD859', desc: 'Vue 状态管理' },
]

const backendTech = [
  { name: 'Go', icon: 'ph--code', color: '#00ADD8', desc: '高性能后端语言' },
  { name: 'Fiber', icon: 'ph--rocket-launch', color: '#00ACD7', desc: 'Express 风格 Go 框架' },
  { name: 'PostgreSQL', icon: 'ph--database', color: '#4169E1', desc: '关系型数据库' },
  { name: 'Redis', icon: 'ph--hard-drives', color: '#DC382D', desc: '缓存与消息队列' },
  { name: 'SvelteKit', icon: 'ph--monitor', color: '#FF3E00', desc: '前台 SSR 渲染引擎' },
]

const features = [
  { icon: 'ph--article', title: 'Markdown 写作', desc: '组件块扩展：相册、提示框、时间轴、链接卡片' },
  { icon: 'ph--newspaper', title: '内容管理', desc: '文章、动态、思考、页面的完整生命周期管理' },
  { icon: 'ph--cloud-arrow-up', title: '媒体资源', desc: '图片与文件上传、预览、重命名与批量管理' },
  { icon: 'ph--shield-check', title: '安全与权限', desc: 'JWT + RBAC 权限控制、OAuth 绑定、登录限流' },
  { icon: 'ph--arrows-clockwise', title: '事件驱动更新', desc: '内容变更触发异步刷新，WebSocket 推送实时更新' },
  { icon: 'ph--chart-line-up', title: '数据分析', desc: '访客画像、行为漏斗、流量趋势、可观测性监控' },
]

onMounted(async () => {
  if (!codeToHtml) {
    // @ts-ignore
    const shiki = await import('https://cdn.jsdelivr.net/npm/shiki@3.7.0/+esm')
    codeToHtml = shiki.codeToHtml
  }

  codeToHtml(JSON.stringify(dependencies, null, 2), {
    lang: 'json',
    themes: {
      light: 'min-light',
      dark: 'dark-plus',
    },
  })
    .then((result: string) => (dependenciesCodeHighlight.value = result))
    .catch(() => (dependenciesCodeHighlight.value = JSON.stringify(dependencies, null, 2)))

  codeToHtml(JSON.stringify(devDependencies, null, 2), {
    lang: 'json',
    themes: {
      light: 'min-light',
      dark: 'dark-plus',
    },
  })
    .then((result: string) => (devDependenciesCodeHighlight.value = result))
    .catch(() => (devDependenciesCodeHighlight.value = JSON.stringify(devDependencies, null, 2)))
})
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4 pb-6">
    <!-- Section 1: Hero -->
    <div class="mt-4 mb-2">
      <div class="flex items-center gap-3">
        <h1 class="text-2xl font-semibold text-neutral-800 dark:text-neutral-100">
          Grtblog Admin
        </h1>
        <NTag size="small" round type="info">{{ version }}{{ commit ? ` (${commit})` : '' }}</NTag>
      </div>
      <p class="mt-1 text-sm font-medium text-neutral-500 dark:text-neutral-400">
        面向创作者与读者的全栈内容平台
      </p>
      <p class="mt-3 max-w-3xl text-sm leading-relaxed text-neutral-600 dark:text-neutral-400">
        grtblog-v2 是对 v1 的系统性重构：回到单体结构、减少依赖与复杂度，以默认 SSG 为主、按需引入 SSR / API。
        项目由 Go API、SvelteKit 前台、Vue 后台与共享 Markdown 组件能力组成，本后台为 Lithe Admin 的二次开发版本，
        专为内容管理、发布与运营流程定制。
      </p>
    </div>

    <!-- Section 2: Tech Stack -->
    <div>
      <h2 class="mb-3 text-base font-medium text-neutral-700 dark:text-neutral-200">前端技术</h2>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <div
          v-for="tech in frontendTech"
          :key="tech.name"
          class="flex items-center gap-3 rounded border border-naive-border bg-naive-card p-4 transition-[background-color,border-color]"
        >
          <div
            class="grid size-10 shrink-0 place-items-center rounded-full"
            :style="{ backgroundColor: tech.color + '18' }"
          >
            <span
              v-if="tech.icon"
              class="size-5 iconify"
              :class="tech.icon"
              :style="{ color: tech.color }"
            />
            <span
              v-else
              class="text-sm font-bold"
              :style="{ color: tech.color }"
            >N</span>
          </div>
          <div class="min-w-0">
            <div class="text-sm font-medium text-neutral-700 dark:text-neutral-200">{{ tech.name }}</div>
            <div class="text-xs text-neutral-500 dark:text-neutral-400">{{ tech.desc }}</div>
          </div>
        </div>
      </div>
    </div>

    <div>
      <h2 class="mb-3 text-base font-medium text-neutral-700 dark:text-neutral-200">后端技术</h2>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <div
          v-for="tech in backendTech"
          :key="tech.name"
          class="flex items-center gap-3 rounded border border-naive-border bg-naive-card p-4 transition-[background-color,border-color]"
        >
          <div
            class="grid size-10 shrink-0 place-items-center rounded-full"
            :style="{ backgroundColor: tech.color + '18' }"
          >
            <span
              class="size-5 iconify"
              :class="tech.icon"
              :style="{ color: tech.color }"
            />
          </div>
          <div class="min-w-0">
            <div class="text-sm font-medium text-neutral-700 dark:text-neutral-200">{{ tech.name }}</div>
            <div class="text-xs text-neutral-500 dark:text-neutral-400">{{ tech.desc }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 3: Features -->
    <div>
      <h2 class="mb-3 text-base font-medium text-neutral-700 dark:text-neutral-200">核心能力</h2>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="feat in features"
          :key="feat.title"
          class="flex gap-3 rounded border border-naive-border bg-naive-card p-4 transition-[background-color,border-color]"
        >
          <div class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/8">
            <span class="size-5 text-primary iconify" :class="feat.icon" />
          </div>
          <div class="min-w-0">
            <div class="text-sm font-medium text-neutral-700 dark:text-neutral-200">{{ feat.title }}</div>
            <div class="mt-0.5 text-xs leading-relaxed text-neutral-500 dark:text-neutral-400">{{ feat.desc }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 4: Architecture Overview -->
    <div>
      <h2 class="mb-3 text-base font-medium text-neutral-700 dark:text-neutral-200">架构概览</h2>
      <div class="rounded border border-naive-border bg-naive-card p-5 transition-[background-color,border-color]">
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <!-- API Layer -->
          <div class="flex flex-col items-center gap-2 rounded-lg bg-neutral-50 p-4 dark:bg-neutral-800/50">
            <span class="iconify ph--cloud size-6 text-sky-500" />
            <span class="text-sm font-medium text-neutral-700 dark:text-neutral-200">Go API 服务</span>
            <div class="flex flex-wrap justify-center gap-1.5">
              <NTag size="tiny" round>Fiber</NTag>
              <NTag size="tiny" round>GORM</NTag>
              <NTag size="tiny" round>JWT</NTag>
              <NTag size="tiny" round>WebSocket</NTag>
            </div>
            <div class="flex items-center gap-1 text-xs text-neutral-400">
              <span class="iconify ph--arrows-left-right size-3.5" />
              <span>PostgreSQL / Redis</span>
            </div>
          </div>

          <!-- SSR Layer -->
          <div class="flex flex-col items-center gap-2 rounded-lg bg-neutral-50 p-4 dark:bg-neutral-800/50">
            <span class="iconify ph--browser size-6 text-orange-500" />
            <span class="text-sm font-medium text-neutral-700 dark:text-neutral-200">SvelteKit 前台</span>
            <div class="flex flex-wrap justify-center gap-1.5">
              <NTag size="tiny" round>SSR 渲染</NTag>
              <NTag size="tiny" round>静态快照</NTag>
              <NTag size="tiny" round>内容哈希</NTag>
            </div>
            <div class="flex items-center gap-1 text-xs text-neutral-400">
              <span class="iconify ph--arrow-right size-3.5" />
              <span>HTML 快照发布</span>
            </div>
          </div>

          <!-- Admin Layer -->
          <div class="flex flex-col items-center gap-2 rounded-lg bg-neutral-50 p-4 dark:bg-neutral-800/50">
            <span class="iconify ph--layout size-6 text-emerald-500" />
            <span class="text-sm font-medium text-neutral-700 dark:text-neutral-200">Vue 3 后台</span>
            <div class="flex flex-wrap justify-center gap-1.5">
              <NTag size="tiny" round>Naive UI</NTag>
              <NTag size="tiny" round>Pinia</NTag>
              <NTag size="tiny" round>TailwindCSS</NTag>
            </div>
            <div class="flex items-center gap-1 text-xs text-neutral-400">
              <span class="iconify ph--arrows-left-right size-3.5" />
              <span>WebSocket 实时通信</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Section 5: Dependencies -->
    <div>
      <h2 class="mb-3 text-base font-medium text-neutral-700 dark:text-neutral-200">依赖信息</h2>
      <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
        <div class="rounded border border-naive-border bg-naive-card p-4 transition-[background-color,border-color]">
          <NTag class="mb-3" :bordered="false" type="info" size="small">dependencies</NTag>
          <NScrollbar style="max-height: 420px">
            <div v-html="dependenciesCodeHighlight"></div>
          </NScrollbar>
        </div>
        <div class="rounded border border-naive-border bg-naive-card p-4 transition-[background-color,border-color]">
          <NTag class="mb-3" :bordered="false" type="info" size="small">devDependencies</NTag>
          <NScrollbar style="max-height: 420px">
            <div v-html="devDependenciesCodeHighlight"></div>
          </NScrollbar>
        </div>
      </div>
    </div>
  </ScrollContainer>
</template>
