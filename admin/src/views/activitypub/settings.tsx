import { NButton, NCard, NCollapse, NCollapseItem, NEmpty, NForm, NSpin, NTag } from 'naive-ui'
import { defineComponent } from 'vue'

import { ScrollContainer } from '@/components'
import { useLeaveConfirm } from '@/composables'
import { listActivityPubConfigs, updateActivityPubConfigs } from '@/services/activitypub-config'
import ConfigItem from '@/views/sysconfig/components/ConfigItem'
import { useConfigCenter } from '@/views/sysconfig/composables/use-config-center'

import type { SysConfigGroup } from '@/services/sysconfig'

export default defineComponent({
  name: 'ActivityPubSettings',
  setup() {
    const {
      loading,
      saving,
      tree,
      valueMap,
      jsonBufferMap,
      expandedGroups,
      pendingCount,
      isItemVisible,
      fetch,
      save,
    } = useConfigCenter(listActivityPubConfigs, updateActivityPubConfigs)

    useLeaveConfirm({ when: () => pendingCount.value > 0 })

    const renderGroups = (groups?: SysConfigGroup[]) => {
      if (!groups || groups.length === 0) return null

      return groups.map((group) => (
        <NCollapseItem
          key={group.path}
          name={group.path}
          title={group.label || group.key}
        >
          <div class='space-y-4 pl-2'>
            {group.items && group.items.length > 0 && (
              <NForm
                labelPlacement='left'
                labelWidth={160}
              >
                {group.items.map((item) => (
                  <ConfigItem
                    key={item.key}
                    item={item}
                    valueMap={valueMap}
                    jsonBufferMap={jsonBufferMap}
                    visible={isItemVisible}
                  />
                ))}
              </NForm>
            )}

            {group.children && group.children.length > 0 && (
              <NCollapse>{renderGroups(group.children)}</NCollapse>
            )}
          </div>
        </NCollapseItem>
      ))
    }

    return () => (
      <ScrollContainer wrapperClass='p-4'>
        <NCard>
          {{
            header: () => (
              <div class='flex flex-wrap items-center justify-between gap-3'>
                <div>
                  <div class='text-base font-semibold'>ActivityPub 设置</div>
                  <div class='text-xs text-neutral-500'>
                    兼容功能独立配置，启用后将使用 ActivityPub 专用密钥
                  </div>
                </div>
                <div class='flex items-center gap-2'>
                  {pendingCount.value > 0 && (
                    <NTag type='warning'>待保存 {pendingCount.value}</NTag>
                  )}
                  <NButton
                    size='small'
                    secondary
                    loading={loading.value}
                    onClick={fetch}
                  >
                    刷新
                  </NButton>
                  <NButton
                    size='small'
                    type='primary'
                    loading={saving.value}
                    onClick={save}
                  >
                    保存
                  </NButton>
                </div>
              </div>
            ),
            default: () => (
              <NSpin show={loading.value}>
                {!tree.value || (!tree.value.items?.length && !tree.value.groups?.length) ? (
                  <div class='py-8'>
                    <NEmpty description='暂无配置项' />
                  </div>
                ) : (
                  <div class='space-y-6'>
                    {tree.value.items && tree.value.items.length > 0 && (
                      <div class='rounded-lg border border-dashed border-neutral-200 p-4 dark:border-neutral-700'>
                        <div class='mb-4 text-sm font-medium text-neutral-600 dark:text-neutral-400'>
                          未分组
                        </div>
                        <NForm
                          labelPlacement='left'
                          labelWidth={160}
                        >
                          {tree.value.items.map((item) => (
                            <ConfigItem
                              key={item.key}
                              item={item}
                              valueMap={valueMap}
                              jsonBufferMap={jsonBufferMap}
                              visible={isItemVisible}
                            />
                          ))}
                        </NForm>
                      </div>
                    )}

                    {tree.value.groups && tree.value.groups.length > 0 && (
                      <NCollapse v-model:expandedNames={expandedGroups.value}>
                        {renderGroups(tree.value.groups)}
                      </NCollapse>
                    )}
                  </div>
                )}
              </NSpin>
            ),
          }}
        </NCard>
      </ScrollContainer>
    )
  },
})
