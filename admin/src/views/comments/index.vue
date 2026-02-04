<script setup lang="ts">
import { h, ref, computed } from 'vue'
import {
  NCard,
  NTabs,
  NTabPane,
  NList,
  NListItem,
  NThing,
  NAvatar,
  NTag,
  NSpace,
  NButton,
  useMessage,
  useDialog,
  NResult,
  NSpin,
  NPopconfirm,
  NInput,
  NPopselect,
  NIcon,
  NEllipsis,
  NSwitch,
  NPagination,
} from 'naive-ui'
import { ScrollContainer, EmptyPlaceholder } from '@/components'
import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  listComments,
  deleteComment,
  updateCommentStatus,
  replyComment,
  setCommentTop,
  setCommentAuthor,
  markCommentsViewed,
} from '@/services/comments'
import { CommentStatus, type Comment } from '@/types/comments'
import CommentSource from './components/CommentSource.vue'
import {
  TrashOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
  ArrowUndoOutline,
  PinOutline,
  PersonOutline,
  GlobeOutline,
  LogoChrome,
  LaptopOutline,
  LocationOutline,
  MailOutline,
  BanOutline,
} from '@vicons/ionicons5'

const message = useMessage()
const dialog = useDialog()
const queryClient = useQueryClient()

const activeStatus = ref<string>('all')
const page = ref(1)
const pageSize = ref(20)
const onlyUnviewed = ref(false)

// Fetch comments
const { data, isLoading, isError } = useQuery({
  queryKey: ['comments', activeStatus, page, pageSize, onlyUnviewed],
  queryFn: () =>
    listComments({
      page: page.value,
      pageSize: pageSize.value,
      status: activeStatus.value === 'all' ? undefined : activeStatus.value,
      onlyUnviewed: onlyUnviewed.value,
    }),
})

const comments = computed(() => data.value?.items || [])
const total = computed(() => data.value?.total || 0)

// Mutations
const updateStatusMutation = useMutation({
  mutationFn: ({ id, status }: { id: number; status: CommentStatus }) =>
    updateCommentStatus(id, { status }),
  onSuccess: () => {
    message.success('状态更新成功')
    queryClient.invalidateQueries({ queryKey: ['comments'] })
  },
  onError: () => message.error('状态更新失败'),
})

const deleteMutation = useMutation({
  mutationFn: (id: number) => deleteComment(id),
  onSuccess: () => {
    message.success('评论已删除')
    queryClient.invalidateQueries({ queryKey: ['comments'] })
  },
  onError: () => message.error('删除失败'),
})

const topMutation = useMutation({
  mutationFn: ({ id, isTop }: { id: number; isTop: boolean }) => setCommentTop(id, { isTop }),
  onSuccess: () => {
    message.success('置顶状态已更新')
    queryClient.invalidateQueries({ queryKey: ['comments'] })
  },
})

const replyMutation = useMutation({
  mutationFn: ({ id, content }: { id: number; content: string }) => replyComment(id, { content }),
  onSuccess: () => {
    message.success('回复成功')
    replyContent.value = ''
    replyTargetId.value = null
    queryClient.invalidateQueries({ queryKey: ['comments'] })
  },
  onError: () => message.error('回复失败'),
})

const markViewedMutation = useMutation({
  mutationFn: (ids: number[]) => markCommentsViewed({ ids, isViewed: true }),
  onSuccess: (_, ids) => {
     queryClient.setQueryData(['comments', activeStatus.value, page.value, pageSize.value, onlyUnviewed.value], (oldData: any) => {
         if (!oldData) return oldData
         return {
             ...oldData,
             items: oldData.items.map((item: Comment) => {
                 if (ids.includes(item.id)) {
                     return { ...item, isViewed: true }
                 }
                 return item
             })
         }
     })
  }
})

// Actions
const handleStatusChange = (comment: Comment, status: CommentStatus) => {
  updateStatusMutation.mutate({ id: comment.id, status })
}

const handleDelete = (comment: Comment) => {
  deleteMutation.mutate(comment.id)
}

const handleTop = (comment: Comment) => {
  topMutation.mutate({ id: comment.id, isTop: !comment.isTop })
}

// Reply Logic
const replyTargetId = ref<number | null>(null)
const replyContent = ref('')
const showReplyInput = (comment: Comment) => {
  if (replyTargetId.value === comment.id) {
    replyTargetId.value = null
  } else {
    replyTargetId.value = comment.id
    replyContent.value = ''
  }
}
const submitReply = () => {
  if (!replyTargetId.value || !replyContent.value) return
  replyMutation.mutate({ id: replyTargetId.value, content: replyContent.value })
}

