<script setup lang="ts">
import {
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NSpace,
  NTag,
  NText,
} from 'naive-ui'
import { formatDate } from '@/utils/format'

import type { VisitorProfile, VisitorRecentComment } from '@/types/visitors'

defineProps<{
  visible: boolean
  loading: boolean
  profile: VisitorProfile | null
  recentComments: VisitorRecentComment[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const statusTagTypeMap: Record<string, 'default' | 'info' | 'warning' | 'success' | 'error'> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
  blocked: 'default',
}
</script>

<template>
  <NDrawer :show="visible" width="760" @update:show="emit('update:visible', $event)">
    <NDrawerContent title="访客画像详情" :native-scrollbar="false">
      <div v-if="loading" class="py-8 text-center">
        <NText depth="3">加载中...</NText>
      </div>

      <template v-else-if="profile">
        <NDescriptions bordered label-placement="left" :column="2" class="mb-4">
          <NDescriptionsItem label="访客 ID"><code>{{ profile.visitorId }}</code></NDescriptionsItem>
          <NDescriptionsItem label="昵称">{{ profile.nickName || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="邮箱">{{ profile.email || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="网站">
            <a v-if="profile.website" :href="profile.website" target="_blank" class="text-primary hover:underline">{{ profile.website }}</a>
            <span v-else>-</span>
          </NDescriptionsItem>
          <NDescriptionsItem label="IP">{{ profile.ip || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="地区">{{ profile.location || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="浏览器 / 平台">{{ [profile.browser, profile.platform].filter(Boolean).join(' / ') || '-' }}</NDescriptionsItem>
          <NDescriptionsItem label="首次出现">{{ formatDate(profile.firstSeenAt) }}</NDescriptionsItem>
          <NDescriptionsItem label="最近活跃">{{ formatDate(profile.lastSeenAt) }}</NDescriptionsItem>
          <NDescriptionsItem label="最近浏览">{{ formatDate(profile.lastViewedAt) }}</NDescriptionsItem>
          <NDescriptionsItem label="最近点赞">{{ formatDate(profile.lastLikedAt) }}</NDescriptionsItem>
        </NDescriptions>

        <NSpace class="mb-4">
          <NTag type="info">浏览 {{ profile.totalViews }}</NTag>
          <NTag type="info">浏览内容数 {{ profile.uniqueViewItems }}</NTag>
          <NTag type="success">点赞 {{ profile.totalLikes }}</NTag>
          <NTag type="success">点赞内容数 {{ profile.uniqueLikedItems }}</NTag>
          <NTag type="warning">评论 {{ profile.totalComments }}</NTag>
        </NSpace>

        <NCard title="最近评论" size="small">
          <div v-if="recentComments.length === 0" class="text-center py-4 text-[var(--text-color-3)]">暂无评论记录</div>
          <NSpace v-else vertical :size="12">
            <div v-for="item in recentComments" :key="item.id" class="rounded border border-gray-200 p-3">
              <NSpace justify="space-between" align="center" class="mb-2">
                <NSpace align="center">
                  <NTag size="small" :type="statusTagTypeMap[item.status] || 'default'">{{ item.status }}</NTag>
                  <NTag v-if="item.isDeleted" size="small" type="error">已删除</NTag>
                </NSpace>
                <NText depth="3" style="font-size: 12px">{{ formatDate(item.createdAt) }}</NText>
              </NSpace>
              <div class="text-sm whitespace-pre-wrap break-all">{{ item.content }}</div>
            </div>
          </NSpace>
        </NCard>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
