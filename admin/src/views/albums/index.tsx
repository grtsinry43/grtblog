import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NPagination,
  NSpace,
  NPopconfirm,
  NDropdown,
  NImage,
} from 'naive-ui'
import { defineComponent, ref, Transition } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import {
  listAlbums,
  deleteAlbum,
  batchSetAlbumPublished,
  batchDeleteAlbums,
} from '@/services/albums'

import type { AlbumListItem } from '@/services/albums'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'

export default defineComponent({
  name: 'AlbumList',
  setup() {
    const router = useRouter()
    const { message } = useDiscreteApi()
    const { data, loading, pagination, refresh } = useTable<AlbumListItem>(listAlbums)
    const checkedRowKeys = ref<DataTableRowKey[]>([])

    function handleCreate() {
      router.push({ name: 'albumCreate' })
    }

    function handleEdit(id: number) {
      router.push({ name: 'albumEdit', params: { id } })
    }

    async function handleDelete(id: number) {
      try {
        await deleteAlbum(id)
        message.success('删除成功')
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '删除失败')
      }
    }

    function handleBatchPublishSelect(key: string) {
      const ids = checkedRowKeys.value.map(Number)
      if (ids.length === 0) return
      batchSetAlbumPublished({ ids, isPublished: key === 'publish' })
        .then(() => {
          message.success(key === 'publish' ? '已批量发布' : '已批量取消发布')
          checkedRowKeys.value = []
          refresh()
        })
        .catch((err) => {
          message.error(err instanceof Error ? err.message : '操作失败')
        })
    }

    async function handleBatchDelete() {
      const ids = checkedRowKeys.value.map(Number)
      if (ids.length === 0) return
      try {
        await batchDeleteAlbums({ ids })
        message.success('批量删除成功')
        checkedRowKeys.value = []
        refresh()
      } catch (err) {
        message.error(err instanceof Error ? err.message : '批量删除失败')
      }
    }

    const batchPublishOptions = [
      { label: '批量发布', key: 'publish' },
      { label: '批量取消发布', key: 'unpublish' },
    ]

    const columns: DataTableColumns<AlbumListItem> = [
      { type: 'selection' },
      {
        title: '封面',
        key: 'cover',
        width: 72,
        render(row) {
          return row.cover ? (
            <NImage
              src={row.cover}
              width={48}
              height={48}
              objectFit='cover'
              previewDisabled
              imgProps={{ style: 'border-radius: 6px;' }}
            />
          ) : (
            <div class='flex h-12 w-12 items-center justify-center rounded-md bg-current/5'>
              <div class='iconify text-lg opacity-30 ph--image' />
            </div>
          )
        },
      },
      {
        title: '标题',
        key: 'title',
        minWidth: 180,
        ellipsis: { tooltip: true },
        render(row) {
          return (
            <NButton
              text
              type='primary'
              onClick={() => handleEdit(row.id)}
            >
              {row.title}
            </NButton>
          )
        },
      },
      {
        title: '照片数',
        key: 'photoCount',
        width: 80,
        align: 'center',
        render(row) {
          return (
            <NTag
              size='small'
              round
              bordered={false}
            >
              {row.photoCount}
            </NTag>
          )
        },
      },
      {
        title: '状态',
        key: 'isPublished',
        width: 80,
        align: 'center',
        render(row) {
          return (
            <NTag
              type={row.isPublished ? 'success' : 'default'}
              size='small'
              bordered={false}
            >
              {row.isPublished ? '已发布' : '草稿'}
            </NTag>
          )
        },
      },
      {
        title: '浏览',
        key: 'views',
        width: 70,
        align: 'center',
      },
      {
        title: '点赞',
        key: 'likes',
        width: 70,
        align: 'center',
      },
      {
        title: '创建时间',
        key: 'createdAt',
        width: 160,
        render(row) {
          return new Date(row.createdAt).toLocaleString('zh-CN')
        },
        sorter: 'default',
      },
      {
        title: '操作',
        key: 'actions',
        width: 120,
        align: 'center',
        render(row) {
          return (
            <NSpace
              justify='center'
              size={4}
            >
              <NButton
                text
                type='primary'
                size='small'
                onClick={() => handleEdit(row.id)}
              >
                编辑
              </NButton>
              <NPopconfirm onPositiveClick={() => handleDelete(row.id)}>
                {{
                  trigger: () => (
                    <NButton
                      text
                      type='error'
                      size='small'
                    >
                      删除
                    </NButton>
                  ),
                  default: () => '确定要删除这个相册吗？',
                }}
              </NPopconfirm>
            </NSpace>
          )
        },
      },
    ]

    return () => (
      <ScrollContainer wrapperClass='flex flex-col gap-y-4'>
        {/* 顶部操作栏 */}
        <NCard bordered={false}>
          <div class='flex items-center justify-between'>
            <div class='text-lg font-medium'>相册列表</div>
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
                        default: () => `确定删除选中的 ${checkedRowKeys.value.length} 个相册吗？`,
                      }}
                    </NPopconfirm>
                  </NSpace>
                )}
              </Transition>
              <NButton
                type='primary'
                onClick={handleCreate}
              >
                {{
                  icon: () => <div class='iconify ph--plus' />,
                  default: () => '新建相册',
                }}
              </NButton>
            </NSpace>
          </div>
        </NCard>

        {/* 表格 */}
        <NCard
          bordered={false}
          contentStyle={{ padding: '0' }}
        >
          <NDataTable
            columns={columns}
            data={data.value}
            loading={loading.value}
            rowKey={(row: AlbumListItem) => row.id}
            checkedRowKeys={checkedRowKeys.value}
            onUpdateCheckedRowKeys={(keys: DataTableRowKey[]) => {
              checkedRowKeys.value = keys
            }}
            scrollX={900}
          />
          <div class='flex justify-end p-4'>
            <NPagination {...pagination} />
          </div>
        </NCard>
      </ScrollContainer>
    )
  },
})
