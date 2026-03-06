<script setup lang="ts">
import { ComingSoon, ScrollContainer } from '@/components'

defineOptions({ name: 'FederationComingSoon' })
</script>

<template>
  <ScrollContainer wrapper-class="p-4">
    <ComingSoon
      title="Federation 联合"
      description="Blog Federation 与 ActivityPub 兼容功能尚未开放，敬请期待后续版本。"
      icon="ph--circles-three"
    />
  </ScrollContainer>
</template>
