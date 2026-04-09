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
  FriendLinkFederationRequestReq,
  FriendLinkUpdateReq,
} from '@/types/friend-link'
import type { DataTableColumns } from 'naive-ui'

const typeLabelMap: Record<FriendLink['type'], string> = {
  federation: '联合',
  rss: 'RSS',
  norss: '无 RSS',
}

const typeTagMap: Record<FriendLink['type'], 'default' | 'info' | 'success'> = {
  federation: 'info',
  rss: 'success',
  norss: 'default',
}

export default defineComponent({
  name: 'FriendLinkList',
  setup() {
    const message = useMessage()

    const linksFilter = reactive({
      keyword: '',
      type: undefined as FriendLink['type'] | undefined,
    })

    const {
      data: links,
      loading: linksLoading,
      pagination: linksPagination,
      refresh: refreshLinks,
    } = useTable<FriendLink>(friendLinkService.getFriendLinks, linksFilter as any)

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
      type: 'norss',
      instanceId: undefined,
      isActive: true,
      syncInterval: 60,
    })

    const showRequestModal = ref(false)
    const requestForm = reactive<FriendLinkFederationRequestReq>({
      target_url: '',
      message: '',
      rss_url: '',
    })
    const requestLoading = ref(false)

    const rules = {
      name: { required: true, message: '请输入名称', trigger: 'blur' },
      url: { required: true, message: '请输入链接', trigger: 'blur' },
    }

    const normalizeOptional = (value?: string) => {
      const trimmed = (value || '').trim()
      return trimmed || undefined
    }

    const buildPayload = (): FriendLinkCreateReq => {
      const payload: FriendLinkCreateReq = {
        name: formModel.name.trim(),
        url: formModel.url.trim(),
        logo: normalizeOptional(formModel.logo),
        description: normalizeOptional(formModel.description),
        rssUrl: normalizeOptional(formModel.rssUrl),
        type: formModel.type,
        instanceId: formModel.instanceId,
        syncInterval:
          formModel.syncInterval && formModel.syncInterval > 0 ? formModel.syncInterval : undefined,
        isActive: formModel.isActive,
      }
      if (payload.type !== 'federation') {
        payload.instanceId = undefined
      }
      return payload
    }

    const toUpdatePayload = (row: FriendLink, isActive = row.isActive): FriendLinkUpdateReq => ({
      name: row.name,
      url: row.url,
      logo: row.logo,
      description: row.description,
      rssUrl: row.rssUrl,
      type: row.type,
      instanceId: row.instanceId,
      syncInterval: row.syncInterval,
      isActive,
    })

    const resetRequestForm = () => {
      requestForm.target_url = ''
      requestForm.message = ''
      requestForm.rss_url = ''
    }

    const handleSave = async () => {
      editFormRef.value?.validate(async (errors) => {
        if (!errors) {
          try {
            const payload = buildPayload()
            if (editingId.value) {
              await friendLinkService.updateFriendLink(editingId.value, payload)
              message.success('更新成功')
            } else {
              await friendLinkService.createFriendLink(payload)
              message.success('创建成功')
            }
            showEditModal.value = false
            refreshLinks()
          } catch (e: any) {
            message.error(e.message || '保存失败')
          }
        }
      })
    }

    const handleFederationRequest = async () => {
      const targetURL = requestForm.target_url.trim()
      if (!targetURL) {
        message.warning('请输入目标地址')
        return
      }
      try {
        requestLoading.value = true
        await friendLinkService.requestFederationFriendLink({
          target_url: targetURL,
          message: normalizeOptional(requestForm.message),
          rss_url: normalizeOptional(requestForm.rss_url),
        })
        message.success('申请已发送')
        showRequestModal.value = false
        resetRequestForm()
      } catch (e: any) {
        message.error(e.message || '发送失败')
      } finally {
        requestLoading.value = false
      }
    }

    const handleAction = async (id: number, action: 'delete' | 'block') => {
      try {
        if (action === 'delete') {
          await friendLinkService.deleteFriendLink(id)
          message.success('删除成功')
        } else if (action === 'block') {
          await friendLinkService.blockFriendLink(id)
          message.success('封禁成功')
        }
        refreshLinks()
      } catch (e: any) {
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
        type: 'norss',
        instanceId: undefined,
        isActive: true,
        syncInterval: 60,
      } satisfies FriendLinkCreateReq)
      showEditModal.value = true
    }

    const openEdit = (row: FriendLink) => {
      editingId.value = row.id
      modalTitle.value = '编辑友链'
      Object.assign(formModel, {
        name: row.name,
        url: row.url,
        logo: row.logo || '',
        description: row.description || '',
        rssUrl: row.rssUrl || '',
        type: row.type,
        instanceId: row.instanceId,
        isActive: row.isActive,
        syncInterval: row.syncInterval,
      } satisfies FriendLinkCreateReq)
      showEditModal.value = true
    }

    const linkColumns: DataTableColumns<FriendLink> = [
      {
        title: 'Logo',
        key: 'logo',
        width: 60,
        render: (row) => {
          if (!row.logo) return null
          return (
            <img
              src={row.logo}
              class='h-8 w-8 rounded object-cover'
            />
          )
        },
      },
      {
        title: '名称',
        key: 'name',
        render: (row) => (
          <a
            href={row.url}
            target='_blank'
            class='font-medium text-primary hover:underline'
          >
            {row.name}
          </a>
        ),
      },
      {
        title: '类型',
        key: 'type',
        width: 110,
        render: (row) => (
          <NTag
            type={typeTagMap[row.type]}
            size='small'
          >
            {{ default: () => typeLabelMap[row.type] }}
          </NTag>
        ),
      },
      {
        title: '同步来源',
        key: 'syncSource',
        minWidth: 180,
        render: (row) => (
          <span class='truncate text-sm text-neutral-500'>
            {row.type === 'federation' ? `实例 #${row.instanceId || '-'}` : row.rssUrl || '-'}
          </span>
        ),
      },
      {
        title: '状态',
        key: 'isActive',
        width: 80,
        render: (row) => (
          <NSwitch
            value={row.isActive}
            size='small'
            onUpdateValue={async (val: boolean) => {
              try {
                await friendLinkService.updateFriendLink(row.id, toUpdatePayload(row, val))
                row.isActive = val
                message.success('已更新状态')
              } catch (e: any) {
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
              size='tiny'
              secondary
              onClick={() => openEdit(row)}
            >
              编辑
            </NButton>
            <NPopconfirm onPositiveClick={() => handleAction(row.id, 'delete')}>
              {{
                default: () => '确认删除该友链吗？',
                trigger: () => (
                  <NButton
                    size='tiny'
                    type='error'
                    secondary
                  >
                    删除
                  </NButton>
                ),
              }}
            </NPopconfirm>
            <NPopconfirm onPositiveClick={() => handleAction(row.id, 'block')}>
              {{
                default: () => '确认封禁该友链吗？（后续申请将被自动拒绝）',
                trigger: () => (
                  <NButton
                    size='tiny'
                    type='error'
                    secondary
                  >
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
      <ScrollContainer wrapper-class='p-4'>
        <NCard
          title='友链列表'
          class='h-full'
        >
          {{
            'header-extra': () => (
              <NSpace>
                <NButton
                  secondary
                  size='small'
                  onClick={() => (showRequestModal.value = true)}
                >
                  发起联合友链申请
                </NButton>
                <NButton
                  type='primary'
                  size='small'
                  onClick={openCreate}
                >
                  新建友链
                </NButton>
              </NSpace>
            ),
            default: () => (
              <>
                <div class='mb-4 flex gap-2'>
                  <NInput
                    value={linksFilter.keyword}
                    placeholder='搜索名称或URL'
                    class='max-w-xs'
                    clearable
                    onUpdateValue={(v) => (linksFilter.keyword = v)}
                    onKeydown={(e) => {
                      if (e.key === 'Enter') refreshLinks()
                    }}
                  />
                  <NSelect
                    value={linksFilter.type}
                    clearable
                    class='w-40'
                    placeholder='类型'
                    options={[
                      { label: '联合', value: 'federation' },
                      { label: 'RSS', value: 'rss' },
                      { label: '无 RSS', value: 'norss' },
                    ]}
                    onUpdateValue={(v) => (linksFilter.type = v as FriendLink['type'] | undefined)}
                  />
                  <NButton
                    secondary
                    onClick={refreshLinks}
                  >
                    搜索
                  </NButton>
                </div>
                <NDataTable
                  remote
                  columns={linkColumns}
                  data={links.value}
                  loading={linksLoading.value}
                  row-key={(row: FriendLink) => row.id}
                  scrollX={860}
                />
                <div class='mt-4 flex justify-end'>
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

        <NModal
          show={showEditModal.value}
          preset='card'
          title={modalTitle.value}
          class='max-w-lg'
          onUpdateShow={(v) => (showEditModal.value = v)}
        >
          {{
            default: () => (
              <NForm
                ref={editFormRef}
                model={formModel}
                rules={rules}
                label-placement='left'
                label-width='80'
              >
                <NFormItem
                  label='名称'
                  path='name'
                >
                  <NInput
                    value={formModel.name}
                    placeholder='站点名称'
                    onUpdateValue={(v) => (formModel.name = v)}
                  />
                </NFormItem>
                <NFormItem
                  label='URL'
                  path='url'
                >
                  <NInput
                    value={formModel.url}
                    placeholder='https://example.com'
                    onUpdateValue={(v) => (formModel.url = v)}
                  />
                </NFormItem>
                <NFormItem
                  label='Logo'
                  path='logo'
                >
                  <NInput
                    value={formModel.logo}
                    placeholder='Logo 图片地址'
                    onUpdateValue={(v) => (formModel.logo = v)}
                  />
                </NFormItem>
                <NFormItem
                  label='描述'
                  path='description'
                >
                  <NInput
                    value={formModel.description}
                    type='textarea'
                    placeholder='站点简介'
                    onUpdateValue={(v) => (formModel.description = v)}
                  />
                </NFormItem>
                <NFormItem
                  label='类型'
                  path='type'
                >
                  <NSelect
                    value={formModel.type}
                    options={[
                      { label: '联合', value: 'federation' },
                      { label: 'RSS', value: 'rss' },
                      { label: '无 RSS', value: 'norss' },
                    ]}
                    onUpdateValue={(v) => {
                      formModel.type = v as FriendLink['type']
                      if (formModel.type !== 'federation') formModel.instanceId = undefined
                    }}
                  />
                </NFormItem>
                <NFormItem
                  label='RSS'
                  path='rssUrl'
                >
                  <NInput
                    value={formModel.rssUrl}
                    placeholder={
                      formModel.type === 'rss'
                        ? 'RSS 必填，例如 https://example.com/feed'
                        : 'RSS 订阅地址（可选）'
                    }
                    onUpdateValue={(v) => (formModel.rssUrl = v)}
                  />
                </NFormItem>

                {formModel.type === 'federation' && (
                  <NFormItem
                    label='实例ID'
                    path='instanceId'
                  >
                    <NInput
                      value={formModel.instanceId?.toString() || ''}
                      allowInput={(v) => !v || /^\d+$/.test(v)}
                      placeholder='联合实例 ID（必填）'
                      onUpdateValue={(v) =>
                        (formModel.instanceId = v ? Number.parseInt(v, 10) : undefined)
                      }
                    />
                  </NFormItem>
                )}

                {formModel.type !== 'norss' && (
                  <NFormItem
                    label='刷新间隔'
                    path='syncInterval'
                  >
                    <NInput
                      value={formModel.syncInterval?.toString() || ''}
                      allowInput={(v) => !v || /^\d+$/.test(v)}
                      placeholder='单位：分钟'
                      onUpdateValue={(v) =>
                        (formModel.syncInterval = v ? Number.parseInt(v, 10) : undefined)
                      }
                    >
                      {{ suffix: () => '分钟' }}
                    </NInput>
                  </NFormItem>
                )}

                <NFormItem
                  label='启用'
                  path='isActive'
                >
                  <NSwitch
                    value={formModel.isActive}
                    onUpdateValue={(v) => (formModel.isActive = v)}
                  />
                </NFormItem>
              </NForm>
            ),
            footer: () => (
              <div class='flex justify-end gap-2'>
                <NButton onClick={() => (showEditModal.value = false)}>取消</NButton>
                <NButton
                  type='primary'
                  loading={linksLoading.value}
                  onClick={handleSave}
                >
                  保存
                </NButton>
              </div>
            ),
          }}
        </NModal>

        <NModal
          show={showRequestModal.value}
          preset='card'
          title='发起联合友链申请'
          class='max-w-lg'
          onUpdateShow={(v) => (showRequestModal.value = v)}
        >
          {{
            default: () => (
              <NForm
                label-placement='left'
                label-width='100'
              >
                <NFormItem
                  label='目标地址'
                  required
                >
                  <NInput
                    value={requestForm.target_url}
                    placeholder='https://target.example.com'
                    onUpdateValue={(v) => (requestForm.target_url = v)}
                  />
                </NFormItem>
                <NFormItem label='本站 RSS'>
                  <NInput
                    value={requestForm.rss_url}
                    placeholder='https://your-site.com/feed（可选）'
                    onUpdateValue={(v) => (requestForm.rss_url = v)}
                  />
                </NFormItem>
                <NFormItem label='留言'>
                  <NInput
                    value={requestForm.message}
                    type='textarea'
                    placeholder='你好，想与贵站建立友链与联合关系'
                    onUpdateValue={(v) => (requestForm.message = v)}
                  />
                </NFormItem>
              </NForm>
            ),
            footer: () => (
              <div class='flex justify-end gap-2'>
                <NButton onClick={() => (showRequestModal.value = false)}>取消</NButton>
                <NButton
                  type='primary'
                  loading={requestLoading.value}
                  onClick={handleFederationRequest}
                >
                  发送申请
                </NButton>
              </div>
            ),
          }}
        </NModal>
      </ScrollContainer>
    )
  },
})