// Utility
const getStatusType = (status: CommentStatus) => {
  switch (status) {
    case CommentStatus.Approved: return 'success'
    case CommentStatus.Pending: return 'warning'
    case CommentStatus.Rejected: return 'error'
    case CommentStatus.Blocked: return 'default'
    default: return 'default'
  }
}

const getStatusLabel = (status: CommentStatus) => {
    switch (status) {
    case CommentStatus.Approved: return '已发布'
    case CommentStatus.Pending: return '待审核'
    case CommentStatus.Rejected: return '已拒绝'
    case CommentStatus.Blocked: return '已封禁'
    default: return '未知'
  }
}

const formatUserAgent = (browser?: string, platform?: string) => {
  return [browser, platform].filter(Boolean).join(' · ')
}

const hoverTimer = ref<ReturnType<typeof setTimeout> | null>(null)

const handleMouseEnter = (comment: Comment) => {
    if (!comment.isViewed) {
        hoverTimer.value = setTimeout(() => {
            markViewedMutation.mutate([comment.id])
        }, 500)
    }
}

const handleMouseLeave = () => {
    if (hoverTimer.value) {
        clearTimeout(hoverTimer.value)
        hoverTimer.value = null
    }
}
</script>

<template>
  <ScrollContainer wrapper-class="p-4" :scrollbar-props="{ trigger: 'none' }">
    <n-card title="评论管理" class="h-full" content-style="display: flex; flex-direction: column; height: 100%;">
       <template #header-extra>
            <!-- Using header-extra for the switch if preferred, or keep in body. 
                 User said "Header and below separate", "Tab width wrap content".
                 Let's put Switch in the body row with Tabs as before but cleaner.
            -->
       </template>
       
      <div class="flex-1 min-h-0 flex flex-col relative">
        <div class="px-4 border-b border-gray-100 dark:border-gray-800">
             <n-tabs v-model:value="activeStatus" type="line" animated>
                <n-tab-pane name="all" tab="全部" />
                <n-tab-pane name="pending" tab="待审核" />
                <n-tab-pane name="approved" tab="已发布" />
                <n-tab-pane name="rejected" tab="垃圾/拒绝" />
                
                <template #suffix>
                    <n-space align="center" size="small" class="pb-1">
                        <span class="text-xs text-gray-500">仅看未读</span>
                        <n-switch v-model:value="onlyUnviewed" size="small" />
                    </n-space>
                </template>
            </n-tabs>
        </div>

        <div class="flex-1 relative min-h-0">
             <n-spin :show="isLoading" class="h-full flex flex-col">
                <EmptyPlaceholder :show="!isLoading && comments.length === 0" description="暂无评论" />
                
                <div class="flex-1 overflow-auto px-4 py-4" v-if="comments.length > 0">
                    <n-list hoverable clickable bordered>
                        <n-list-item 
                            v-for="comment in comments" 
                            :key="comment.id"
                            @mouseenter="handleMouseEnter(comment)"
                            @mouseleave="handleMouseLeave"
                            :class="{ 'bg-blue-50/50 dark:bg-blue-900/10': !comment.isViewed }"
                            class="transition-colors duration-200"
                        >
                            <n-thing content-indented>
                                 <template #avatar>
                                    <div class="relative">
                                        <n-avatar round :src="`https://cravatar.cn/avatar/${comment.email ? comment.email : 'default'}?d=mp`" />
                                        <div v-if="!comment.isViewed" class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-white dark:border-gray-800"></div>
                                    </div>
                                </template>
                                <template #header>
                                    <span class="font-bold mr-2 text-base" :class="{ 'line-through text-gray-400': comment.isDeleted }">{{ comment.nickName }}</span>
                                    <n-tag v-if="comment.isDeleted" type="error" size="small" class="mr-1">已删除</n-tag>
                                    <n-tag v-if="comment.isTop" type="warning" size="small" class="mr-1">
                                        置顶
                                    </n-tag>
                                    <n-tag v-if="comment.isOwner" type="primary" size="small" class="mr-1">站长</n-tag>
                                    <n-tag v-if="comment.isFriend" type="success" size="small">友链</n-tag>
                                </template>
                                <template #header-extra>
                                    <div class="flex items-center gap-2 text-xs text-gray-500">
                                        <n-tag :type="getStatusType(comment.status)" :bordered="false" size="small">{{ getStatusLabel(comment.status) }}</n-tag>
                                        <span>{{ new Date(comment.createdAt).toLocaleString() }}</span>
                                    </div>
                                </template>
                                <template #description>
                                    <div class="flex flex-wrap gap-3 text-xs text-gray-400 items-center mt-1">
                                        <span v-if="comment.ip" class="flex items-center gap-1"><n-icon :component="GlobeOutline" /> {{ comment.ip }}</span>
                                        <span v-if="comment.location" class="flex items-center gap-1"><n-icon :component="LocationOutline" /> {{ comment.location }}</span>
                                        <span v-if="comment.email" class="flex items-center gap-1"><n-icon :component="MailOutline" /> {{ comment.email }}</span>
                                        <span v-if="comment.browser || comment.platform" class="flex items-center gap-1"><n-icon :component="LaptopOutline" /> {{ formatUserAgent(comment.browser, comment.platform) }}</span>
                                    </div>
                                </template>
                                
                                <div class="py-3 text-sm whitespace-pre-wrap leading-relaxed" :class="{ 'opacity-50': comment.isDeleted }">
                                    {{ comment.content }}
                                </div>

                                 <div class="bg-gray-50 dark:bg-gray-800/50 rounded-md p-2 mb-2 text-xs flex items-center justify-between border border-gray-100 dark:border-gray-800">
                                    <comment-source 
                                        :type="comment.areaType" 
                                        :id="comment.areaRefId" 
                                        :initial-title="comment.areaTitle || comment.areaName" 
                                    />
                                 </div>

                                <template #action v-if="!comment.isDeleted">
                                     <n-space size="small" class="mt-2">
                                        <n-button text size="tiny" @click="showReplyInput(comment)">
                                            <template #icon><n-icon :component="ArrowUndoOutline" /></template>
                                            回复
                                        </n-button>
                                        
                                        <n-popconfirm v-if="comment.status !== CommentStatus.Approved" @positive-click="handleStatusChange(comment, CommentStatus.Approved)">
                                            <template #trigger>
                                                <n-button text type="success" size="tiny">
                                                    <template #icon><n-icon :component="CheckmarkCircleOutline" /></template>
                                                    通过
                                                </n-button>
                                            </template>
                                            确定通过这条评论吗？
                                        </n-popconfirm>

                                        <n-popconfirm v-if="comment.status !== CommentStatus.Blocked" @positive-click="handleStatusChange(comment, CommentStatus.Blocked)">
                                            <template #trigger>
                                                <n-button text type="error" size="tiny">
                                                    <template #icon><n-icon :component="BanOutline" /></template>
                                                    封禁
                                                </n-button>
                                            </template>
                                            确定封禁该评论的作者吗？封禁后将不再接受同一用户或邮箱地址的后续评论
                                        </n-popconfirm>

                                        <n-popconfirm v-if="comment.status !== CommentStatus.Rejected" @positive-click="handleStatusChange(comment, CommentStatus.Rejected)">
                                            <template #trigger>
                                                <n-button text type="warning" size="tiny">
                                                    <template #icon><n-icon :component="CloseCircleOutline" /></template>
                                                    拒绝
                                                </n-button>
                                            </template>
                                            确定拒绝这条评论吗？
                                        </n-popconfirm>

                                        <n-button text :type="comment.isTop ? 'primary' : 'default'" size="tiny" @click="handleTop(comment)">
                                             <template #icon><n-icon :component="PinOutline" /></template>
                                             {{ comment.isTop ? '取消置顶' : '置顶' }}
                                        </n-button>

                                        <n-popconfirm @positive-click="handleDelete(comment)">
                                            <template #trigger>
                                                <n-button text type="error" size="tiny">
                                                    <template #icon><n-icon :component="TrashOutline" /></template>
                                                    删除
                                                </n-button>
                                            </template>
                                            确定删除这条评论吗？此操作不可恢复。
                                        </n-popconfirm>
                                     </n-space>
                                     
                                     <div v-if="replyTargetId === comment.id" class="mt-3 flex gap-2">
                                        <n-input v-model:value="replyContent" type="textarea" placeholder="输入回复内容..." :rows="2" autosize />
                                        <n-button type="primary" @click="submitReply" :loading="replyMutation.isPending.value">发送</n-button>
                                     </div>
                                </template>
                            </n-thing>
                        </n-list-item>
                    </n-list>

                    <div class="flex justify-end mt-4 pb-2" v-if="total > 0">
                         <n-pagination
                            v-model:page="page"
                            v-model:page-size="pageSize"
                            :item-count="total"
                            show-size-picker
                            :page-sizes="[10, 20, 50, 100]"
                         />
                    </div>
                </div>
             </n-spin>
        </div>
      </div>
    </n-card>
  </ScrollContainer>
</template>
