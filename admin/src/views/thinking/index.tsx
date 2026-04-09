import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NPagination,
  NSpace,
  NPopconfirm,
  useDialog,
} from 'naive-ui'
import { defineComponent, onMounted, ref, Transition } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { deleteThinking, listThinkings, batchDeleteThinkings } from '@/services/thinking'

import type { ThinkingListItem } from '@/services/thinking'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'

export default defineComponent({
  name: 'ThinkingList',
  setup() {
    const router = useRouter()
    const dialog = useDialog()
    const { message } = useDiscreteApi()
    const { data, loading, pagination, refresh } = useTable<ThinkingListItem>(listThinkings)
    const checkedRowKeys = ref<DataTableRowKey[]>([])

    const handleEdit = (id: number) => {
      router.push({ name: 'thinkingEdit', params: { id } })
    }

    const handleCreate = () => {
      router.push({ name: 'thinkingCreate' })
    }

    const handleDelete = (id: number) => {
      dialog.warning({
        title: '确认删除',
        content: '删除后无法恢复，是否继续？',
        positiveText: '确认',
        negativeText: '取消',
        onPositiveClick: async () => {
          await deleteThinking(id)
          await refresh()
        },
      })
    }

    const handleCheck = (rowKeys: DataTableRowKey[]) => {
      checkedRowKeys.value = rowKeys
    }

    const handleBatchDelete = async () => {
      const ids = checkedRowKeys.value as number[]
      if (ids.length === 0) return
      try {
        await batchDeleteThinkings({ ids })
        checkedRowKeys.value = []
        message.success('批量删除成功')
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const columns: DataTableColumns<ThinkingListItem> = [
      {
        type: 'selection',
      },
      {
        title: '内容',
        key: 'content',
        minWidth: 300,
        ellipsis: { tooltip: true },
        render: (row) => (
          <div class='font-medium text-gray-700 dark:text-gray-200'>{row.content}</div>
        ),
      },
      {
        title: '作者',
        key: 'authorName',
        width: 140,
        render: (row) => row.authorName || <span class='text-gray-400'>-</span>,
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
            <NButton
              size='small'
              type='error'
              secondary
              onClick={() => handleDelete(row.id)}
            >
              删除
            </NButton>
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
            <div class='text-lg font-medium'>思考列表</div>
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
                        default: () => `确定删除选中的 ${checkedRowKeys.value.length} 条思考吗？`,
                      }}
                    </NPopconfirm>
                  </NSpace>
                )}
              </Transition>
              <NButton
                type='primary'
                onClick={handleCreate}
              >
                新建思考
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
