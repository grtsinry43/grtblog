<script setup lang="ts">
import {
  NButton,
  NButtonGroup,
  NDivider,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NForm,
  NFormItem,
  NImage,
  NInput,
  NModal,
  NPopconfirm,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  NTooltip,
} from 'naive-ui'
import { ref, watch } from 'vue'

import { ScrollContainer } from '@/components'
import ImageInput from '@/components/image-picker/ImageInput.vue'
import ImagePickerModal from '@/components/image-picker/ImagePickerModal.vue'
import { uploadFile } from '@/services/uploads'

import { useAlbumForm } from './composables/useAlbumForm'
import {
  extractExif,
  exifDevice,
  exifShootingInfo,
  exifLocation,
} from './composables/usePhotoUtils'

import type { PhotoItem } from '@/services/albums'

const {
  isEdit,
  loading,
  saving,
  form,
  photos,
  save,
  uploadPhotos,
  updatePhoto,
  deletePhoto,
  handleReorder,
  markDirty,
} = useAlbumForm()

const uploading = ref(false)
const showMeta = ref(false)
const showPhotoPicker = ref(false)
const editingPhoto = ref<PhotoItem | null>(null)
const showPhotoModal = ref(false)
const photoForm = ref({
  description: '',
  caption: '',
})
watch(
  () => [
    form.title,
    form.description,
    form.cover,
    form.shortUrl,
    form.isPublished,
    form.allowComment,
  ],
  () => markDirty(),
  { deep: true },
)

async function handlePhotoUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  uploading.value = true
  try {
    const items: { url: string; exif?: Record<string, unknown> | null }[] = []
    for (const file of Array.from(input.files)) {
      const [res, exif] = await Promise.all([uploadFile(file, 'picture'), extractExif(file)])
      // Merge backend imageMeta (width/height/dominantColor) into exif
      const merged = { ...(exif || {}) }
      if (res.imageMeta) {
        if (res.imageMeta.width) merged.imageWidth = res.imageMeta.width
        if (res.imageMeta.height) merged.imageHeight = res.imageMeta.height
        if (res.imageMeta.dominantColor) merged.dominantColor = res.imageMeta.dominantColor
      }
      items.push({ url: res.publicUrl, exif: Object.keys(merged).length ? merged : null })
    }
    await uploadPhotos(items)
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function handlePickFromGallery(url: string) {
  await uploadPhotos([{ url }])
}

function openPhotoEdit(photo: PhotoItem) {
  editingPhoto.value = photo
  photoForm.value = {
    description: photo.description ?? '',
    caption: photo.caption ?? '',
  }
  showPhotoModal.value = true
}

async function savePhotoEdit() {
  if (!editingPhoto.value) return
  await updatePhoto(editingPhoto.value.id, {
    description: photoForm.value.description || null,
    caption: photoForm.value.caption || null,
  } as any)
  showPhotoModal.value = false
}

function movePhoto(index: number, direction: -1 | 1) {
  const ids = photos.value.map((p) => p.id)
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= ids.length) return
  const a = ids[index]!
  const b = ids[targetIndex]!
  ids[index] = b
  ids[targetIndex] = a
  handleReorder(ids)
}

function thumbUrl(photo: PhotoItem): string {
  return photo.thumbnailUrl || photo.url
}
</script>

