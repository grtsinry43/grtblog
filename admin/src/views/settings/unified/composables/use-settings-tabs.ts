import { ref, reactive, watch, type Component, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import SiteInfoTab from '../components/tabs/SiteInfoTab.vue'
import ThemeExtendTab from '../components/tabs/ThemeExtendTab.vue'
import SecurityTab from '../components/tabs/SecurityTab.vue'
import ContentTab from '../components/tabs/ContentTab.vue'
import EmailTab from '../components/tabs/EmailTab.vue'
import WebhookTab from '../components/tabs/WebhookTab.vue'
import FederationTab from '../components/tabs/FederationTab.vue'
import AiTab from '../components/tabs/AiTab.vue'
import ApiTokensTab from '../components/tabs/ApiTokensTab.vue'
import AdvancedTab from '../components/tabs/AdvancedTab.vue'

export interface SettingsTab {
  key: string
  label: string
  icon: string
  component: Component
  /** When true, tab content fills the available height instead of scrolling externally. */
  fillHeight?: boolean
}

export const settingsTabs: SettingsTab[] = [
  { key: 'site-info', label: '基本信息', icon: 'iconify ph--globe-hemisphere-west', component: SiteInfoTab },
  { key: 'theme-extend', label: '主题扩展', icon: 'iconify ph--paint-brush', component: ThemeExtendTab, fillHeight: true },
  { key: 'security', label: '安全与登录', icon: 'iconify ph--shield-check', component: SecurityTab },
  { key: 'content', label: '内容与评论', icon: 'iconify ph--article', component: ContentTab },
  { key: 'email', label: '邮件', icon: 'iconify ph--envelope', component: EmailTab },
  { key: 'webhook', label: 'Webhook', icon: 'iconify ph--webhooks-logo', component: WebhookTab },
  { key: 'federation', label: '联合', icon: 'iconify ph--circles-three', component: FederationTab },
  { key: 'ai', label: 'AI', icon: 'iconify ph--robot', component: AiTab },
  { key: 'api-tokens', label: 'API Tokens', icon: 'iconify ph--key', component: ApiTokensTab },
  { key: 'advanced', label: '高级', icon: 'iconify ph--gear', component: AdvancedTab },
]

const validKeys = new Set(settingsTabs.map((t) => t.key))

export function useSettingsTabs() {
  const route = useRoute()
  const router = useRouter()

  const initialTab = validKeys.has(route.query.tab as string)
    ? (route.query.tab as string)
    : 'site-info'

  const activeTab = ref(initialTab)
  const dirtyTabs = reactive(new Set<string>())

  const currentTabDef = computed(() => settingsTabs.find((t) => t.key === activeTab.value)!)

  watch(activeTab, (tab) => {
    if (route.query.tab !== tab) {
      router.replace({ query: { ...route.query, tab } })
    }
  })

  function setDirty(tabKey: string, dirty: boolean) {
    if (dirty) {
      dirtyTabs.add(tabKey)
    } else {
      dirtyTabs.delete(tabKey)
    }
  }

  return {
    tabs: settingsTabs,
    activeTab,
    currentTabDef,
    dirtyTabs,
    setDirty,
  }
}
