import {
  NButton,
  NCard,
  NDataTable,
  NInput,
  NPagination,
  NSelect,
  NTag,
  NTooltip,
} from 'naive-ui'
import { defineComponent, reactive } from 'vue'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { friendLinkService } from '@/services/friend-links'

import type { FriendLinkSyncJob } from '@/types/friend-link'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
  name: 'FriendLinkSyncJobs',
  setup() {
    const syncJobsFilter = reactive({
      status: undefined as string | undefined,
      targetType: undefined as string | undefined,
      syncMethod: undefined as string | undefined,
      keyword: '',
    })

    const {
      data: syncJobs,
      loading: syncJobsLoading,
      pagination: syncJobsPagination,
      refresh: refreshSyncJobs,
    } = useTable<FriendLinkSyncJob>(friendLinkService.getSyncJobs, syncJobsFilter as any)

    const syncJobColumns: DataTableColumns<FriendLinkSyncJob> = [
      {
        title: '目标',
        key: 'target',
        minWidth: 280,
        render: (row) => (
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <NTag size="small" bordered={false} type={row.targetType === 'friend_link' ? 'success' : 'info'}>
                {{ default: () => row.targetType === 'friend_link' ? '友链 RSS' : '联合实例' }}
              </NTag>
              <span class="text-xs text-neutral-400">#{row.friendLinkId || row.instanceId || '-'}</span>
            </div>
            <NTooltip placement="top-start">
              {{
                trigger: () => (
                  <div class="truncate text-sm text-neutral-600">
                    {row.targetUrl}
                  </div>
                ),
                default: () => (
                  <div class="max-w-[520px] break-all">
                    {row.targetUrl}
                  </div>
                ),
              }}
            </NTooltip>
          </div>
        ),
      },
      {
        title: '同步方式',
        key: 'syncMethod',
        width: 100,
        render: (row) => <NTag size="small">{row.syncMethod}</NTag>,
      },
      {
        title: '状态',
        key: 'status',
        width: 100,
        render: (row) => {
          const typeMap: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
            queued: 'default',
            running: 'warning',
            success: 'success',
            failed: 'error',
          }
          return <NTag type={typeMap[row.status] || 'default'} size="small">{row.status}</NTag>
        },
      },
      {
        title: '拉取',
        key: 'pulledCount',
        width: 80,
        render: (row) => row.pulledCount ?? 0,
      },
      {
        title: '耗时',
        key: 'durationMs',
        width: 90,
        render: (row) => row.durationMs != null ? `${row.durationMs}ms` : '-',
      },
      {
        title: '错误信息',
        key: 'errorMessage',
        minWidth: 220,
        render: (row) => {
          if (!row.errorMessage)
            return '-'
          return (
            <NTooltip placement="top-start">
              {{
                trigger: () => (
                  <div class="truncate text-sm text-red-500">
                    {row.errorMessage}
                  </div>
                ),
                default: () => (
                  <div class="max-w-[520px] whitespace-pre-wrap break-all text-red-500">
                    {row.errorMessage}
                  </div>
                ),
              }}
            </NTooltip>
          )
        },
      },
      {
        title: '创建时间',
        key: 'createdAt',
        width: 170,
      },
    ]

    return () => (
      <ScrollContainer wrapper-class="p-4">
        <NCard title="友链同步任务" class="h-full">
          <div class="mb-4 flex gap-2">
            <NSelect
              value={syncJobsFilter.status}
              placeholder="状态"
              clearable
              options={[
                { label: '排队中 (Queued)', value: 'queued' },
                { label: '运行中 (Running)', value: 'running' },
                { label: '成功 (Success)', value: 'success' },
                { label: '失败 (Failed)', value: 'failed' },
              ]}
              class="w-44"
              onUpdateValue={(v) => {
                syncJobsFilter.status = v
                refreshSyncJobs()
              }}
            />
            <NSelect
              value={syncJobsFilter.targetType}
              placeholder="目标类型"
              clearable
              options={[
                { label: '友链 RSS', value: 'friend_link' },
                { label: '联合实例', value: 'federation_instance' },
              ]}
              class="w-44"
              onUpdateValue={(v) => {
                syncJobsFilter.targetType = v
                refreshSyncJobs()
              }}
            />
            <NSelect
              value={syncJobsFilter.syncMethod}
              placeholder="同步方式"
              clearable
              options={[
                { label: 'Timeline', value: 'timeline' },
                { label: 'RSS', value: 'rss' },
              ]}
              class="w-40"
              onUpdateValue={(v) => {
                syncJobsFilter.syncMethod = v
                refreshSyncJobs()
              }}
            />
            <NInput
              value={syncJobsFilter.keyword}
              placeholder="搜索目标/错误信息"
              class="max-w-xs"
              clearable
              onUpdateValue={(v) => syncJobsFilter.keyword = v}
              onKeydown={(e) => {
                if (e.key === 'Enter')
                  refreshSyncJobs()
              }}
            />
            <NButton secondary onClick={refreshSyncJobs}>
              刷新
            </NButton>
          </div>
          <NDataTable
            remote
            columns={syncJobColumns}
            data={syncJobs.value}
            loading={syncJobsLoading.value}
            row-key={(row: FriendLinkSyncJob) => row.id}
            scrollX={1100}
          />
          <div class="mt-4 flex justify-end">
            <NPagination
              page={syncJobsPagination.page}
              page-size={syncJobsPagination.pageSize}
              item-count={syncJobsPagination.itemCount}
              show-size-picker={syncJobsPagination.showSizePicker}
              page-sizes={syncJobsPagination.pageSizes}
              onUpdatePage={syncJobsPagination.onChange}
              onUpdatePageSize={syncJobsPagination.onUpdatePageSize}
            />
          </div>
        </NCard>
      </ScrollContainer>
    )
  },
})
