import { NButton, NCard } from 'naive-ui'
import { defineComponent } from 'vue'
import { useRouter } from 'vue-router'

import { ScrollContainer } from '@/components'

export default defineComponent({
  name: 'FederationInstances',
  setup() {
    const router = useRouter()

    return () => (
      <ScrollContainer wrapper-class='p-4'>
        <NCard>
          <div class='space-y-6'>
            <div>
              <div class='text-base font-semibold'>联邦实例</div>
              <div class='text-xs text-neutral-500'>发起友链申请已归属到友链管理模块</div>
            </div>

            <div class='flex justify-end'>
              <NButton
                type='primary'
                onClick={() => router.push({ name: 'friendLinkList' })}
              >
                前往友链管理
              </NButton>
            </div>
          </div>
        </NCard>
      </ScrollContainer>
    )
  },
})
