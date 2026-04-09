import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NPagination,
  NSpace,
  NPopconfirm,
  NDropdown,
  useDialog,
} from 'naive-ui'
import { defineComponent, onMounted, ref, Transition } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { deletePage, listPages, batchSetPageEnabled, batchDeletePages } from '@/services/page'

import type { PageListItem } from '@/services/page'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'

export default defineComponent({
  name: 'PageList',
  setup() {
    const router = useRouter()
    const dialog = useDialog()
    const { message } = useDiscreteApi()
    const { data, loading, pagination, refresh } = useTable<PageListItem>(listPages)
    const checkedRowKeys = ref<DataTableRowKey[]>([])

    const handleEdit = (id: number) => {
      router.push({ name: 'pageEdit', params: { id } })
    }

    const handleCreate = () => {
      router.push({ name: 'pageCreate' })
    }

    const handleDelete = (id: number) => {
      dialog.warning({
        title: '确认删除',
        content: '删除后无法恢复，是否继续？',
        positiveText: '确认',
        negativeText: '取消',
        onPositiveClick: async () => {
          await deletePage(id)
          await refresh()
        },
      })
    }

    const handleCheck = (rowKeys: DataTableRowKey[]) => {
      // Exclude builtin pages from batch selection
      checkedRowKeys.value = rowKeys.filter((key) => {
        const row = data.value.find((item) => item.id === key)
        return row && !row.isBuiltin
      })
    }

    const handleToggleEnabled = async (row: PageListItem) => {
      try {
        await batchSetPageEnabled({ ids: [row.id], isEnabled: !row.isEnabled })
        row.isEnabled = !row.isEnabled
        message.success(row.isEnabled ? '已启用' : '已禁用')
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const handleBatchEnabled = async (isEnabled: boolean) => {
      const ids = checkedRowKeys.value as number[]
      if (ids.length === 0) return
      try {
        await batchSetPageEnabled({ ids, isEnabled })
        data.value.forEach((item) => {
          if (ids.includes(item.id)) item.isEnabled = isEnabled
        })
        checkedRowKeys.value = []
        message.success(isEnabled ? '批量启用成功' : '批量禁用成功')
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const handleBatchDelete = async () => {
      const ids = checkedRowKeys.value as number[]
      if (ids.length === 0) return
      try {
        await batchDeletePages({ ids })
        checkedRowKeys.value = []
        message.success('批量删除成功')
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const batchEnabledOptions = [
      { label: '设为启用', key: 'enable' },
      { label: '设为禁用', key: 'disable' },
    ]

    const handleBatchEnabledSelect = (key: string) => {
      handleBatchEnabled(key === 'enable')
    }

    const columns: DataTableColumns<PageListItem> = [
      {
        type: 'selection',
      },
      {
        title: '标题',
        key: 'title',
        width: 260,
        render: (row) => (
          <div class='flex items-center gap-2 font-medium text-gray-700 dark:text-gray-200'>
            {row.title}
            {row.isBuiltin && (
              <NTag
                size='tiny'
                type='info'
                bordered={false}
              >
                内置
              </NTag>
            )}
          </div>
        ),
      },
      {
        title: '短链接',
        key: 'shortUrl',
        width: 140,
        render: (row) => row.shortUrl || <span class='text-gray-400'>-</span>,
      },
      {
        title: '是否启用',
        key: 'isEnabled',
        width: 100,
        render: (row) => (
          <span
            style={{ cursor: 'pointer' }}
            onClick={() => handleToggleEnabled(row)}
          >
            <NTag
              size='small'
              type={row.isEnabled ? 'success' : 'default'}
              bordered={false}
            >
              {{
                default: () => (row.isEnabled ? '已启用' : '已禁用'),
                icon: () => (
                  <span
                    class={`iconify ${row.isEnabled ? 'ph--check-circle' : 'ph--circle-dashed'} size-3.5`}
                  />
                ),
              }}
            </NTag>
          </span>
        ),
      },
      {
        title: '数据 (阅/赞/评)',
        key: 'metrics',
        width: 180,
        render: (row) => (
          <span class='font-mono text-xs text-gray-500'>
            {row.views} / {row.likes} / {row.comments}
          </span>
        ),
      },
      {
        title: '更新时间',
        key: 'updatedAt',
        width: 180,
        render: (row) => new Date(row.updatedAt).toLocaleString(),
      },
      {
        title: '操作',
        key: 'actions',
        width: 180,
        fixed: 'right',
        render: (row) => (
          <NSpace>
            <NButton
              size='small'
              type='primary'
              secondary
              onClick={() => handleEdit(row.id)}
            >
              编辑
            </NButton>
            {!row.isBuiltin && (
              <NButton
                size='small'
                type='error'
                secondary
                onClick={() => handleDelete(row.id)}
              >
                删除
              </NButton>
            )}
          </NSpace>
        ),
      },
    ]

    onMounted(() => {
      refresh()
    })

    return () => (
      <ScrollContainer wrapperClass='flex flex-col gap-y-4'>
        <NCard bordered={false}>
          <div class='flex items-center justify-between'>
            <div class='text-lg font-medium'>页面列表</div>
            <NSpace
              align='center'
              size={12}
            >
              <Transition name='fade'>
                {checkedRowKeys.value.length > 0 && (
                  <NSpace
                    align='center'
                    size={8}
                  >
                    <NTag
                      type='info'
                      size='small'
                    >
                      已选 {checkedRowKeys.value.length} 项
                    </NTag>
                    <NDropdown
                      options={batchEnabledOptions}
                      onSelect={handleBatchEnabledSelect}
                    >
                      <NButton
                        size='small'
                        secondary
                      >
                        批量启用
                      </NButton>
                    </NDropdown>
                    <NPopconfirm onPositiveClick={handleBatchDelete}>
                      {{
                        trigger: () => (
                          <NButton
                            size='small'
                            type='error'
                            secondary
                          >
                            批量删除
                          </NButton>
                        ),
                        default: () => `确定删除选中的 ${checkedRowKeys.value.length} 个页面吗？`,
                      }}
                    </NPopconfirm>
                  </NSpace>
                )}
              </Transition>
              <NButton
                type='primary'
                onClick={handleCreate}
              >
                新建页面
              </NButton>
            </NSpace>
          </div>
        </NCard>

        <NCard
          bordered={false}
          contentStyle={{ padding: '0' }}
        >
          <NDataTable
            columns={columns}
            data={data.value}
            loading={loading.value}
            rowKey={(row) => row.id}
            checkedRowKeys={checkedRowKeys.value}
            onUpdateCheckedRowKeys={handleCheck}
            bordered={false}
            scrollX={1100}
          />

          <div class='flex justify-end p-4'>
            <NPagination
              v-model:page={pagination.page}
              v-model:pageSize={pagination.pageSize}
              itemCount={pagination.itemCount}
              pageSizes={pagination.pageSizes}
              showSizePicker={pagination.showSizePicker}
              onUpdatePage={pagination.onChange}
              onUpdatePageSize={pagination.onUpdatePageSize}
            />
          </div>
        </NCard>
      </ScrollContainer>
    )
  },
})
