import { isEmpty } from 'lodash-es'

import { useEventBus } from '@/event-bus'
import { getSetupState } from '@/services/auth'
import { useUserStore, toRefsUserStore } from '@/stores'
import { isFederationBetaRoute, showFederationBetaDialog } from '@/utils/federation-beta'
import { isFederationEnabled } from '@/utils/federation-gate'
import { applyDocumentTitle, ensureBackendSiteName, getCachedSiteName } from '@/utils/document-title'

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

export function setupRouterGuard(router: Router) {
  const { resolveMenuRoute, cleanup, refreshAccessInfo } = useUserStore()

  const { token, user } = toRefsUserStore()
  const { routerEventBus } = useEventBus()
  let allowFederationRouteOnce = false

  router.beforeEach(async (to, _from, next) => {
    routerEventBus.emit('beforeEach')

    if (isFederationEnabled && isFederationBetaRoute(to)) {
      if (allowFederationRouteOnce) {
        allowFederationRouteOnce = false
      } else {
        const confirmed = await showFederationBetaDialog()
        if (!confirmed) {
          next(false)
          return false
        }
        allowFederationRouteOnce = true
      }
    }

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
