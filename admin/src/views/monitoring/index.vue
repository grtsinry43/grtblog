<script setup lang="ts">
import { useIntervalFn } from '@vueuse/core'
import {
  NCard,
  NGrid,
  NGi,
  NProgress,
  NStatistic,
  NSpace,
  NTag,
  NEmpty,
  NSpin,
  NIcon,
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NNumberAnimation,
  NResult
} from 'naive-ui'
import { 
  Desktop24Regular, 
  ArrowClockwise24Regular,
} from '@vicons/fluent'
import { ref, onMounted, computed } from 'vue'

import { ScrollContainer } from '@/components'
import { getSystemStatus } from '@/services/system'

import type { SystemStatus } from '@/services/system'

defineOptions({
  name: 'SystemMonitor',
})

const loading = ref(true)
const status = ref<SystemStatus | null>(null)
const lastUpdated = ref<Date>(new Date())

const fetchData = async () => {
  try {
    const data = await getSystemStatus()
    status.value = data
    lastUpdated.value = new Date()
  } catch (err) {
    console.error('Failed to fetch system status:', err)
  } finally {
    loading.value = false
  }
}

// Auto-refresh every 10 seconds
useIntervalFn(fetchData, 10000)

onMounted(() => {
  fetchData()
})

const formatBytes = (bytes: number | string, decimals = 2) => {
  const b = typeof bytes === 'string' ? parseFloat(bytes) : bytes
  if (isNaN(b) || b === 0) return '0 Bytes'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return parseFloat((b / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

const memoryPercentage = computed(() => {
  if (!status.value) return 0
  return Math.round((status.value.memory.alloc / status.value.memory.sys) * 100)
})

const diskPercentage = computed(() => {
  if (!status.value) return 0
  if (status.value.disk.all === 0) return 0
  return Math.round((status.value.disk.used / status.value.disk.all) * 100)
})

const getStatusType = (usage: number) => {
  if (usage > 90) return 'error'
  if (usage > 75) return 'warning'
  return 'success'
}
</script>

<template>
  <ScrollContainer wrapper-class="p-4 md:p-6 space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
      <div>
        <h1 class="text-2xl font-light text-gray-900 dark:text-gray-100 flex items-center gap-3">
          <NIcon :component="Desktop24Regular" class="text-primary" />
          <span>系统状态</span>
        </h1>
        <p class="text-sm text-gray-500 mt-1 pl-1">
          实时监控应用运行各项指标
          <span class="text-xs opacity-75 ml-2">更新于 {{ lastUpdated.toLocaleTimeString() }}</span>
        </p>
      </div>
      <NButton :loading="loading" @click="fetchData" size="small" secondary round>
        <template #icon>
          <NIcon :component="ArrowClockwise24Regular" />
        </template>
        刷新
      </NButton>
    </div>

    <NSpin :show="loading && !status">
      <div v-if="status" class="space-y-6">
        
        <!-- Status Overview -->
        <NGrid :x-gap="16" :y-gap="16" cols="1 s:2 m:4" responsive="screen">
          <NGi>
            <NCard size="small" :bordered="false" class="h-full">
              <NStatistic label="运行时间">
                {{ status.app.uptime }}
              </NStatistic>
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" :bordered="false" class="h-full">
              <NStatistic label="Go 版本">
                {{ status.app.goVersion }}
              </NStatistic>
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" :bordered="false" class="h-full">
              <NStatistic label="CPU 核心">
                <NNumberAnimation :from="0" :to="status.cpu.cores" />
                <template #suffix>
                   <span class="text-sm text-gray-400">Cores</span>
                </template>
              </NStatistic>
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" :bordered="false" class="h-full">
              <NStatistic label="操作系统">
                {{ status.platform.os }}
                <template #suffix>
                  <span class="text-xs text-gray-400">/ {{ status.platform.arch }}</span>
                </template>
              </NStatistic>
            </NCard>
          </NGi>
        </NGrid>

        <!-- Detailed Metrics -->
        <NGrid :x-gap="16" :y-gap="16" cols="1 m:2" responsive="screen">
          <!-- Memory -->
          <NGi>
             <NCard title="内存概览" :bordered="false" size="small">
                <div class="flex items-center justify-between mb-6 px-2">
                   <div class="space-y-1">
                      <div class="text-3xl font-light text-primary">
                         {{ memoryPercentage }}%
                      </div>
                      <div class="text-xs text-gray-500">当前使用率</div>
                   </div>
                   <NProgress
                      type="circle"
                      :percentage="memoryPercentage"
                      :status="getStatusType(memoryPercentage)"
                      :radius="40"
                      :stroke-width="6"
                      :show-indicator="false"
                    />
                </div>
                <NDescriptions :column="2" size="small" label-placement="left" content-style="text-align: right;">
                   <NDescriptionsItem label="当前分配 (Alloc)">{{ formatBytes(status.memory.alloc) }}</NDescriptionsItem>
                   <NDescriptionsItem label="累积分配 (Total)">{{ formatBytes(status.memory.totalAlloc) }}</NDescriptionsItem>
                   <NDescriptionsItem label="系统占用 (Sys)">{{ formatBytes(status.memory.sys) }}</NDescriptionsItem>
                   <NDescriptionsItem label="GC 次数">{{ status.memory.numGC }}</NDescriptionsItem>
                </NDescriptions>
             </NCard>
          </NGi>

          <!-- Disk -->
          <NGi>
             <NCard title="磁盘空间" :bordered="false" size="small">
                <div class="flex items-center justify-between mb-6 px-2">
                   <div class="space-y-1">
                      <div class="text-3xl font-light text-primary">
                         {{ diskPercentage }}%
                      </div>
                      <div class="text-xs text-gray-500">已使用</div>
                   </div>
                   <!-- Visualization of disk space -->
                   <div class="w-1/3">
                      <NProgress
                        type="line"
                        :percentage="diskPercentage"
                        :status="getStatusType(diskPercentage)"
                        processing
                        :height="8"
                      />
                   </div>
                </div>
                 <NDescriptions :column="1" size="small" label-placement="left" content-style="text-align: right;">
                   <NDescriptionsItem label="挂载路径">
                      <span class="truncate max-w-37.5 inline-block align-bottom" :title="status.disk.path">{{ status.disk.path }}</span>
                   </NDescriptionsItem>
                   <NDescriptionsItem label="容量状态">
                      {{ formatBytes(status.disk.used) }} / {{ formatBytes(status.disk.all) }}
                   </NDescriptionsItem>
                   <NDescriptionsItem label="应用存储占用">
                      {{ formatBytes(status.storage.size) }}
                      <span class="text-xs text-gray-400 ml-1">({{ status.storage.path }})</span>
                   </NDescriptionsItem>
                </NDescriptions>
             </NCard>
          </NGi>
        </NGrid>


        <!-- Database & Redis (Full Detail) -->
        <NGrid :x-gap="16" :y-gap="16" cols="1" responsive="screen">
          <NGi>
            <NCard title="服务连接状态" :bordered="false" size="small">
              <NGrid :x-gap="24" :y-gap="24" cols="1 m:2">
                <NGi>
                  <div class="flex items-center gap-3 mb-4">
                     <span class="text-base font-medium">PostgreSQL</span>
                     <NTag :type="status.database.status === 'connected' ? 'success' : 'error'" size="small" round bordered>
                        {{ status.database.status }}
                     </NTag>
                  </div>
                  <NDescriptions :column="2" size="small" label-placement="left">
                     <NDescriptionsItem label="驱动">{{ status.database.driver }}</NDescriptionsItem>
                     <NDescriptionsItem label="打开连接">{{ status.database.poolStats.openConnections }}</NDescriptionsItem>
                     <NDescriptionsItem label="使用中">{{ status.database.poolStats.inUse }}</NDescriptionsItem>
                     <NDescriptionsItem label="空闲连接">{{ status.database.poolStats.idle }}</NDescriptionsItem>
                     <NDescriptionsItem label="最大连接">{{ status.database.poolStats.maxOpenConnections }}</NDescriptionsItem>
                     <NDescriptionsItem label="等待计数">{{ status.database.poolStats.waitCount }}</NDescriptionsItem>
                  </NDescriptions>
                </NGi>
                 <!-- Divider for mobile/desktop handled by grid gap usually, but manual divider can help if needed. Naive grid handles this well. -->
                <NGi>
                  <div class="border-t md:border-t-0 md:border-l border-gray-100 dark:border-gray-800 pt-4 md:pt-0 md:pl-6 h-full">
                      <div class="flex items-center gap-3 mb-4">
                        <span class="text-base font-medium">Redis</span>
                        <NTag :type="status.redis.status === 'connected' ? 'success' : 'error'" size="small" round bordered>
                            {{ status.redis.status }}
                        </NTag>
                      </div>
                      <NDescriptions :column="1" size="small" label-placement="left">
                        <NDescriptionsItem label="内存使用">{{ status.redis.usedMemory }}</NDescriptionsItem>
                        <NDescriptionsItem label="状态">
                           {{ status.redis.status === 'connected' ? '连接正常，准备就绪' : '连接失败' }}
                        </NDescriptionsItem>
                      </NDescriptions>
                  </div>
                </NGi>
              </NGrid>
            </NCard>
          </NGi>
        </NGrid>
        
        <!-- Raw Info Footer (Collapsible or subtle) -->
        <div class="text-center">
           <span class="text-xs text-gray-400">系统版本: {{ status.app.version }} | 启动时间: {{ new Date(status.app.startTime).toLocaleString() }}</span>
        </div>

      </div>
      <NResult
         v-else-if="!loading"
         status="info"
         title="暂无数据"
         description="无法获取系统监控信息，请检查后端服务是否正常运行。"
         class="mt-20"
      >
        <template #footer>
          <NButton @click="fetchData">重试</NButton>
        </template>
      </NResult>
    </NSpin>
  </ScrollContainer>
</template>