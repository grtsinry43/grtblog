import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NPagination,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
} from 'naive-ui'
import { defineComponent, reactive, ref } from 'vue'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { friendLinkService } from '@/services/friend-links'

import type {
  FriendLink,
  FriendLinkCreateReq,
  FriendLinkUpdateReq,
} from '@/types/friend-link'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
  name: 'FriendLinkList',
  setup() {
    const message = useMessage()

    const linksFilter = reactive({
      keyword: '',
    })

    const {
      data: links,
      loading: linksLoading,
      pagination: linksPagination,
      refresh: refreshLinks,
    } = useTable<FriendLink>(friendLinkService.getFriendLinks, linksFilter)

    const showEditModal = ref(false)
    const modalTitle = ref('新建友链')
    const editFormRef = ref<InstanceType<typeof NForm> | null>(null)
    const editingId = ref<number | null>(null)
    const formModel = reactive<FriendLinkCreateReq>({
      name: '',
      url: '',
      logo: '',
      description: '',
      rssUrl: '',
      kind: 'manual',
      syncMode: 'none',
      isActive: true,
      syncInterval: 60,
    })

    const rules = {
      name: { required: true, message: '请输入名称', trigger: 'blur' },
      url: { required: true, message: '请输入链接', trigger: 'blur' },
    }

    const handleSave = async () => {
      editFormRef.value?.validate(async (errors) => {
        if (!errors) {
          try {
            if (editingId.value) {
              await friendLinkService.updateFriendLink(editingId.value, formModel as FriendLinkUpdateReq)
              message.success('更新成功')
            }
            else {
              await friendLinkService.createFriendLink(formModel)
              message.success('创建成功')
            }
            showEditModal.value = false
            refreshLinks()
          }
          catch (e: any) {
            message.error(e.message || '保存失败')
          }
        }
      })
    }

    const handleAction = async (id: number, action: 'delete' | 'block') => {
      try {
        if (action === 'delete') {
          await friendLinkService.deleteFriendLink(id)
          message.success('删除成功')
        }
        else if (action === 'block') {
          await friendLinkService.blockFriendLink(id)
          message.success('封禁成功')
        }
        refreshLinks()
      }
      catch (e: any) {
        message.error(e.message || '操作失败')
      }
    }

    const openCreate = () => {
      editingId.value = null
      modalTitle.value = '新建友链'
      Object.assign(formModel, {
        name: '',
        url: '',
        logo: '',
        description: '',
        rssUrl: '',
        kind: 'manual',
        syncMode: 'none',
        isActive: true,
        syncInterval: 60,
      })
      showEditModal.value = true
    }

    const openEdit = (row: FriendLink) => {
      editingId.value = row.id
      modalTitle.value = '编辑友链'
      Object.assign(formModel, { ...row })
      showEditModal.value = true
    }

    const linkColumns: DataTableColumns<FriendLink> = [
      {
        title: 'Logo',
        key: 'logo',
        width: 60,
        render: (row) => {
          if (!row.logo)
            return null
          return <img src={row.logo} class="w-8 h-8 rounded object-cover" />
        },
      },
      {
        title: '名称',
        key: 'name',
        render: (row) => (
          <a
            href={row.url}
            target="_blank"
            class="text-primary hover:underline font-medium"
          >
            {row.name}
          </a>
        ),
      },
      {
        title: '类型',
        key: 'kind',
        width: 100,
        render: (row) => (
          <NTag type={row.kind === 'federation' ? 'info' : 'default'} size="small">
            {{ default: () => (row.kind === 'federation' ? '联邦' : '传统友链') }}
          </NTag>
        ),
      },
      {
        title: '同步',
        key: 'syncMode',
        width: 80,
        render: (row) => (
          <NTag size="small" bordered={false}>
            {{ default: () => row.syncMode }}
          </NTag>
        ),
      },
      {
        title: '状态',
        key: 'isActive',
        width: 80,
        render: (row) => (
          <NSwitch
            value={row.isActive}
            size="small"
            onUpdateValue={async (val: boolean) => {
              try {
                await friendLinkService.updateFriendLink(row.id, { ...row, isActive: val } as any)
                row.isActive = val
                message.success('已更新状态')
              }
              catch (e: any) {
                message.error(e.message || '更新状态失败')
              }
            }}
          />
        ),
      },
      {
        title: '操作',
        key: 'actions',
        width: 150,
        render: (row) => (
          <NSpace>
            <NButton
              size="tiny"
              secondary
              onClick={() => openEdit(row)}
            >
              编辑
            </NButton>
            <NPopconfirm
              onPositiveClick={() => handleAction(row.id, 'delete')}
            >
              {{
                default: () => '确认删除该友链吗？',
                trigger: () => (
                  <NButton size="tiny" type="error" secondary>
                    删除
                  </NButton>
                ),
              }}
            </NPopconfirm>
            <NPopconfirm
              onPositiveClick={() => handleAction(row.id, 'block')}
            >
              {{
                default: () => '确认封禁该友链吗？（后续申请将被自动拒绝）',
                trigger: () => (
                  <NButton size="tiny" type="error" secondary>
                    封禁
                  </NButton>
                ),
              }}
            </NPopconfirm>
          </NSpace>
        ),
      },
    ]

    return () => (
      <ScrollContainer wrapper-class="p-4">
        <NCard title="友链列表" class="h-full">
          {{
            'header-extra': () => (
              <NButton type="primary" size="small" onClick={openCreate}>
                新建友链
              </NButton>
            ),
            'default': () => (
              <>
                <div class="mb-4 flex gap-2">
                  <NInput
                    value={linksFilter.keyword}
                    placeholder="搜索名称或URL"
                    class="max-w-xs"
                    clearable
                    onUpdateValue={(v) => linksFilter.keyword = v}
                    onKeydown={(e) => {
                      if (e.key === 'Enter')
                        refreshLinks()
                    }}
                  />
                  <NButton secondary onClick={refreshLinks}>
                    搜索
                  </NButton>
                </div>
                <NDataTable
                  remote
                  columns={linkColumns}
                  data={links.value}
                  loading={linksLoading.value}
                  row-key={(row: FriendLink) => row.id}
                  scrollX={800}
                />
                <div class="mt-4 flex justify-end">
                  <NPagination
                    page={linksPagination.page}
                    page-size={linksPagination.pageSize}
                    item-count={linksPagination.itemCount}
                    show-size-picker={linksPagination.showSizePicker}
                    page-sizes={linksPagination.pageSizes}
                    onUpdatePage={linksPagination.onChange}
                    onUpdatePageSize={linksPagination.onUpdatePageSize}
                  />
                </div>
              </>
            ),
          }}
        </NCard>

        <NModal show={showEditModal.value} preset="card" title={modalTitle.value} class="max-w-lg" onUpdateShow={(v) => showEditModal.value = v}>
          {{
            default: () => (
              <NForm ref={editFormRef} model={formModel} rules={rules} label-placement="left" label-width="80">
                <NFormItem label="名称" path="name">
                  <NInput value={formModel.name} placeholder="站点名称" onUpdateValue={(v) => formModel.name = v} />
                </NFormItem>
                <NFormItem label="URL" path="url">
                  <NInput value={formModel.url} placeholder="https://example.com" onUpdateValue={(v) => formModel.url = v} />
                </NFormItem>
                <NFormItem label="Logo" path="logo">
                  <NInput value={formModel.logo} placeholder="Logo 图片地址" onUpdateValue={(v) => formModel.logo = v} />
                </NFormItem>
                <NFormItem label="描述" path="description">
                  <NInput value={formModel.description} type="textarea" placeholder="站点简介" onUpdateValue={(v) => formModel.description = v} />
                </NFormItem>
                <NFormItem label="RSS" path="rssUrl">
                  <NInput value={formModel.rssUrl} placeholder="RSS 订阅地址（可选）" onUpdateValue={(v) => formModel.rssUrl = v} />
                </NFormItem>

                <div class="grid grid-cols-2 gap-4">
                  <NFormItem label="类型" path="kind">
                    <NSelect
                      value={formModel.kind}
                      options={[
                        { label: '传统友链', value: 'manual' },
                        { label: '联邦', value: 'federation' },
                      ]}
                      onUpdateValue={(v) => formModel.kind = v}
                    />
                  </NFormItem>
                  <NFormItem label="同步" path="syncMode">
                    <NSelect
                      value={formModel.syncMode}
                      options={[
                        { label: '无', value: 'none' },
                        { label: 'RSS', value: 'rss' },
                        { label: '联邦', value: 'federation' },
                      ]}
                      onUpdateValue={(v) => formModel.syncMode = v}
                    />
                  </NFormItem>
                </div>

                {formModel.syncMode !== 'none' && (
                  <NFormItem label="刷新间隔" path="syncInterval">
                    <NInput
                      value={formModel.syncInterval?.toString()}
                      allowInput={(v) => !v || /^\d+$/.test(v)}
                      placeholder="单位：分钟"
                      onUpdateValue={(v) => formModel.syncInterval = v ? parseInt(v) : undefined}
                    >
                      {{ suffix: () => '分钟' }}
                    </NInput>
                  </NFormItem>
                )}

                <NFormItem label="启用" path="isActive">
                  <NSwitch value={formModel.isActive} onUpdateValue={(v) => formModel.isActive = v} />
                </NFormItem>
              </NForm>
            ),
            footer: () => (
              <div class="flex justify-end gap-2">
                <NButton onClick={() => showEditModal.value = false}>
                  取消
                </NButton>
                <NButton type="primary" loading={linksLoading.value} onClick={handleSave}>
                  保存
                </NButton>
              </div>
            ),
          }}
        </NModal>
      </ScrollContainer>
    )
  },
})
