import { createRouter, createWebHistory } from 'vue-router'

import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/init',
    name: 'init',
    meta: {
      title: '初始化',
    },
    component: () => import('@/views/init/index.vue'),
  },
  {
    path: '/sign-in',
    name: 'signIn',
    meta: {
      title: '登录',
    },
    component: () => import('@/views/sign-in/index.vue'),
  },
  {
    path: '/upgrade-guide',
    name: 'upgradeGuide',
    meta: {
      title: '升级引导',
    },
    component: () => import('@/views/upgrade-guide/index.vue'),
  },
  {
    name: 'errorPage',
    path: '/:pathMatch(.*)*',
    meta: {
      title: '错误页',
    },
    component: () => import('@/views/error-page/index.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  strict: true,
})

export default router
