<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { debounce } from 'lodash-es'
import {
  NButton,
  NCard,
  NEmpty,
  NList,
  NListItem,
  NPagination,
  NSpace,
  NSwitch,
  NTag,
  NThing,
  useMessage,
} from 'naive-ui'
import { ref } from 'vue'

import { ScrollContainer } from '@/components'
import { adminNotificationService } from '@/services/admin-notifications'

import type { AdminNotificationResp } from '@/services/admin-notifications'

const queryClient = useQueryClient()
const message = useMessage()

const page = ref(1)
const pageSize = ref(20)
const unreadOnly = ref(false)

const { data } = useQuery({
  queryKey: ['admin-notifications', 'list', page, pageSize, unreadOnly],
  queryFn: () => adminNotificationService.listMine(unreadOnly.value, page.value, pageSize.value),
  placeholderData: (previousData) => previousData,
})

const markReadMutation = useMutation({
  mutationFn: (id: number) => adminNotificationService.markRead(id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
  },
})

const markAllReadMutation = useMutation({
  mutationFn: () => adminNotificationService.markAllRead(),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
    message.success('已全部标记为已读')
  },
})

const handleMarkRead = (item: AdminNotificationResp) => {
  if (!item.is_read) {
    markReadMutation.mutate(item.id)
  }
}

// Debounce the mark read action to prevent accidental reads while scrolling
const debouncedMarkRead = debounce((id: number) => {
  markReadMutation.mutate(id)
}, 500)

const handleMouseEnter = (item: AdminNotificationResp) => {
  if (!item.is_read) {
    debouncedMarkRead(item.id)
  }
}

const handlePageChange = (p: number) => {
  page.value = p
}

const handleMarkAllRead = () => {
  markAllReadMutation.mutate()
}
</script>

<template>
  <ScrollContainer
    wrapper-class="p-4"
    :scrollbar-props="{ trigger: 'none' }"
  >
    <NCard
      title="通知中心"
      :bordered="false"
    >
      <template #header-extra>
        <NSpace align="center">
          <span class="text-sm">仅看未读</span>
          <NSwitch v-model:value="unreadOnly" />
          <NButton @click="handleMarkAllRead">全部已读</NButton>
        </NSpace>
      </template>

      <NList
        hoverable
        clickable
      >
        <template v-if="data?.items?.length">
          <NListItem
            v-for="item in data.items"
            :key="item.id"
            @click="handleMarkRead(item)"
            @mouseenter="handleMouseEnter(item)"
            :class="{ 'bg-blue-50/50 dark:bg-blue-900/10': !item.is_read }"
          >
            <NThing>
              <template #header>
                <NSpace align="center">
                  <NTag
                    v-if="!item.is_read"
                    type="success"
                    size="small"
                    round
                    >NEW</NTag
                  >
                  <span :class="{ 'font-bold': !item.is_read }">{{ item.title }}</span>
                </NSpace>
              </template>
              <template #description>
                <div class="text-naive-text-2 mt-1">{{ item.content }}</div>
              </template>
              <template #footer>
                <div class="text-naive-text-3 mt-2 text-xs">
                  {{ new Date(item.created_at).toLocaleString() }}
                </div>
              </template>
            </NThing>
          </NListItem>
        </template>
        <template v-else>
          <NEmpty description="暂无通知" />
        </template>
      </NList>

      <div
        class="mt-4 flex justify-end"
        v-if="data?.total"
      >
        <NPagination
          v-model:page="page"
          :page-size="pageSize"
          :item-count="data.total"
          @update:page="handlePageChange"
        />
      </div>
    </NCard>
  </ScrollContainer>
</template>
