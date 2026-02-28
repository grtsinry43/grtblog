<script setup lang="ts">
import { NScrollbar } from 'naive-ui'
import { computed } from 'vue'

import { ScrollContainer } from '@/components'
import { useLeaveConfirm } from '@/composables'
import { useComponentThemeOverrides } from '@/composables/useComponentThemeOverrides'

import { useSettingsTabs } from './composables/use-settings-tabs'

defineOptions({ name: 'UnifiedSettings' })

const { tabs, activeTab, currentTabDef, dirtyTabs, setDirty } = useSettingsTabs()
const { scrollbarInMainLayout } = useComponentThemeOverrides()

useLeaveConfirm({ when: () => dirtyTabs.size > 0 })

const currentComponent = computed(() => currentTabDef.value?.component)
const currentScrollable = computed(() => !currentTabDef.value?.fillHeight)
</script>

<template>
  <div class="flex h-full max-sm:flex-col">
    <!-- Mobile: horizontal scrollable tab bar -->
    <div
      class="hidden shrink-0 border-b border-neutral-200 max-sm:block dark:border-neutral-700"
    >
      <NScrollbar x-scrollable :theme-overrides="scrollbarInMainLayout">
        <div class="flex gap-1 px-2 py-2">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="relative flex shrink-0 items-center gap-1.5 rounded-md px-3 py-1.5 text-xs transition-colors"
            :class="
              activeTab === tab.key
                ? 'bg-[var(--primary-color-hover)]/10 font-medium text-[var(--primary-color)]'
                : 'text-neutral-500 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-neutral-800'
            "
            @click="activeTab = tab.key"
          >
            <span :class="tab.icon" class="text-sm" />
            <span>{{ tab.label }}</span>
            <span
              v-if="dirtyTabs.has(tab.key)"
              class="absolute -right-0.5 -top-0.5 h-1.5 w-1.5 rounded-full bg-amber-500"
            />
          </button>
        </div>
      </NScrollbar>
    </div>

    <!-- Desktop: left sidebar -->
    <div
      class="w-48 shrink-0 border-r border-neutral-200 max-sm:hidden dark:border-neutral-700"
    >
      <ScrollContainer wrapper-class="!p-0">
        <div class="space-y-0.5 px-2 py-3">
          <div
            v-for="tab in tabs"
            :key="tab.key"
            class="cursor-pointer rounded-md px-3 py-2 text-sm transition-colors"
            :class="
              activeTab === tab.key
                ? 'bg-[var(--primary-color-hover)]/10 font-medium text-[var(--primary-color)]'
                : 'text-neutral-600 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:bg-neutral-800'
            "
            @click="activeTab = tab.key"
          >
            <div class="flex items-center gap-2">
              <span :class="tab.icon" class="text-base" />
              <span>{{ tab.label }}</span>
              <span
                v-if="dirtyTabs.has(tab.key)"
                class="ml-auto h-1.5 w-1.5 shrink-0 rounded-full bg-amber-500"
              />
            </div>
          </div>
        </div>
      </ScrollContainer>
    </div>

    <!-- Right content -->
    <div class="min-w-0 flex-1">
      <ScrollContainer :scrollable="currentScrollable" wrapper-class="!p-3 max-sm:!p-2">
        <KeepAlive>
          <component
            :is="currentComponent"
            :key="activeTab"
            @dirty-change="(dirty: boolean) => setDirty(activeTab, dirty)"
          />
        </KeepAlive>
      </ScrollContainer>
    </div>
  </div>
</template>
