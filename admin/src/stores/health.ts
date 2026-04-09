import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { useDiscreteApi } from '@/composables/useDiscreteApi'
import {
  deriveMode,
  fetchReadiness,
  probeNginx,
  probeRenderer,
  type HealthWSPayload,
  type SystemMode,
} from '@/services/health'

export const useHealthStore = defineStore('healthStore', () => {
  // ── Backend bits (from WS) ────────────────────────────────
  // bits 4,3,2 = backend, db, redis
  const backendBits = ref(0b111)
  const maintenance = ref(false) // bit 0 (from backend WS)
  const isDev = ref(import.meta.env.MODE === 'development')

  // ── Frontend probe results ────────────────────────────────
  const nginxOk = ref(true)
  const rendererOk = ref(true)

  // ── Polling control ───────────────────────────────────────
  let pollTimer: ReturnType<typeof setInterval> | null = null

  // ── 6-bit composite state ─────────────────────────────────
  const state = computed(() => {
    let v = 0
    if (nginxOk.value) v |= 1 << 5
    v |= (backendBits.value & 0b111) << 2 // bits 4,3,2
    if (rendererOk.value) v |= 1 << 1
    if (!maintenance.value) v |= 1 // bit 0: 1 = normal, 0 = maintenance
    return v
  })

  const mode = computed<SystemMode>(() => {
    if (isDev.value) return 'healthy'
    return deriveMode(state.value)
  })

  const components = computed(() => ({
    nginx: nginxOk.value,
    backend: !!(backendBits.value & 0b100),
    database: !!(backendBits.value & 0b010),
    redis: !!(backendBits.value & 0b001),
    renderer: rendererOk.value,
  }))

  const showBanner = computed(() => {
    if (isDev.value) return false
    return mode.value !== 'healthy'
  })

  // ── WS message handler ────────────────────────────────────
  function handleWSMessage(payload: HealthWSPayload) {
    const c = payload.components ?? {}
    let bits = 0
    if (c.backend) bits |= 0b100
    if (c.database) bits |= 0b010
    if (c.redis) bits |= 0b001
    backendBits.value = bits
    rendererOk.value = c.renderer ?? true
    maintenance.value = payload.maintenance ?? false
    if (typeof payload.isDev === 'boolean') {
      isDev.value = payload.isDev
    }
  }

  // ── Frontend probes ───────────────────────────────────────
  const networkToastShown = ref(false)

  async function runProbes() {
    const [nginx, renderer] = await Promise.all([probeNginx(), probeRenderer()])

    if (nginx && renderer) {
      // Both probes succeeded — update state normally and reset toast flag.
      nginxOk.value = true
      rendererOk.value = true
      networkToastShown.value = false
    } else {
      // Probe failed — keep current state (don't flip bits).
      // Show a one-time toast so the user knows their network may be unstable.
      if (!networkToastShown.value) {
        const { message } = useDiscreteApi()
        message.warning('当前网络连接不太稳定，部分页面加载可能较慢', {
          duration: 5000,
          closable: true,
        })
        networkToastShown.value = true
      }
    }
  }

  async function fetchInitialState() {
    const data = await fetchReadiness()
    if (!data) return
    maintenance.value = data.maintenance ?? false
    if (typeof data.isDev === 'boolean') isDev.value = data.isDev
    const c = data.components ?? []
    let bits = 0b111
    for (const comp of c) {
      if (comp.name === 'backend' && !comp.healthy) bits &= ~0b100
      if (comp.name === 'database' && !comp.healthy) bits &= ~0b010
      if (comp.name === 'redis' && !comp.healthy) bits &= ~0b001
      if (comp.name === 'renderer') rendererOk.value = comp.healthy
    }
    backendBits.value = bits
  }

  function startPolling() {
    if (isDev.value || pollTimer) return
    fetchInitialState()
    runProbes()
    pollTimer = setInterval(runProbes, 30_000)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  return {
    // State
    state,
    mode,
    components,
    showBanner,
    maintenance,
    isDev,
    backendBits,
    nginxOk,
    rendererOk,
    // Actions
    handleWSMessage,
    startPolling,
    stopPolling,
  }
})
