import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NPagination,
  NSpace,
  NPopconfirm,
  NTooltip,
  NDropdown,
} from 'naive-ui'
import { defineComponent, onMounted, ref, Transition } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import {
  listMoments,
  deleteMoment,
  batchSetMomentPublished,
  batchSetMomentTop,
  batchDeleteMoments,
} from '@/services/moments'
import { listWebsiteInfo } from '@/services/website-info'

import type { MomentListItem } from '@/services/moments'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'

export default defineComponent({
  name: 'NoteList',
  setup() {
    const router = useRouter()
    const { message } = useDiscreteApi()
    const { data, loading, pagination, refresh } = useTable<MomentListItem>(listMoments)
    const checkedRowKeys = ref<DataTableRowKey[]>([])
    const publicUrl = ref('')

    function normalizePublicUrl(value: string) {
      return value.trim().replace(/\/+$/, '')
    }

    function buildMomentPath(shortUrl: string, createdAt: string) {
      const matched = createdAt.match(/^(\d{4})-(\d{2})-(\d{2})/)
      if (!matched) return `/moments/${encodeURIComponent(shortUrl)}`
      const [, year, month, day] = matched
      return `/moments/${year}/${month}/${day}/${encodeURIComponent(shortUrl)}`
    }

    async function fetchWebsiteInfo() {
      try {
        const list = await listWebsiteInfo()
        const item = list?.find((info) => info.key === 'public_url')
        publicUrl.value = item?.value?.trim() ?? ''
      } catch (err) {
        message.error(err instanceof Error ? err.message : '加载站点地址失败')
      }
    }

    onMounted(() => {
      fetchWebsiteInfo()
    })

    const handleEdit = (id: number) => {
      router.push({ name: 'noteEdit', params: { id } })
    }

    const handleCreate = () => {
      router.push({ name: 'noteCreate' })
    }

    const handleDelete = async (id: number) => {
      try {
        await deleteMoment(id)
        message.success('删除成功')
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    }

    const handleCheck = (rowKeys: DataTableRowKey[]) => {
      checkedRowKeys.value = rowKeys
    }

    const handleTogglePublished = async (row: MomentListItem) => {
      try {
        await batchSetMomentPublished({ ids: [row.id], isPublished: !row.isPublished })
        row.isPublished = !row.isPublished
        message.success(row.isPublished ? '已发布' : '已设为草稿')
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const handleToggleTop = async (row: MomentListItem) => {
      try {
        await batchSetMomentTop({ ids: [row.id], isTop: !row.isTop })
        row.isTop = !row.isTop
        message.success(row.isTop ? '已置顶' : '已取消置顶')
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const handleBatchPublish = async (isPublished: boolean) => {
      const ids = checkedRowKeys.value as number[]
      if (ids.length === 0) return
      try {
        await batchSetMomentPublished({ ids, isPublished })
        data.value.forEach((item) => {
          if (ids.includes(item.id)) item.isPublished = isPublished
        })
        checkedRowKeys.value = []
        message.success(isPublished ? '批量发布成功' : '批量取消发布成功')
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const handleBatchDelete = async () => {
      const ids = checkedRowKeys.value as number[]
      if (ids.length === 0) return
      try {
        await batchDeleteMoments({ ids })
        checkedRowKeys.value = []
        message.success('批量删除成功')
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '操作失败')
      }
    }

    const batchPublishOptions = [
      { label: '设为已发布', key: 'publish' },
      { label: '设为草稿', key: 'unpublish' },
    ]

    const handleBatchPublishSelect = (key: string) => {
      handleBatchPublish(key === 'publish')
    }

    const columns: DataTableColumns<MomentListItem> = [
      {
        type: 'selection',
      },
      {
        title: '标题',
        key: 'title',
        minWidth: 280,
        render: (row) => (
          <div class='font-medium text-gray-700 dark:text-gray-200'>
            <span>{row.title}</span>
            {row.isHot && (
              <NTooltip trigger='hover'>
                {{
                  trigger: () => (
                    <span class='ml-2 iconify size-4 cursor-help align-middle text-red-500 ph--fire-fill' />
                  ),
                  default: () => (
                    <div class='flex flex-col gap-y-0.5'>
                      <span class='font-bold'>热门手记</span>
                      <span class='text-xs opacity-80'>
                        热门标准：浏览量 &gt; 1000 或 点赞数 &gt; 50
                      </span>
                    </div>
                  ),
                }}
              </NTooltip>
            )}
            <div
              class='inline-block cursor-pointer'
              onClick={() => {
                window.open(
                  `${normalizePublicUrl(publicUrl.value)}${buildMomentPath(row.shortUrl, row.createdAt)}`,
                  '_blank',
                )
              }}
            >
              <span class='ml-2 iconify size-4 cursor-pointer align-middle text-black/50 ph--link-simple dark:text-gray-400' />
            </div>
          </div>
        ),
        sorter: 'default',
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
        minWidth: 160,
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
        title: '是否发布',
        key: 'isPublished',
        width: 100,
        render: (row) => (
          <span
            style={{ cursor: 'pointer' }}
            onClick={() => handleTogglePublished(row)}
          >
            <NTag
              size='small'
              type={row.isPublished ? 'success' : 'default'}
              bordered={false}
            >
              {{
                default: () => (row.isPublished ? '已发布' : '草稿'),
                icon: () => (
                  <span
                    class={`iconify ${row.isPublished ? 'ph--check-circle' : 'ph--circle-dashed'} size-3.5`}
                  />
                ),
              }}
            </NTag>
          </span>
        ),
        sorter: (row1, row2) => Number(row1.isPublished) - Number(row2.isPublished),
      },
      {
        title: '属性',
        key: 'attributes',
        width: 160,
        render: (row) => (
          <NSpace size={4}>
            <span
              style={{ cursor: 'pointer' }}
              onClick={() => handleToggleTop(row)}
            >
              <NTag
                size='small'
                type={row.isTop ? 'warning' : 'default'}
                bordered={false}
              >
                {{
                  default: () => (row.isTop ? '置顶' : '未置顶'),
                  icon: () => (
                    <span
                      class={`iconify ${row.isTop ? 'ph--push-pin-fill' : 'ph--push-pin'} size-3.5`}
                    />
                  ),
                }}
              </NTag>
            </span>
            {row.isOriginal ? (
              <NTag
                size='small'
                type='success'
                bordered={false}
              >
                原创
              </NTag>
            ) : (
              <NTag
                size='small'
                type='default'
                bordered={false}
              >
                转载
              </NTag>
            )}
          </NSpace>
        ),
      },
      {
        title: '浏览',
        key: 'views',
        width: 80,
        render: (row) => <span class='font-mono text-xs text-gray-500'>{row.views}</span>,
        sorter: 'default',
      },
      {
        title: '点赞',
        key: 'likes',
        width: 80,
        render: (row) => <span class='font-mono text-xs text-gray-500'>{row.likes}</span>,
        sorter: 'default',
      },
      {
        title: '创建时间',
        key: 'createdAt',
        width: 180,
        render: (row) => new Date(row.createdAt).toLocaleString(),
        sorter: (row1, row2) =>
          new Date(row1.createdAt).getTime() - new Date(row2.createdAt).getTime(),
      },
      {
        title: '更新时间',
        key: 'updatedAt',
        width: 180,
        render: (row) => new Date(row.updatedAt).toLocaleString(),
        sorter: (row1, row2) =>
          new Date(row1.updatedAt).getTime() - new Date(row2.updatedAt).getTime(),
      },
      {
        title: '操作',
        key: 'actions',
        width: 160,
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
            <NPopconfirm
              onPositiveClick={() => handleDelete(row.id)}
              v-slots={{
                trigger: () => (
                  <NButton
                    size='small'
                    type='error'
                    secondary
                  >
                    删除
                  </NButton>
                ),
              }}
            >
              确定删除吗？
            </NPopconfirm>
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
            <div class='text-lg font-medium'>手记列表</div>
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
                      options={batchPublishOptions}
                      onSelect={handleBatchPublishSelect}
                    >
                      <NButton
                        size='small'
                        secondary
                      >
                        批量发布
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
                        default: () => `确定删除选中的 ${checkedRowKeys.value.length} 条手记吗？`,
                      }}
                    </NPopconfirm>
                  </NSpace>
                )}
              </Transition>
              <NButton
                type='primary'
                onClick={handleCreate}
              >
                新建手记
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
            scrollX={1360}
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
