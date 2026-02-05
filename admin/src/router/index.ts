import { createRouter, createWebHistory } from 'vue-router'

import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/init', name: 'init', component: () => import('@/views/init/index.vue') },
  { path: '/sign-in', name: 'signIn', component: () => import('@/views/sign-in/index.vue') },
  {
    name: 'errorPage',
    path: '/:pathMatch(.*)*',
    component: () => import('@/views/error-page/index.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  strict: true,
})

export default router
