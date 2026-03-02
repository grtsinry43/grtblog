import { onUnmounted, ref, type Ref } from 'vue'

type TurnstileApi = {
  render: (container: HTMLElement, options: Record<string, unknown>) => string
  remove: (widgetId: string) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileApi
  }
}

let turnstileScriptPromise: Promise<void> | null = null

function loadTurnstileScript(): Promise<void> {
  if (window.turnstile && typeof window.turnstile.render === 'function') return Promise.resolve()
  if (turnstileScriptPromise) return turnstileScriptPromise

  turnstileScriptPromise = new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Turnstile 脚本加载失败'))
    document.head.appendChild(script)
  })

  return turnstileScriptPromise
}

function waitForTurnstile(timeoutMs = 3000): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.turnstile && typeof window.turnstile.render === 'function') {
      resolve()
      return
    }
    const start = Date.now()
    const timer = window.setInterval(() => {
      if (window.turnstile && typeof window.turnstile.render === 'function') {
        window.clearInterval(timer)
        resolve()
        return
      }
      if (Date.now() - start > timeoutMs) {
        window.clearInterval(timer)
        reject(new Error('Turnstile API 未就绪'))
      }
    }, 50)
  })
}

export function useTurnstile(containerRef: Ref<HTMLElement | null>, siteKey: Ref<string>) {
  const token = ref('')
  const error = ref('')
  const expired = ref(false)

  let widgetId: string | null = null
  let destroyed = false
  let renderVersion = 0

  function cleanup() {
    if (widgetId && window.turnstile) {
      window.turnstile.remove(widgetId)
    }
    widgetId = null
    if (containerRef.value) {
      containerRef.value.innerHTML = ''
    }
  }

  async function render() {
    const version = ++renderVersion
    const key = siteKey.value
    const container = containerRef.value

    if (!key || !container) {
      cleanup()
      return
    }

    try {
      await loadTurnstileScript()
      await waitForTurnstile()

      if (destroyed || version !== renderVersion || !window.turnstile) return

      cleanup()

      widgetId = window.turnstile.render(container, {
        sitekey: key,
        theme: 'auto',
        size: 'normal',
        callback: (t: string) => {
          token.value = t
          expired.value = false
          error.value = ''
        },
        'expired-callback': () => {
          token.value = ''
          expired.value = true
        },
        'error-callback': () => {
          token.value = ''
          error.value = 'Turnstile 验证失败'
        },
      })
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Turnstile 加载失败'
    }
  }

  onUnmounted(() => {
    destroyed = true
    cleanup()
  })

  return { token, error, expired, render, cleanup }
}
