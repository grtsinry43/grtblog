<script setup lang="ts">
import { CloudArrowUp24Regular } from '@vicons/fluent'
import {
  NCard,
  NEmpty,
  NIcon,
  NImage,
  NModal,
  NPagination,
  NTree,
  NUpload,
  useMessage,
} from 'naive-ui'

import { ScrollContainer } from '@/components'
import { useFileList } from './composables/use-file-list'
import FileTable from './components/FileTable.vue'
import FileUploader from './components/FileUploader.vue'
import RenameModal from './components/RenameModal.vue'

const message = useMessage()

const {
  loading,
  uploading,
  page,
  pageSize,
  total,
  uploadType,
  treeSelectedKeys,
  renameModalVisible,
  newFileName,
  deleteModalVisible,
  deletingFile,
  previewVisible,
  previewImageUrl,
  isEmpty,
  filteredFiles,
  treeData,
  handleUpload,
  handleCopyUrl,
  openRenameModal,
  handleRename,
  openDeleteModal,
  handleDelete,
  handleDownload,
  handlePageChange,
  handlePageSizeChange,
  openPreview,
  handleTreeSelect,
} = useFileList(message)
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <NCard title="文件管理" :bordered="false">
      <template #header-extra>
        <FileUploader
          :upload-type="uploadType"
          :uploading="uploading"
          @update:upload-type="uploadType = $event"
          @upload="handleUpload"
        />
      </template>

      <div class="upload-area">
        <NUpload
          :show-file-list="false"
          :custom-request="handleUpload"
          :disabled="uploading"
          directory-dnd
        >
          <div class="upload-dragger">
            <div class="upload-icon">
              <NIcon size="48" :depth="3"><CloudArrowUp24Regular /></NIcon>
            </div>
            <div class="upload-text">
              <p class="upload-hint">点击或拖拽文件到此区域上传</p>
              <p class="upload-description">
                当前类型：{{ uploadType === 'picture' ? '图片' : '文件' }}
              </p>
            </div>
          </div>
        </NUpload>
      </div>

      <div v-if="isEmpty" class="empty-container">
        <NEmpty description="暂无文件" />
      </div>

      <div v-else class="content-layout">
        <div class="tree-panel">
          <div class="tree-title">文件树</div>
          <NTree
            :data="treeData"
            :selected-keys="treeSelectedKeys"
            block-line
            default-expand-all
            @update:selected-keys="handleTreeSelect"
          />
        </div>

        <div class="table-panel">
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
      </div>
    </NCard>

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

.upload-area {
  margin-bottom: 24px;
}

.upload-dragger {
  padding: 40px 20px;
  background: transparent;
  border: 2px dashed var(--n-border-color);
  border-radius: 8px;
  text-align: center;
  transition: all 0.3s ease;
  cursor: pointer;
}

.upload-dragger:hover {
  border-color: var(--n-border-color);
  background: rgba(0, 0, 0, 0.03);
}

.upload-icon {
  margin-bottom: 12px;
  color: var(--n-text-color-disabled);
}

.upload-hint {
  font-size: 16px;
  color: var(--n-text-color);
  margin: 0 0 8px;
}

.upload-description {
  font-size: 14px;
  color: var(--n-text-color-disabled);
  margin: 0;
}

.empty-container {
  padding: 60px 0;
}

.content-layout {
  display: flex;
  gap: 16px;
}

.tree-panel {
  width: 220px;
  min-width: 200px;
  padding: 12px;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  background: transparent;
  height: fit-content;
}

.tree-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--n-text-color);
  margin-bottom: 8px;
}

.table-panel {
  flex: 1;
  min-width: 0;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 16px 0;
}

.preview-container {
  display: flex;
  justify-content: center;
  align-items: center;
}

@media (max-width: 900px) {
  .content-layout {
    flex-direction: column;
  }

  .tree-panel {
    width: 100%;
  }
}
</style>
