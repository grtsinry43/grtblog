<script setup lang="ts">
import {
  NButton,
  NCard,
  NEmpty,
  NImage,
  NModal,
  NPagination,
  NSpin,
  NUpload,
  type UploadFileInfo,
  useMessage,
} from 'naive-ui'
import { computed, ref, watch } from 'vue'
import { listUploads, uploadFile, type UploadFileResponse } from '@/services/uploads'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  select: [url: string]
}>()

const message = useMessage()

const loading = ref(false)
const uploading = ref(false)
const items = ref<UploadFileResponse[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 12
const selectedUrl = ref<string | null>(null)

const pictureItems = computed(() => items.value.filter((i) => i.type === 'picture'))

async function fetchList() {
  loading.value = true
  try {
    const res = await listUploads({ page: page.value, pageSize })
    items.value = res.items
    total.value = res.total
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '加载图片列表失败')
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number) {
  page.value = p
  selectedUrl.value = null
  fetchList()
}

async function handleUpload({ file }: { file: UploadFileInfo }) {
  if (!file.file) return
  uploading.value = true
  try {
    const res = await uploadFile(file.file, 'picture')
    message.success('上传成功')
    await fetchList()
    selectedUrl.value = res.publicUrl
  } catch (e: unknown) {
    message.error(e instanceof Error ? e.message : '上传失败')
  } finally {
    uploading.value = false
  }
}

function handleConfirm() {
  if (selectedUrl.value) {
    emit('select', selectedUrl.value)
    emit('update:show', false)
  }
}

function handleClose() {
  emit('update:show', false)
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      page.value = 1
      selectedUrl.value = null
      fetchList()
    }
  },
)
</script>

<template>
  <NModal
    :show="show"
    @update:show="emit('update:show', $event)"
    style="width: 720px; max-width: 90vw"
  >
    <NCard
      title="选择图片"
      closable
      @close="handleClose"
    >
      <div class="flex flex-col gap-4">
        <div>
          <NUpload
            :show-file-list="false"
            :custom-request="({ file }) => handleUpload({ file })"
            accept="image/*"
            :disabled="uploading"
          >
            <NButton
              :loading="uploading"
              type="primary"
              ghost
            >
              <template #icon><div class="iconify ph--upload-simple" /></template>
              上传新图片
            </NButton>
          </NUpload>
        </div>

        <NSpin :show="loading">
          <div
            v-if="pictureItems.length > 0"
            class="grid grid-cols-4 gap-3"
          >
            <div
              v-for="item in pictureItems"
              :key="item.id"
              class="group cursor-pointer overflow-hidden rounded-lg border-2 transition-all"
              :class="
                selectedUrl === item.publicUrl
                  ? 'border-primary shadow-md'
                  : 'border-transparent hover:border-current/20'
              "
              @click="selectedUrl = item.publicUrl"
            >
              <div class="relative aspect-square overflow-hidden bg-gray-100 dark:bg-gray-800">
                <NImage
                  :src="item.publicUrl"
                  :alt="item.name"
                  object-fit="cover"
                  preview-disabled
                  class="h-full w-full"
                  :img-props="{ class: 'h-full w-full object-cover' }"
                />
                <div
                  v-if="selectedUrl === item.publicUrl"
                  class="absolute inset-0 flex items-center justify-center bg-primary/20"
                >
                  <div class="iconify text-2xl text-white ph--check-circle-fill" />
                </div>
              </div>
              <div class="truncate px-2 py-1.5 text-xs opacity-70">
                {{ item.name }}
              </div>
            </div>
          </div>
          <NEmpty
            v-else-if="!loading"
            description="暂无图片"
            class="py-8"
          />
        </NSpin>

        <div class="flex items-center justify-between">
          <NPagination
            v-if="total > pageSize"
            :page="page"
            :page-size="pageSize"
            :item-count="total"
            @update:page="handlePageChange"
          />
          <div v-else />
          <div class="flex gap-2">
            <NButton @click="handleClose">取消</NButton>
            <NButton
              type="primary"
              :disabled="!selectedUrl"
              @click="handleConfirm"
            >
              确认
            </NButton>
          </div>
        </div>
      </div>
    </NCard>
  </NModal>
</template>
