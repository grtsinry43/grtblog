<script setup lang="ts">
import { NButton, NSelect } from 'naive-ui'
import { computed, shallowRef, watch } from 'vue'

import {
  contentTagNameKey,
  findContentTagByName,
  normalizeContentTagName,
} from '@/views/shared/content-editor/model/content-tags'

import type { ContentTagOption } from '@/views/shared/content-editor/model/content-tags'
import type { SelectOption } from 'naive-ui'

const selectedIds = defineModel<number[]>({ required: true })
const props = defineProps<{
  options: ContentTagOption[]
  noun: '标签' | '话题'
  loading?: boolean
  creating?: boolean
}>()

const emit = defineEmits<{
  create: [name: string]
}>()

const searchValue = shallowRef('')
const show = shallowRef(false)
const pendingCreateName = shallowRef('')
const normalizedSearchValue = computed(() => normalizeContentTagName(searchValue.value))
const exactMatch = computed(() => findContentTagByName(props.options, normalizedSearchValue.value))
const canCreate = computed(() => !!normalizedSearchValue.value && !exactMatch.value)

function filterOption(pattern: string, option: SelectOption) {
  return contentTagNameKey(String(option.label ?? '')).includes(contentTagNameKey(pattern))
}

function createFromSearch() {
  if (!canCreate.value || props.creating) return
  pendingCreateName.value = normalizedSearchValue.value
  emit('create', normalizedSearchValue.value)
}

function handleSearch(value: string) {
  searchValue.value = value
  if (
    pendingCreateName.value &&
    contentTagNameKey(value) !== contentTagNameKey(pendingCreateName.value)
  ) {
    pendingCreateName.value = ''
  }
}

watch(
  [() => props.options, selectedIds],
  () => {
    if (!pendingCreateName.value) return
    const created = findContentTagByName(props.options, pendingCreateName.value)
    if (!created || !selectedIds.value.includes(created.value)) return
    searchValue.value = ''
    pendingCreateName.value = ''
    show.value = false
  },
  { deep: false },
)
</script>

<template>
  <NSelect
    v-model:value="selectedIds"
    v-model:show="show"
    :options="options"
    :loading="loading"
    :placeholder="`搜索并选择${noun}`"
    :filter="filterOption"
    :max-tag-count="'responsive'"
    multiple
    filterable
    clearable
    show-on-focus
    @search="handleSearch"
  >
    <template #action>
      <NButton
        v-if="canCreate"
        text
        block
        type="primary"
        :loading="creating"
        @click="createFromSearch"
      >
        <template #icon>
          <div class="iconify ph--plus" />
        </template>
        创建并选择“{{ normalizedSearchValue }}”
      </NButton>
      <span
        v-else-if="normalizedSearchValue"
        class="block px-1 text-xs opacity-55"
      >
        已有同名{{ noun }}，请在上方选择
      </span>
      <span
        v-else
        class="block px-1 text-xs opacity-55"
      >
        输入关键词搜索；没有结果时可在这里新建
      </span>
    </template>
  </NSelect>
</template>
