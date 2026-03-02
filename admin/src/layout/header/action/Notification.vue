<script setup lang="ts">
import { useQuery, useQueryClient } from '@tanstack/vue-query'
import { NBadge, NButton, NEmpty, NList, NListItem, NPopover, NText, NThing, useNotification } from 'naive-ui'
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { adminNotificationService, type AdminNotificationResp } from '@/services/admin-notifications'
import { toRefsUserStore } from '@/stores'

const router = useRouter()
const queryClient = useQueryClient()
const notification = useNotification()
const { token } = toRefsUserStore()

const ws = ref<WebSocket | null>(null)
const isPopoverShow = ref(false)
const reconnectTimer = ref<number | null>(null)
const isUnmounted = ref(false)
const reconnectAttempts = ref(0)
const seenNotifIds = ref(new Set<number>())
const pausedByVisibility = ref(false)

const { data: unreadData } = useQuery({
  queryKey: ['admin-notifications', 'unread'],
  queryFn: () => adminNotificationService.listMine(true, 1, 5),
  refetchOnWindowFocus: true,
})

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v2').replace(/\/$/, '')

const buildWsUrl = () => {
  const url = new URL(API_BASE_URL, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/ws/notifications`
  url.search = ''
  return url.toString()
}

const clearReconnectTimer = () => {
  if (reconnectTimer.value != null) {
    window.clearTimeout(reconnectTimer.value)
    reconnectTimer.value = null
  }
}

const scheduleReconnect = () => {
  if (isUnmounted.value) return
  clearReconnectTimer()
  const delay = Math.min(1000 * 2 ** reconnectAttempts.value, 10000)
  reconnectAttempts.value += 1
  reconnectTimer.value = window.setTimeout(() => {
    void connectWs()
  }, delay)
}

const connectWs = async () => {
  if (isUnmounted.value) return
  const jwt = token.value?.trim()
  if (!jwt) {
    scheduleReconnect()
    return
  }
  const wsUrl = buildWsUrl()

  if (ws.value) {
    ws.value.close(1000, 'refresh')
  }

  ws.value = new WebSocket(wsUrl, ['grtblog.jwt', jwt])

  ws.value.onopen = () => {
    reconnectAttempts.value = 0
    clearReconnectTimer()
  }

  ws.value.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (!data || typeof data !== 'object' || (!data.title && !data.content)) {
        return
      }
      const id = typeof data.id === 'number' ? data.id : 0
      if (id > 0 && seenNotifIds.value.has(id)) return
      if (id > 0) seenNotifIds.value.add(id)
      queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
      notification.create({
        title: data.title || '收到新通知',
        content: data.content,
        duration: 5000,
      })
    } catch (e) {
      console.error('WS message parse error', e)
    }
  }

  ws.value.onerror = () => {}

  ws.value.onclose = (event) => {
    ws.value = null
    if (!isUnmounted.value && !pausedByVisibility.value && event.code !== 1000) scheduleReconnect()
  }
}

const handleVisibilityChange = () => {
  if (isUnmounted.value) return
  if (document.hidden) {
    pausedByVisibility.value = true
    clearReconnectTimer()
    if (ws.value) {
      ws.value.close(1000, 'visibility')
      ws.value = null
    }
  } else {
    if (pausedByVisibility.value) {
      pausedByVisibility.value = false
      reconnectAttempts.value = 0
      void connectWs()
    }
  }
}

onMounted(() => {
  document.addEventListener('visibilitychange', handleVisibilityChange)
  if (!document.hidden) {
    void connectWs()
  } else {
    pausedByVisibility.value = true
  }
})

onUnmounted(() => {
  isUnmounted.value = true
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  clearReconnectTimer()
  ws.value?.close()
})

const handleMarkRead = async (id: number) => {
  try {
    await adminNotificationService.markRead(id)
    queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
  } catch {
    // ignore
  }
}

const handleMarkReadAll = async () => {
    try {
        await adminNotificationService.markAllRead()
        queryClient.invalidateQueries({ queryKey: ['admin-notifications'] })
    } catch {
        // ignore
    }
}

const handleViewAll = () => {
  isPopoverShow.value = false
  router.push({ name: 'adminNotificationList' })
}

const handleNotificationClick = (item: AdminNotificationResp) => {
    if (!item.is_read) {
        handleMarkRead(item.id)
    }
    // Logic to navigate or show details based on payload if needed
}



</script>

<template>
  <NPopover v-model:show="isPopoverShow" trigger="click" placement="bottom-end" :width="350">
    <template #trigger>
      <NBadge :value="unreadData?.total || 0" :max="99">
        <NButton quaternary circle>
             <span class="icon-[ph--bell] text-xl" />
        </NButton>
      </NBadge>
    </template>
    <div class="flex flex-col">
      <div class="flex items-center justify-between px-4 py-2">
        <NText strong>未读通知</NText>
        <NButton text type="primary" size="small" @click="handleMarkReadAll" v-if="(unreadData?.total ?? 0) > 0">
           全部已读
        </NButton>
      </div>
      <div class="h-[400px]">
        <ScrollContainer wrapper-class="!p-0">
          <NList hoverable clickable v-if="unreadData?.items && unreadData.items.length > 0">
            <NListItem 
              v-for="item in unreadData.items" 
              :key="item.id" 
              @click="handleNotificationClick(item)"
              :class="{ 'bg-blue-50/50 dark:bg-blue-900/10': !item.is_read }"
            >
              <NThing :title="item.title">
                <template #description>
                  <div class="line-clamp-2 text-xs text-naive-text-2">
                    {{ item.content }}
                  </div>
                </template>
                <template #footer>
                   <div class="text-[10px] text-naive-text-3">
                      {{ new Date(item.created_at).toLocaleString() }}
                   </div>
                </template>
              </NThing>
            </NListItem>
          </NList>
          <div v-else class="py-8 text-center">
              <NEmpty description="暂无未读通知" />
          </div>
        </ScrollContainer>
      </div>
      <div class="p-2">
        <NButton block secondary @click="handleViewAll">
          查看全部
        </NButton>
      </div>
    </div>
  </NPopover>
</template>
