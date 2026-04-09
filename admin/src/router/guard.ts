import { isEmpty } from 'lodash-es'

import { useEventBus } from '@/event-bus'
import { getSetupState } from '@/services/auth'
import { useUserStore, toRefsUserStore } from '@/stores'
import {
  applyDocumentTitle,
  ensureBackendSiteName,
  getCachedSiteName,
} from '@/utils/document-title'

import type { Router } from 'vue-router'

const Layout = () => import('@/layout/index.vue')

const SETUP_STATE_TTL_MS = 5000
let setupStateCache: Awaited<ReturnType<typeof getSetupState>> | null = null
let setupStateCachedAt = 0

async function getCachedSetupState(force = false) {
  const now = Date.now()
  if (!force && setupStateCache && now - setupStateCachedAt < SETUP_STATE_TTL_MS) {
    return setupStateCache
  }
  const state = await getSetupState()
  setupStateCache = state
  setupStateCachedAt = now
  return state
}

// Reset when user logs out so the guide is re-checked on next login.
let upgradeGuideChecked = false
export function resetUpgradeGuideCheck() {
  upgradeGuideChecked = false
}

export function setupRouterGuard(router: Router) {
  const { resolveMenuRoute, cleanup, refreshAccessInfo } = useUserStore()

  const { token, user } = toRefsUserStore()
  const { routerEventBus } = useEventBus()
  router.beforeEach(async (to, _from, next) => {
    routerEventBus.emit('beforeEach')

    if (to.name === 'init') {
      if (token.value) {
        next({ path: '/' })
        return false
      }
      try {
        const setupState = await getCachedSetupState()
        if (!setupState.needsSetup) {
          next({ name: 'signIn' })
          return false
        }
      } catch (error) {
        console.error('Error checking setup state:', error)
      }
      next()
      return false
    }

    if (to.name === 'signIn') {
      try {
        const setupState = await getCachedSetupState()
        if (setupState.needsSetup && !setupState.hasUser) {
          next({ name: 'init' })
          return false
        }
      } catch (error) {
        console.error('Error checking setup state:', error)
      }
      if (!token.value) {
        next()
      } else {
        next({ path: '/' })
      }

      return false
    }

    // Allow upgrade guide page through if user has token
    if (to.name === 'upgradeGuide') {
      if (!token.value) {
        next({ name: 'signIn' })
        return false
      }
      if (user.value.id === null) {
        try {
          await refreshAccessInfo()
        } catch {
          cleanup()
          next({ name: 'signIn' })
          return false
        }
      }
      next()
      return false
    }

    if (!token.value) {
      try {
        const setupState = await getCachedSetupState()
        if (setupState.needsSetup) {
          next({ name: 'init' })
          return false
        }
      } catch (error) {
        console.error('Error checking setup state:', error)
      }
      next({
        name: 'signIn',
        query: {
          r: to.fullPath,
        },
      })
      return false
    }

    if (token.value && user.value.id === null) {
      try {
        await refreshAccessInfo()
      } catch (error) {
        console.error('Error refreshing user access info:', error)
        cleanup()
        next()
        return false
      }
    }

    // Check upgrade guide once per session after login.
    // Set flag before await to prevent concurrent navigations from double-redirecting.
    // Reset on error so it retries on the next navigation.
    if (token.value && !upgradeGuideChecked) {
      upgradeGuideChecked = true
      try {
        const setupState = await getCachedSetupState(true)
        if (setupState.pendingUpgradeGuides?.length > 0 && user.value.isAdmin) {
          next({ name: 'upgradeGuide' })
          return false
        }
      } catch (error) {
        upgradeGuideChecked = false
        console.error('Error checking upgrade guide state:', error)
      }
    }

    if (token.value && !router.hasRoute('layout')) {
      try {
        const { routeList } = await resolveMenuRoute()

        if (isEmpty(routeList)) {
          cleanup()
          next()
          return false
        }

        router.addRoute({
          path: '/',
          name: 'layout',
          component: Layout,
          // if you need to have a redirect when accessing / routing
          redirect: '/dashboard',
          children: routeList,
        })

        next(to.fullPath)
      } catch (error) {
        console.error('Error resolving user menu or adding route:', error)
        cleanup()
        next()
      }

      return false
    }

    next()
    return false
  })

  router.beforeResolve((_, __, next) => {
    next()
  })

  router.afterEach((to) => {
    routerEventBus.emit('afterEach')
    applyDocumentTitle(to, getCachedSiteName())

    const routePath = to.fullPath
    void ensureBackendSiteName().then((siteName) => {
      if (router.currentRoute.value.fullPath !== routePath) return
      applyDocumentTitle(to, siteName)
    })
  })
}
