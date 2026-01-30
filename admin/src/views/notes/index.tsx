import { NCard, NDataTable, NButton, NTag, NPagination, NSpace } from 'naive-ui'
import { defineComponent, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { listMoments } from '@/services/moments'

import type { MomentListItem } from '@/services/moments'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
  name: 'NoteList',
  setup() {
    const router = useRouter()
    const { data, loading, pagination, refresh } = useTable<MomentListItem>(listMoments)

    const handleEdit = (id: number) => {
      router.push({ name: 'noteEdit', params: { id } })
    }

    const handleCreate = () => {
      router.push({ name: 'noteCreate' })
    }

    const columns: DataTableColumns<MomentListItem> = [
      {
        title: '标题',
        key: 'title',
        width: 260,
        render: (row) => (
          <div class='font-medium text-gray-700 dark:text-gray-200'>{row.title}</div>
        ),
      },
      {
        title: '分区',
        key: 'columnName',
        width: 140,
        render: (row) => row.columnName || <span class='text-gray-400'>-</span>,
      },
      {
        title: '话题',
        key: 'topics',
        render: (row) => {
          if (!row.topics || row.topics.length === 0) return '-'
          return (
            <NSpace size={4}>
              {row.topics.map((topic) => (
                <NTag
                  size='small'
                  type='info'
                  bordered={false}
                >
                  {topic}
                </NTag>
              ))}
            </NSpace>
          )
        },
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
        width: 120,
        fixed: 'right',
        render: (row) => (
          <NButton
            size='small'
            type='primary'
            secondary
            onClick={() => handleEdit(row.id)}
          >
            编辑
          </NButton>
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
            <div class='text-lg font-medium'>手记列表</div>
            <NButton
              type='primary'
              onClick={handleCreate}
            >
              新建手记
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
