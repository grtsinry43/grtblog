<script setup lang="ts">
defineOptions({
  name: 'ComingSoon',
})

withDefaults(
  defineProps<{
    title: string
    description?: string
    icon?: string
  }>(),
  {
    description: '',
    icon: 'ph--rocket-launch',
  },
)
</script>

<template>
  <div class="flex flex-col items-center justify-center px-6 py-24 text-center">
    <div
      class="mb-6 grid size-20 place-items-center rounded-full border border-naive-border bg-naive-card transition-[background-color,border-color]"
    >
      <span
        class="iconify text-4xl text-primary"
        :class="icon"
      />
    </div>
    <h2 class="mb-2 text-xl font-medium text-neutral-700 dark:text-neutral-200">{{ title }}</h2>
    <p
      v-if="description"
      class="mb-4 max-w-sm text-sm text-neutral-500 dark:text-neutral-400"
    >
      {{ description }}
    </p>
    <span
      class="inline-block rounded-full border border-naive-border bg-naive-card px-4 py-1.5 text-xs text-neutral-400 transition-[background-color,border-color]"
    >
      即将在未来版本中推出
    </span>
  </div>
</template>
