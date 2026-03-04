import {
  NButton,
  NCard,
  NDataTable,
  NDropdown,
  NPagination,
  NSelect,
  NTag,
  NThing,
  NTooltip,
  useMessage,
} from 'naive-ui'
import { defineComponent, reactive } from 'vue'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { friendLinkService } from '@/services/friend-links'

import type { FriendLinkApplication } from '@/types/friend-link'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
  name: 'FriendLinkApplications',
  setup() {
    const message = useMessage()

    const appsFilter = reactive({
      status: undefined as string | undefined,
    })

    const {
      data: apps,
      loading: appsLoading,
      pagination: appsPagination,
      refresh: refreshApps,
    } = useTable<FriendLinkApplication>(friendLinkService.getApplications, appsFilter as any)

    const handleAppStatusUpdate = async (id: number, status: string) => {
      try {
        await friendLinkService.updateApplicationStatus(id, status)
        message.success('状态变更成功')
        refreshApps()
      }
      catch (e: any) {
        message.error(e.message || '操作失败')
      }
    }

    const appColumns: DataTableColumns<FriendLinkApplication> = [
      {
        title: '申请信息',
        key: 'info',
        render: (row) => (
          <NThing title={row.name} description={row.url}>
            {{
              avatar: () => row.logo ? <img src={row.logo} class="w-10 h-10 rounded" /> : null,
            }}
          </NThing>
        ),
      },
      {
        title: '来源',
        key: 'channel',
        width: 100,
        render: (row) => (
          <div>
            <NTag size="small">{row.applyChannel}</NTag>
            {row.userId && <div class="text-xs text-neutral-400 mt-1">UID: {row.userId}</div>}
          </div>
        ),
      },
      {
        title: '留言',
        key: 'message',
        minWidth: 220,
        render: (row) => {
          const content = row.message?.trim()
          if (!content)
            return '-'
          return (
            <NTooltip placement="top-start">
              {{
                trigger: () => (
                  <div class="max-w-[260px] truncate text-sm text-neutral-600">
                    {content}
                  </div>
                ),
                default: () => (
                  <div class="max-w-[420px] whitespace-pre-wrap break-all">
                    {content}
                  </div>
                ),
              }}
            </NTooltip>
          )
        },
      },
      {
        title: '状态',
        key: 'status',
        width: 90,
        render: (row) => {
          const typeMap: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
            pending: 'warning',
            approved: 'success',
            rejected: 'error',
            blocked: 'error',
          }
          return (
            <NTag type={typeMap[row.status] || 'default'} size="small">
              {{ default: () => row.status }}
            </NTag>
          )
        },
      },
      {
        title: '时间',
        key: 'createdAt',
        width: 160,
      },
      {
        title: '操作',
        key: 'actions',
        width: 150,
        render: (row) => (
          <NDropdown
            trigger="click"
            options={[
              { label: '通过 (Approve)', key: 'approved' },
              { label: '拒绝 (Reject)', key: 'rejected' },
              { label: '封禁 (Block)', key: 'blocked' },
              { label: '重置为待审核 (Pending)', key: 'pending' },
            ]}
            onSelect={(key: string) => handleAppStatusUpdate(row.id, key)}
          >
            <NButton size="tiny" secondary>
              变更状态
            </NButton>
          </NDropdown>
        ),
      },
    ]

    return () => (
      <ScrollContainer wrapper-class="p-4">
        <NCard title="友链申请审核" class="h-full">
          <div class="mb-4 flex gap-2">
            <NSelect
              value={appsFilter.status}
              placeholder="状态筛选"
              clearable
              options={[
                { label: '待审核 (Pending)', value: 'pending' },
                { label: '已通过 (Approved)', value: 'approved' },
                { label: '已拒绝 (Rejected)', value: 'rejected' },
                { label: '已封禁 (Blocked)', value: 'blocked' },
              ]}
              class="w-40"
              onUpdateValue={(v) => {
                appsFilter.status = v
                refreshApps()
              }}
            />
            <NButton secondary onClick={refreshApps}>
              刷新
            </NButton>
          </div>
          <NDataTable
            remote
            columns={appColumns}
            data={apps.value}
            loading={appsLoading.value}
            row-key={(row: FriendLinkApplication) => row.id}
            scrollX={800}
          />
          <div class="mt-4 flex justify-end">
            <NPagination
              page={appsPagination.page}
              page-size={appsPagination.pageSize}
              item-count={appsPagination.itemCount}
              show-size-picker={appsPagination.showSizePicker}
              page-sizes={appsPagination.pageSizes}
              onUpdatePage={appsPagination.onChange}
              onUpdatePageSize={appsPagination.onUpdatePageSize}
            />
          </div>
        </NCard>
      </ScrollContainer>
    )
  },
})
