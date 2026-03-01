<script setup lang="ts">
defineOptions({
  name: 'PageHeader',
})

withDefaults(defineProps<{
  title: string
  description?: string
  icon?: string
}>(), {
  description: '',
  icon: '',
})
</script>

<template>
  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div>
      <h1 class="flex items-center gap-3 text-2xl font-light text-gray-900 dark:text-gray-100">
        <span v-if="icon" class="iconify text-primary" :class="icon" />
        <span>{{ title }}</span>
        <slot name="badge" />
      </h1>
      <slot name="description">
        <p v-if="description" class="mt-1 pl-1 text-sm text-gray-500">{{ description }}</p>
      </slot>
    </div>
    <div v-if="$slots.actions" class="flex items-center gap-2 self-end sm:self-auto">
      <slot name="actions" />
    </div>
  </div>
</template>
