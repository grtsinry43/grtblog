<script setup lang="ts">
import { onMounted, ref, toRef, watch } from 'vue'

import { useTurnstile } from '@/composables/use-turnstile'

const props = defineProps<{
  siteKey: string
}>()

const emit = defineEmits<{
  'update:token': [value: string]
  error: [message: string]
  expired: []
}>()

const containerRef = ref<HTMLElement | null>(null)
const siteKeyRef = toRef(props, 'siteKey')

const { token, error, expired, render } = useTurnstile(containerRef, siteKeyRef)

watch(token, (v) => emit('update:token', v))
watch(error, (v) => {
  if (v) emit('error', v)
})
watch(expired, (v) => {
  if (v) emit('expired')
})

onMounted(() => {
  if (props.siteKey) {
    render()
  }
})

watch(siteKeyRef, (key) => {
  if (key) render()
})
</script>

<template>
  <div
    v-if="siteKey"
    ref="containerRef"
    class="turnstile-container"
  />
  <div
    v-else
    class="text-sm text-neutral-400"
  >
    Turnstile 未配置
  </div>
</template>

<style scoped>
.turnstile-container {
  display: flex;
  justify-content: center;
}
</style>
