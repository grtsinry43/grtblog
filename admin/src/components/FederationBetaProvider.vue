<script setup lang="ts">
import { useDialog } from 'naive-ui'
import { onMounted, onUnmounted } from 'vue'

import {
  federationBetaMessage,
  federationBetaTitle,
  markFederationBetaAcknowledged,
  registerFederationBetaConfirmHandler,
} from '@/utils/federation-beta'

const dialog = useDialog()

function confirmFederationBeta() {
  return new Promise<boolean>((resolve) => {
    dialog.warning({
      title: federationBetaTitle,
      content: federationBetaMessage,
      positiveText: '我要尝鲜',
      negativeText: '我再看看',
      maskClosable: false,
      closable: false,
      onPositiveClick: () => {
        markFederationBetaAcknowledged()
        resolve(true)
      },
      onNegativeClick: () => resolve(false),
      onClose: () => resolve(false),
    })
  })
}

onMounted(() => {
  registerFederationBetaConfirmHandler(confirmFederationBeta)
})

onUnmounted(() => {
  registerFederationBetaConfirmHandler(null)
})
</script>

<template>
  <slot />
</template>