<template>
  <NSpin
    :show="loading"
    class="flex h-full min-h-0 flex-col"
  >
    <!-- ====== Sticky Header ====== -->
    <header
      class="z-10 flex shrink-0 flex-col gap-3 px-6 py-6 backdrop-blur sm:h-20 sm:flex-row sm:items-center sm:justify-between sm:px-10 sm:py-0"
    >
      <!-- Left: Title -->
      <div class="flex min-w-0 flex-1 items-center gap-3">
        <NInput
          v-model:value="form.title"
          :bordered="false"
          placeholder="输入相册标题..."
          class="flex-1 text-xl font-bold sm:text-2xl"
          :style="{ '--n-caret-color': 'var(--primary-color)', backgroundColor: 'transparent' }"
        />
      </div>

      <!-- Right: Actions -->
      <div class="flex shrink-0 flex-wrap items-center gap-2 sm:flex-nowrap sm:gap-3">
        <!-- Short URL -->
        <div class="flex items-center gap-1 text-[11px] opacity-50">
          <div class="iconify ph--link-simple" />
          <span>/albums/</span>
          <input
            v-model="form.shortUrl"
            class="w-20 border-b border-current/20 bg-transparent transition-colors outline-none focus:border-current/50 sm:w-28"
            placeholder="auto"
          />
        </div>

        <!-- Publish toggle -->
        <NButtonGroup size="small">
          <NButton
            :type="form.isPublished ? 'default' : 'primary'"
            :ghost="form.isPublished"
            @click="form.isPublished = false"
          >
            草稿
          </NButton>
          <NButton
            :type="form.isPublished ? 'primary' : 'default'"
            :ghost="!form.isPublished"
            @click="form.isPublished = true"
          >
            发布
          </NButton>
        </NButtonGroup>

        <!-- Metadata drawer toggle -->
        <NTooltip>
          <template #trigger>
            <NButton
              quaternary
              circle
              size="small"
              @click="showMeta = !showMeta"
            >
              <template #icon><div class="iconify ph--sliders-horizontal" /></template>
            </NButton>
          </template>
          元数据
        </NTooltip>

        <!-- Save -->
        <NButton
          type="primary"
          :loading="saving"
          @click="save"
        >
          <template #icon><div class="iconify ph--floppy-disk" /></template>
          {{ isEdit ? '保存' : '创建' }}
        </NButton>
      </div>
    </header>

    <!-- ====== Main Content ====== -->
    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div class="flex min-h-0 flex-1 flex-col px-6 pb-20 sm:px-10">
        <!-- Description -->
        <div class="mb-6">
          <NInput
            v-model:value="form.description"
            type="textarea"
            :bordered="false"
            placeholder="添加相册描述..."
            :autosize="{ minRows: 2, maxRows: 5 }"
            :style="{ backgroundColor: 'transparent' }"
          />
        </div>

        <!-- ====== Photo Grid ====== -->
        <div class="mb-4 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="iconify text-lg opacity-50 ph--images" />
            <span class="text-sm font-medium">照片</span>
            <NTag
              v-if="photos.length > 0"
              size="small"
              round
              :bordered="false"
            >
              {{ photos.length }}
            </NTag>
          </div>
          <NSpace
            v-if="isEdit"
            size="small"
          >
            <NButton
              size="small"
              quaternary
              @click="showPhotoPicker = true"
            >
              <template #icon><div class="iconify ph--folder-open" /></template>
              图库
            </NButton>
            <NButton
              size="small"
              :loading="uploading"
              tag="label"
              style="cursor: pointer"
            >
              <template #icon><div class="iconify ph--upload-simple" /></template>
              上传
              <input
                type="file"
                accept="image/*"
                multiple
                style="display: none"
                @change="handlePhotoUpload"
              />
            </NButton>
          </NSpace>
        </div>

        <NEmpty
          v-if="!isEdit"
          description="请先创建相册，再添加照片"
          class="py-16"
        />
        <NEmpty
          v-else-if="photos.length === 0"
          description="点击上方按钮添加照片"
          class="py-16"
        >
          <template #icon>
            <div class="iconify text-4xl opacity-20 ph--camera" />
          </template>
        </NEmpty>

        <!-- Photo cards -->
        <ScrollContainer
          v-else
          class="min-h-0 flex-1"
          wrapper-class="!p-0"
        >
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            <div
              v-for="(photo, index) in photos"
              :key="photo.id"
              class="group overflow-hidden rounded-lg border border-current/5 transition-all hover:border-current/15 hover:shadow-sm"
            >
              <!-- Image -->
              <div class="relative aspect-square overflow-hidden bg-current/3">
                <NImage
                  :src="thumbUrl(photo)"
                  :preview-src="photo.url"
                  object-fit="cover"
                  :img-props="{ class: 'h-full w-full object-cover' }"
                  class="h-full w-full"
                />
                <!-- Index badge (always visible) -->
                <div
                  class="absolute top-1.5 left-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-black/50 text-[10px] font-medium text-white"
                >
                  {{ index + 1 }}
                </div>
              </div>

              <!-- Info + Actions bar below image -->
              <div class="px-2.5 py-2">
                <!-- Caption / EXIF summary -->
                <div
                  class="min-h-[1.25rem] cursor-pointer truncate text-xs opacity-60"
                  @click="openPhotoEdit(photo)"
                >
                  <template v-if="photo.caption">{{ photo.caption }}</template>
                  <template v-else-if="exifDevice(photo.exif)">{{
                    exifDevice(photo.exif)
                  }}</template>
                  <template v-else><span class="italic opacity-40">点击编辑信息</span></template>
                </div>

                <!-- Action buttons -->
                <div class="mt-1.5 flex items-center justify-between">
                  <NSpace :size="2">
                    <NTooltip>
                      <template #trigger>
                        <NButton
                          quaternary
                          circle
                          size="tiny"
                          :disabled="index === 0"
                          @click="movePhoto(index, -1)"
                        >
                          <template #icon><div class="iconify ph--caret-left" /></template>
                        </NButton>
                      </template>
                      前移
                    </NTooltip>
                    <NTooltip>
                      <template #trigger>
                        <NButton
                          quaternary
                          circle
                          size="tiny"
                          :disabled="index === photos.length - 1"
                          @click="movePhoto(index, 1)"
                        >
                          <template #icon><div class="iconify ph--caret-right" /></template>
                        </NButton>
                      </template>
                      后移
                    </NTooltip>
                    <NTooltip>
                      <template #trigger>
                        <NButton
                          quaternary
                          circle
                          size="tiny"
                          @click="openPhotoEdit(photo)"
                        >
                          <template #icon><div class="iconify ph--pencil-simple" /></template>
                        </NButton>
                      </template>
                      编辑
                    </NTooltip>
                  </NSpace>
                  <NPopconfirm @positive-click="deletePhoto(photo.id)">
                    <template #trigger>
                      <NButton
                        quaternary
                        circle
                        size="tiny"
                        type="error"
                      >
                        <template #icon><div class="iconify ph--trash" /></template>
                      </NButton>
                    </template>
                    确定删除这张照片？
                  </NPopconfirm>
                </div>
              </div>
            </div>
          </div>
        </ScrollContainer>
      </div>
    </main>

    <!-- ====== Metadata Drawer ====== -->
    <NDrawer
      v-model:show="showMeta"
      placement="right"
      :width="380"
    >
      <NDrawerContent
        title="相册设置"
        :native-scrollbar="false"
        closable
      >
        <div class="flex flex-col gap-6">
          <!-- Cover -->
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--image" />
              <span>封面</span>
            </div>
            <ImageInput v-model:value="form.cover" />
          </div>

          <NDivider style="margin: 0" />

          <!-- Properties -->
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--gear-six" />
              <span>属性</span>
            </div>
            <NForm
              label-placement="left"
              label-width="auto"
              :show-feedback="false"
              class="space-y-3"
            >
              <NFormItem label="允许评论">
                <NSwitch v-model:value="form.allowComment" />
              </NFormItem>
            </NForm>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>

    <!-- ====== Photo Edit Modal ====== -->
    <NModal
      v-model:show="showPhotoModal"
      preset="card"
      title="编辑照片信息"
      style="width: 520px; max-width: 90vw"
    >
      <div class="flex flex-col gap-4">
        <!-- EXIF info (read-only) -->
        <div
          v-if="editingPhoto?.exif"
          class="rounded-lg bg-current/3 p-3 text-xs"
        >
          <div class="mb-2 flex items-center gap-1.5 text-sm font-medium opacity-70">
            <div class="iconify ph--info" />
            EXIF 信息
          </div>
          <div class="grid grid-cols-2 gap-x-4 gap-y-1.5 opacity-60">
            <div v-if="exifDevice(editingPhoto.exif)">
              <span class="mr-1 iconify inline-block align-[-2px] ph--camera" />{{
                exifDevice(editingPhoto.exif)
              }}
            </div>
            <div v-if="editingPhoto.exif.lensModel">
              <span class="mr-1 iconify inline-block align-[-2px] ph--aperture" />{{
                editingPhoto.exif.lensModel
              }}
            </div>
            <div
              v-if="exifShootingInfo(editingPhoto.exif)"
              class="col-span-2"
            >
              {{ exifShootingInfo(editingPhoto.exif) }}
            </div>
            <div v-if="exifLocation(editingPhoto.exif)">
              <span class="mr-1 iconify inline-block align-[-2px] ph--map-pin" />{{
                exifLocation(editingPhoto.exif)
              }}
            </div>
            <div v-if="editingPhoto.exif.dateTimeOriginal">
              <span class="mr-1 iconify inline-block align-[-2px] ph--clock" />{{
                editingPhoto.exif.dateTimeOriginal
              }}
            </div>
            <div v-if="editingPhoto.exif.imageWidth && editingPhoto.exif.imageHeight">
              {{ editingPhoto.exif.imageWidth }} × {{ editingPhoto.exif.imageHeight }}
            </div>
          </div>
        </div>

        <!-- Editable fields -->
        <NForm
          label-placement="top"
          :show-feedback="false"
          class="space-y-3"
        >
          <NFormItem label="说明文字">
            <NInput
              v-model:value="photoForm.caption"
              placeholder="给照片加一句话..."
            />
          </NFormItem>
          <NFormItem label="详细描述">
            <NInput
              v-model:value="photoForm.description"
              type="textarea"
              :rows="3"
              placeholder="更多描述..."
            />
          </NFormItem>
        </NForm>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPhotoModal = false">取消</NButton>
          <NButton
            type="primary"
            @click="savePhotoEdit"
          >
            <template #icon><div class="iconify ph--check" /></template>
            保存
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- ====== Image Picker Modal ====== -->
    <ImagePickerModal
      v-model:show="showPhotoPicker"
      @select="handlePickFromGallery"
    />
  </NSpin>
</template>
