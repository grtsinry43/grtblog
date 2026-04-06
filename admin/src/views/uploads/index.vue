<script setup lang="ts">
import { NButton, NCard, NEmpty, NImage, NModal, NPagination, NProgress, NTag, NTabs, NTabPane, useMessage } from 'naive-ui'

import { ScrollContainer } from '@/components'
import { formatFileSize } from '@/utils/format'
import { useFileList } from './composables/use-file-list'
import FileTable from './components/FileTable.vue'
import FileUploader from './components/FileUploader.vue'
import RenameModal from './components/RenameModal.vue'

const message = useMessage()

const {
  loading,
  uploading,
  syncing,
  page,
  pageSize,
  total,
  uploadType,
  activeFilter,
  uploadTasks,
  renameModalVisible,
  newFileName,
  deleteModalVisible,
  deletingFile,
  previewVisible,
  previewImageUrl,
  filteredFiles,
  isEmpty,
  handleUpload,
  handleSync,
  handleCopyUrl,
  openRenameModal,
  handleRename,
  openDeleteModal,
  handleDelete,
  handleDownload,
  handlePageChange,
  handlePageSizeChange,
  openPreview,
} = useFileList(message)
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <div class="page-layout">
      <NCard :bordered="false">
        <div class="header-row">
          <div class="header-main">
            <div class="page-title">文件管理</div>
            <NTabs v-model:value="activeFilter" size="small" animated>
              <NTabPane name="all" tab="全部" />
              <NTabPane name="picture" tab="图片" />
              <NTabPane name="file" tab="文件" />
            </NTabs>
          </div>
          <div class="header-actions">
            <NButton secondary :loading="syncing" @click="handleSync">
              同步索引
            </NButton>
            <FileUploader
              :upload-type="uploadType"
              :uploading="uploading"
              @update:upload-type="uploadType = $event"
              @upload="handleUpload"
            />
          </div>
        </div>
      </NCard>

      <NCard v-if="isEmpty" :bordered="false">
        <div class="empty-container">
          <NEmpty description="暂无文件" />
        </div>
      </NCard>

      <NCard v-if="uploadTasks.length > 0" :bordered="false" title="上传进度">
        <div class="upload-task-list">
          <div v-for="task in uploadTasks" :key="task.id" class="upload-task-item">
            <div class="upload-task-meta">
              <div class="upload-task-main">
                <span class="upload-task-name">{{ task.name }}</span>
                <NTag size="small" :type="task.type === 'picture' ? 'success' : 'info'" :bordered="false">
                  {{ task.type === 'picture' ? '图片' : '文件' }}
                </NTag>
                <NTag
                  size="small"
                  :type="task.status === 'success' ? 'success' : task.status === 'error' ? 'error' : 'warning'"
                  :bordered="false"
                >
                  {{ task.status === 'success' ? '完成' : task.status === 'error' ? '失败' : '上传中' }}
                </NTag>
              </div>
              <span class="upload-task-size">{{ formatFileSize(task.size) }}</span>
            </div>
            <NProgress
              type="line"
              :percentage="task.percentage"
              :status="task.status === 'error' ? 'error' : task.status === 'success' ? 'success' : 'info'"
              :show-indicator="true"
              :processing="task.status === 'uploading'"
            />
            <div v-if="task.error" class="upload-task-error">{{ task.error }}</div>
          </div>
        </div>
      </NCard>

      <NCard v-else :bordered="false" content-style="padding: 0;">
        <div class="table-card-body">
          <FileTable
            :files="filteredFiles"
            :loading="loading"
            @copy-url="handleCopyUrl"
            @rename="openRenameModal"
            @download="handleDownload"
            @delete="openDeleteModal"
            @preview="openPreview"
          />

          <div class="pagination-container">
            <NPagination
              v-model:page="page"
              v-model:page-size="pageSize"
              :page-count="Math.ceil(total / pageSize)"
              :page-sizes="[10, 20, 50, 100]"
              show-size-picker
              @update:page="handlePageChange"
              @update:page-size="handlePageSizeChange"
            />
          </div>
        </div>
      </NCard>
    </div>

    <RenameModal
      :visible="renameModalVisible"
      :file-name="newFileName"
      @update:visible="renameModalVisible = $event"
      @update:file-name="newFileName = $event"
      @confirm="handleRename"
    />

    <NModal
      v-model:show="deleteModalVisible"
      preset="dialog"
      title="确认删除"
      type="warning"
      positive-text="删除"
      negative-text="取消"
      @positive-click="handleDelete"
    >
      <p>确定要删除文件 "{{ deletingFile?.name }}" 吗？</p>
      <p style="color: #f5222d; margin-top: 8px">此操作将永久删除文件，无法恢复。</p>
    </NModal>

    <NModal v-model:show="previewVisible" preset="card" style="max-width: 800px">
      <template #header><span>图片预览</span></template>
      <div class="preview-container">
        <NImage :src="previewImageUrl" />
      </div>
    </NModal>
  </ScrollContainer>
</template>

<style scoped>
.page-layout {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.header-main {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.page-title {
  font-size: 18px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-container {
  padding: 60px 0;
}

.upload-task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.upload-task-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-task-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.upload-task-main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.upload-task-name {
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.upload-task-size {
  flex-shrink: 0;
  color: var(--n-text-color-3);
  font-size: 12px;
}

.upload-task-error {
  color: var(--n-error-color);
  font-size: 12px;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 16px 0;
}

.table-card-body {
  padding: 0 16px;
}

.preview-container {
  display: flex;
  justify-content: center;
  align-items: center;
}

@media (max-width: 900px) {
  .header-row {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    flex-wrap: wrap;
  }

  .upload-task-meta {
    flex-direction: column;
    align-items: flex-start;
  }

  .upload-task-name {
    max-width: 100%;
  }
}
</style>
