import { NCard, NDataTable, NButton, NTag, NPagination, NSpace, useDialog } from 'naive-ui'
import { defineComponent, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { deletePage, listPages } from '@/services/page'

import type { PageListItem } from '@/services/page'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
  name: 'PageList',
  setup() {
    const router = useRouter()
    const dialog = useDialog()
    const { data, loading, pagination, refresh } = useTable<PageListItem>(listPages)

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

    const columns: DataTableColumns<PageListItem> = [
      {
        title: '标题',
        key: 'title',
        width: 260,
        render: (row) => (
          <div class='font-medium text-gray-700 dark:text-gray-200'>{row.title}</div>
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
        width: 80,
        render: (row) => (row.isEnabled ? <NTag type='success'>是</NTag> : <NTag type='warning'>否</NTag>),
      },
      {
        title: '数据 (阅/赞/评)',
        key: 'metrics',
        width: 180,
        render: (row) => (
          <span class='font-mono text-xs text-gray-500'>
            {row.metrics.views} / {row.metrics.likes} / {row.metrics.comments}
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

    onMounted(()=>{
      refresh()
    })

    return () => (
      <ScrollContainer wrapperClass='flex flex-col gap-y-4'>
        <NCard bordered={false}>
          <div class='flex items-center justify-between'>
            <div class='text-lg font-medium'>页面列表</div>
            <NButton
              type='primary'
              onClick={handleCreate}
            >
              新建页面
            </NButton>
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
            bordered={false}
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
