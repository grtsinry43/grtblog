import {
  NCard,
  NDataTable,
  NButton,
  NTag,
  NPagination,
  NSpace,
  NPopconfirm,
  NTooltip,
} from 'naive-ui'
import { defineComponent, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { useDiscreteApi } from '@/composables/useDiscreteApi'
import { deleteArticle, listArticles } from '@/services/articles'

import type { ArticleListItem } from '@/services/articles'
import type { DataTableColumns, DataTableRowKey } from 'naive-ui'
import { listWebsiteInfo } from '@/services/website-info'
import Preview from './preview.vue'

export default defineComponent({
  name: 'ArticleList',
  setup() {
    const router = useRouter()
    const { message } = useDiscreteApi()
    const { data, loading, pagination, refresh } = useTable<ArticleListItem>(listArticles)
    const checkedRowKeys = ref<DataTableRowKey[]>([])
    const publicUrl = ref('')

    function normalizePublicUrl(value: string) {
  return value.trim().replace(/\/+$/, '')
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


    onMounted(()=>{
      fetchWebsiteInfo()
    })
    const handleEdit = (id: number) => {
      router.push({ name: 'articleEdit', params: { id } })
    }

    const handleCreate = () => {
      router.push({ name: 'articleCreate' })
    }

    const handleDelete = async (id: number) => {
      try {
        await deleteArticle(id)
        message.success('删除成功')
        refresh()
      } catch (err) {
        console.error(err)
      }
    }

    const handleCheck = (rowKeys: DataTableRowKey[]) => {
      checkedRowKeys.value = rowKeys
    }

    const columns: DataTableColumns<ArticleListItem> = [
      {
        type: 'selection',
      },
      {
        title: '标题',
        key: 'title',
        width: 400,
        render: (row) => (
          <div class='font-medium text-gray-700 dark:text-gray-200'>
              <span>{row.title}</span>
            {row.isHot && (
              <NTooltip trigger='hover'>
                {{
                  trigger: () => (
                    <span class='iconify ph--fire-fill size-4 cursor-help text-red-500 ml-2 align-middle' />
                  ),
                  default: () => (
                    <div class='flex flex-col gap-y-0.5'>
                      <span class='font-bold'>热门文章</span>
                      <span class='text-xs opacity-80'>热门标准：浏览量 &gt; 1000 或 点赞数 &gt; 50</span>
                    </div>
                  ),
                }}
              </NTooltip>
            )}
            <NTooltip trigger='hover'>
                {{
                  trigger: () => (
                    <span class='iconify ph--file-search size-4 cursor-help text-black/50 dark:text-gray-40 ml-2 align-middle' />
                  ),
                  default: () => (
                    <ScrollContainer class="max-w-120 max-h-80 overflow-auto text-sm">
                      <Preview articleId={row.id}/>
                    </ScrollContainer>
                  ),
                }}
              </NTooltip>
            <div class="cursor-pointer inline-block" onClick={
              () => {
                window.open(`${normalizePublicUrl(publicUrl.value)}/posts/${row.shortUrl}`, '_blank')
              }
            }>
              <span class='iconify ph--link-simple size-4 cursor-pointer text-black/50 dark:text-gray-400 ml-2 align-middle' />
            </div>
          </div>
        ),
        sorter: 'default',
      },
      {
        title: '分类',
        key: 'categoryName',
        width: 140,
        render: (row) => row.categoryName || <span class='text-gray-400'>-</span>,
        sorter: 'default',
      },
      {
        title: '标签',
        key: 'tags',
        render: (row) => {
          if (!row.tags || row.tags.length === 0) return '-'
          return (
            <NSpace size={4}>
              {row.tags.map((tag) => (
                <NTag
                  size='small'
                  type='info'
                  bordered={false}
                >
                  {tag}
                </NTag>
              ))}
            </NSpace>
          )
        },
      },
      {
        title: '是否发布',
        key: 'isPublished',
        width: 120,
        render: (row) => (
          row.isPublished ? (
            <NTag size='small' type='success' bordered={false}>已发布</NTag>
          ) : (
            <NTag size='small' type='default' bordered={false}>草稿</NTag>
          )
        ),
        sorter: (row1, row2) => Number(row1.isPublished) - Number(row2.isPublished),
      },
      {
        title: '属性',
        key: 'attributes',
        width: 180,
        render: (row) => (
          <NSpace size={4}>
            {row.isTop && <NTag size='small' type='warning' bordered={false}>置顶</NTag>}
            {row.isOriginal ? (
              <NTag size='small' type='success' bordered={false}>原创</NTag>
            ) : (
              <NTag size='small' type='default' bordered={false}>转载</NTag>
            )}
          </NSpace>
        ),
      },
      {
        title: '浏览',
        key: 'views',
        width: 100,
        render: (row) => (
          <span class='font-mono text-xs text-gray-500'>
            {row.views}
          </span>
        ),
        sorter: 'default',
      },
      {
        title: '点赞',
        key: 'likes',
        width: 100,
        render: (row) => (
          <span class='font-mono text-xs text-gray-500'>
            {row.likes}
          </span>
        ),
        sorter: 'default',
      },
      {
        title: '创建时间',
        key: 'createdAt',
        width: 180,
        render: (row) => new Date(row.createdAt).toLocaleString(),
        sorter: (row1, row2) => new Date(row1.createdAt).getTime() - new Date(row2.createdAt).getTime(),
      },
      {
        title: '更新时间',
        key: 'updatedAt',
        width: 180,
        render: (row) => new Date(row.updatedAt).toLocaleString(),
        sorter: (row1, row2) => new Date(row1.updatedAt).getTime() - new Date(row2.updatedAt).getTime(),
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

    onMounted(()=>{
      refresh()
    })

    // 4. 渲染视图
    return () => (
      <ScrollContainer wrapperClass='flex flex-col gap-y-4'>
        {/* 顶部操作栏 */}
        <NCard bordered={false}>
          <div class='flex items-center justify-between'>
            <div class='text-lg font-medium'>文章列表</div>
            <NButton
              type='primary'
              onClick={handleCreate}
            >
              新建文章
            </NButton>
          </div>
        </NCard>

        {/* 表格主体 */}
        <NCard
          bordered={false}
          contentStyle={{ padding: '0' }}
        >
          <NDataTable
            columns={columns}
            data={data.value}
            loading={loading.value}
            rowKey={(row) => row.id}
            onUpdateCheckedRowKeys={handleCheck}
            bordered={false}
          />

          {/* 分页栏 */}
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
