import './assets/main.css'

import { VueQueryPlugin } from '@tanstack/vue-query'
import { createApp } from 'vue'

import { setupEventBus } from '@/event-bus'
import { setupRouterGuard } from '@/router/guard'
import { setupApiInterceptors } from '@/services/api-interceptors'
import { pinia } from '@/stores'

import App from './App.vue'
import router from './router'

async function setupApp() {
  const app = createApp(App)

  app.use(pinia)

  setupApiInterceptors()

  app.use(router)

  setupRouterGuard(router)

  setupEventBus()

  app.use(VueQueryPlugin)

  await router.isReady()

  if (window.loaderElement) {
    window.loaderElement.remove()
    window.loaderElement = null
  }

  app.mount('#app')
}

setupApp()
