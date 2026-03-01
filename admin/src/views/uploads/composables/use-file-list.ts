import { computed, onMounted, ref } from 'vue'
import { deleteFile, downloadFile, listUploads, renameFile, uploadFile } from '@/services/uploads'

import type { FileType, UploadFileResponse } from '@/services/uploads'
import type { UploadFileInfo } from 'naive-ui'

export function useFileList(message: { error: (m: string) => void; success: (m: string) => void; warning: (m: string) => void }) {
  const files = ref<UploadFileResponse[]>([])
  const loading = ref(false)
  const uploading = ref(false)
  const page = ref(1)
  const pageSize = ref(10)
  const total = ref(0)
  const uploadType = ref<FileType>('picture')
  const treeSelectedKeys = ref<string[]>(['all'])

  // Rename
  const renameModalVisible = ref(false)
  const renamingFile = ref<UploadFileResponse | null>(null)
  const newFileName = ref('')

  // Delete
  const deleteModalVisible = ref(false)
  const deletingFile = ref<UploadFileResponse | null>(null)

  // Preview
  const previewVisible = ref(false)
  const previewImageUrl = ref('')

  const isEmpty = computed(() => files.value.length === 0 && !loading.value)

  const filteredFiles = computed(() => {
    const selected = treeSelectedKeys.value[0]
    if (selected === 'picture' || selected === 'file') {
      return files.value.filter((item) => item.type === selected)
    }
    return files.value
  })

  const treeData = computed(() => {
    const pictureCount = files.value.filter((item) => item.type === 'picture').length
    const fileCount = files.value.filter((item) => item.type === 'file').length
    return [
      {
        key: 'all',
        label: `全部 (${files.value.length})`,
        children: [
          { key: 'picture', label: `图片 (${pictureCount})` },
          { key: 'file', label: `文件 (${fileCount})` },
        ],
      },
    ]
  })

  async function fetchFiles() {
    loading.value = true
    try {
      const response = await listUploads({ page: page.value, pageSize: pageSize.value })
      files.value = response.items
      total.value = response.total
    } catch (error) {
      message.error('加载文件列表失败')
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  async function handleUpload({ file }: { file: UploadFileInfo }) {
    if (!file.file) return
    uploading.value = true
    try {
      const response = await uploadFile(file.file, uploadType.value)
      message.success(response.duplicated ? '文件已存在，已复用' : '上传成功')
      await fetchFiles()
    } catch (error) {
      message.error('上传失败')
      console.error(error)
    } finally {
      uploading.value = false
    }
  }

  async function handleCopyUrl(file: UploadFileResponse) {
    try {
      await navigator.clipboard.writeText(file.publicUrl)
      message.success('链接已复制到剪贴板')
    } catch (error) {
      message.error('复制失败')
      console.error(error)
    }
  }

  function openRenameModal(file: UploadFileResponse) {
    renamingFile.value = file
    newFileName.value = file.name
    renameModalVisible.value = true
  }

  async function handleRename() {
    if (!renamingFile.value || !newFileName.value.trim()) {
      message.warning('请输入文件名')
      return
    }
    try {
      await renameFile(renamingFile.value.id, { name: newFileName.value.trim() })
      message.success('重命名成功')
      renameModalVisible.value = false
      await fetchFiles()
    } catch (error) {
      message.error('重命名失败')
      console.error(error)
    }
  }

  function openDeleteModal(file: UploadFileResponse) {
    deletingFile.value = file
    deleteModalVisible.value = true
  }

  async function handleDelete() {
    if (!deletingFile.value) return
    try {
      await deleteFile(deletingFile.value.id)
      message.success('删除成功')
      deleteModalVisible.value = false
      if (files.value.length === 1 && page.value > 1) page.value--
      await fetchFiles()
    } catch (error) {
      message.error('删除失败')
      console.error(error)
    }
  }

  async function handleDownload(file: UploadFileResponse) {
    try {
      await downloadFile(file.id, file.name)
      message.success('下载开始')
    } catch (error) {
      message.error('下载失败')
      console.error(error)
    }
  }

  function handlePageChange(newPage: number) {
    page.value = newPage
    fetchFiles()
  }

  function handlePageSizeChange(newPageSize: number) {
    pageSize.value = newPageSize
    page.value = 1
    fetchFiles()
  }

  function openPreview(url: string) {
    previewImageUrl.value = url
    previewVisible.value = true
  }

  function handleTreeSelect(keys: Array<string | number>) {
    const selectedKey = String(keys[0] ?? 'all')
    treeSelectedKeys.value = [selectedKey]
  }

  onMounted(() => fetchFiles())

  return {
    files,
    loading,
    uploading,
    page,
    pageSize,
    total,
    uploadType,
    treeSelectedKeys,
    renameModalVisible,
    renamingFile,
    newFileName,
    deleteModalVisible,
    deletingFile,
    previewVisible,
    previewImageUrl,
    isEmpty,
    filteredFiles,
    treeData,
    fetchFiles,
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
  }
}
