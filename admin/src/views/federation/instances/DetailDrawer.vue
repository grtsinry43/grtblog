<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useWindowSize } from '@vueuse/core'
import {
  NDrawer,
  NDrawerContent,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NCode,
  NSpin,
  NTabs,
  NTabPane,
  NEmpty,
} from 'naive-ui'
import { computed, watch } from 'vue'

import { ScrollContainer } from '@/components'
import { getFederationInstanceDetail } from '@/services/federation-admin'

import type { FederationInstanceDetailResp } from '@/types/federation'

const props = defineProps<{
  show: boolean
  instanceId?: number
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const { width } = useWindowSize()
const drawerWidth = computed(() => (width.value < 640 ? '100%' : 600))

const {
  data: instance,
  isPending,
  refetch,
} = useQuery({
  queryKey: ['federation-instance-detail', props.instanceId],
  queryFn: () => getFederationInstanceDetail(props.instanceId!),
  enabled: computed(() => !!props.instanceId && props.show),
})

watch(
  () => props.show,
  (newVal) => {
    if (newVal && props.instanceId) {
      refetch()
    }
  },
)
</script>

<template>
  <NDrawer
    :show="show"
    @update:show="(val) => emit('update:show', val)"
    :width="drawerWidth"
  >
    <NDrawerContent
      title="实例详情"
      closable
      :native-scrollbar="false"
    >
      <ScrollContainer>
        <div
          v-if="isPending"
          class="flex justify-center p-8"
        >
          <NSpin />
        </div>
        <div
          v-else-if="instance"
          class="space-y-6 p-4"
        >
          <NDescriptions
            bordered
            :column="1"
            label-placement="left"
            title="基础信息"
          >
            <NDescriptionsItem label="ID">{{ instance.id }}</NDescriptionsItem>
            <NDescriptionsItem label="域名">{{ instance.base_url }}</NDescriptionsItem>
            <NDescriptionsItem label="名称">{{ instance.name || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="描述">{{ instance.description || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="协议版本">{{
              instance.protocol_version || '-'
            }}</NDescriptionsItem>
            <NDescriptionsItem label="状态">
              <NTag
                :type="
                  instance.status === 'active'
                    ? 'success'
                    : instance.status === 'blocked'
                      ? 'error'
                      : 'warning'
                "
              >
                {{ instance.status }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="最后可见">{{
              instance.last_seen_at ? new Date(instance.last_seen_at).toLocaleString() : '-'
            }}</NDescriptionsItem>
            <NDescriptionsItem label="加入时间">{{
              new Date(instance.created_at).toLocaleString()
            }}</NDescriptionsItem>
          </NDescriptions>

          <NDescriptions
            bordered
            :column="1"
            label-placement="left"
            title="技术细节"
          >
            <NDescriptionsItem label="Key ID">{{ instance.key_id || '-' }}</NDescriptionsItem>
            <NDescriptionsItem
              label="Remote Error"
              v-if="instance.remote_error"
            >
              <span class="text-red-500">{{ instance.remote_error }}</span>
            </NDescriptionsItem>
          </NDescriptions>

          <NDescriptions
            bordered
            :column="1"
            label-placement="left"
            title="实例元数据"
            v-if="instance.manifest?.software"
          >
            <NDescriptionsItem label="软件">
              {{ instance.manifest.software?.name || 'Unknown' }}
              {{ instance.manifest.software?.version || '' }}
            </NDescriptionsItem>
          </NDescriptions>

          <NDescriptions
            bordered
            :column="1"
            label-placement="left"
            title="策略配置 (Policies)"
            v-if="instance.policies"
          >
            <NDescriptionsItem label="允许引用 (Citation)">
              <NTag
                :type="instance.policies.allow_citation ? 'success' : 'error'"
                size="small"
              >
                {{ instance.policies.allow_citation ? '允许' : '禁止' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="允许提及 (Mention)">
              <NTag
                :type="instance.policies.allow_mention ? 'success' : 'error'"
                size="small"
              >
                {{ instance.policies.allow_mention ? '允许' : '禁止' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="自动通过友链/引用">
              <NTag
                :type="instance.policies.auto_approve_friendlink_citation ? 'success' : 'warning'"
                size="small"
              >
                {{ instance.policies.auto_approve_friendlink_citation ? '开启' : '关闭' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="强制 HTTPS">
              <NTag
                :type="instance.policies.require_https ? 'success' : 'warning'"
                size="small"
              >
                {{ instance.policies.require_https ? '开启' : '关闭' }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="最大缓存时间">
              {{ instance.policies.max_cache_age }} 秒
            </NDescriptionsItem>
          </NDescriptions>

          <div v-if="instance.public_key">
            <h3 class="mb-2 font-bold">Public Key</h3>
            <NCode
              :code="instance.public_key"
              language="text"
              word-wrap
              class="rounded bg-gray-100 p-2 text-xs dark:bg-gray-800"
            />
          </div>

          <NTabs
            type="line"
            animated
          >
            <NTabPane
              name="manifest"
              tab="Manifest"
            >
              <NCode
                v-if="instance.manifest"
                :code="JSON.stringify(instance.manifest, null, 2)"
                language="json"
                word-wrap
                class="rounded bg-gray-100 p-2 text-xs dark:bg-gray-800"
              />
              <NEmpty
                v-else
                description="无数据"
              />
            </NTabPane>
            <NTabPane
              name="endpoints"
              tab="Endpoints"
            >
              <NCode
                v-if="instance.endpoints"
                :code="JSON.stringify(instance.endpoints, null, 2)"
                language="json"
                word-wrap
                class="rounded bg-gray-100 p-2 text-xs dark:bg-gray-800"
              />
              <NEmpty
                v-else
                description="无数据"
              />
            </NTabPane>
            <NTabPane
              name="features"
              tab="Features"
            >
              <NCode
                v-if="instance.features"
                :code="JSON.stringify(instance.features, null, 2)"
                language="json"
                word-wrap
                class="rounded bg-gray-100 p-2 text-xs dark:bg-gray-800"
              />
              <NEmpty
                v-else
                description="无数据"
              />
            </NTabPane>
          </NTabs>
        </div>
        <div
          v-else
          class="flex justify-center p-8"
        >
          <NEmpty description="未找到实例信息" />
        </div>
      </ScrollContainer>
    </NDrawerContent>
  </NDrawer>
</template>
