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
    NTabPane,
    NTabs,
    NTag,
    NThing,
    NDropdown,
    useMessage,
} from 'naive-ui'
import { defineComponent, onMounted, reactive, ref } from 'vue'

import { ScrollContainer } from '@/components'
import { useTable } from '@/composables/table/use-table'
import { friendLinkService } from '@/services/friend-links'

import type {
    FriendLink,
    FriendLinkApplication,
    FriendLinkCreateReq,
    FriendLinkUpdateReq,
} from '@/types/friend-link'
import type { DataTableColumns } from 'naive-ui'

export default defineComponent({
    name: 'FriendLinkList',
    setup() {
        const message = useMessage()

        // --- State ---
        const currentTab = ref('list')

        // Link List Data
        const linksFilter = reactive({
            keyword: '',
        })

        // Application List Data
        const appsFilter = reactive({
            status: undefined as string | undefined,
        })

        // Use useTable for both lists
        const {
            data: links,
            loading: linksLoading,
            pagination: linksPagination,
            refresh: refreshLinks,
        } = useTable<FriendLink>(friendLinkService.getFriendLinks, linksFilter)

        const {
            data: apps,
            loading: appsLoading,
            pagination: appsPagination,
            refresh: refreshApps,
        } = useTable<FriendLinkApplication>(friendLinkService.getApplications, appsFilter as any)

        // Modal & Form
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

        // --- Actions ---

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

        const handleAppStatusUpdate = async (id: number, status: string) => {
            try {
                await friendLinkService.updateApplicationStatus(id, status)
                message.success('状态变更成功')
                refreshApps()
                // If approved, maybe refresh links too?
                if (status === 'approved')
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

        // --- Columns ---

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
                render: (row) => {
                    return (
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
                    )
                },
            },
        ]

        return () => (
            <ScrollContainer wrapper-class="p-4">
                <NCard title="友链管理" class="h-full">
                    {{
                        'header-extra': () => (
                            <NButton type="primary" size="small" onClick={openCreate}>
                                新建友链
                            </NButton>
                        ),
                        'default': () => (
                            <NTabs
                                value={currentTab.value}
                                type="line"
                                animated
                                onUpdateValue={(val: string) => {
                                    currentTab.value = val
                                    if (val === 'list')
                                        refreshLinks()
                                    else refreshApps()
                                }}
                            >
                                <NTabPane name="list" tab="友链列表">
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
                                </NTabPane>

                                <NTabPane name="applications" tab="申请审核">
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
                                </NTabPane>
                            </NTabs>
                        ),
                    }}
                </NCard>

                {/* Create/Edit Modal */}
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
