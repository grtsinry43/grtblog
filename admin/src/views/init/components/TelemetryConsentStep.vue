<script setup lang="ts">
import { NSwitch } from 'naive-ui'

defineProps<{
  primaryColorRgb?: string
}>()

const enabled = defineModel<boolean>('enabled', { default: false })
</script>

<template>
  <div class="space-y-5">
    <div
      class="rounded-lg border border-neutral-100 p-5 dark:border-neutral-800"
      :style="
        enabled && primaryColorRgb
          ? {
              borderColor: `rgba(${primaryColorRgb}, 0.35)`,
              background: `rgba(${primaryColorRgb}, 0.04)`,
            }
          : undefined
      "
    >
      <div class="flex items-start justify-between gap-4">
        <div class="flex items-start gap-3">
          <span
            class="mt-0.5 iconify text-xl ph--heartbeat"
            :style="primaryColorRgb ? { color: `rgb(${primaryColorRgb})` } : undefined"
          ></span>
          <div>
            <div class="text-sm font-medium text-neutral-800 dark:text-neutral-100">匿名遥测</div>
            <p class="mt-1 text-xs leading-relaxed text-neutral-500">
              完全可选。开启后只会发送脱敏后的运行数据，帮助我们更快发现并修复问题。
            </p>
          </div>
        </div>
        <NSwitch v-model:value="enabled" />
      </div>
    </div>

    <div class="space-y-4 text-xs leading-relaxed text-neutral-500">
      <div>
        <div
          class="mb-1.5 text-[11px] font-semibold tracking-wide text-neutral-700 uppercase dark:text-neutral-200"
        >
          我们会收集
        </div>
        <ul class="list-disc space-y-1 pl-4">
          <li>脱敏后的错误摘要与 Panic 指纹</li>
          <li>基础运行指标（版本、部署模式、请求量、延迟等）</li>
          <li>功能开关状态与内容计数（文章数、评论数等）</li>
        </ul>
      </div>

      <div>
        <div
          class="mb-1.5 text-[11px] font-semibold tracking-wide text-neutral-700 uppercase dark:text-neutral-200"
        >
          我们不会收集
        </div>
        <ul class="list-disc space-y-1 pl-4">
          <li>任何个人信息或账号凭据</li>
          <li>文章、手记、评论等正文内容</li>
          <li>访客身份、浏览轨迹或 IP</li>
        </ul>
      </div>

      <p>
        可随时在「设置 → 帮助我们变得更好」中预览将要上报的完整数据，或关闭此功能。GrtBlog
        是开源项目，遥测相关代码均可在 GitHub 上查看与审计。
      </p>
    </div>
  </div>
</template>
