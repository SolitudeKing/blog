import { createRouter, createWebHistory } from 'vue-router'
import { publicRoutes } from './public.routes'
import { adminRoutes } from './admin.routes'
import { setupRouterGuards } from './guards'

export const router = createRouter({
  history: createWebHistory(),
  routes: [...publicRoutes, ...adminRoutes],
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    if (to.hash) {
      return { el: to.hash, top: 88 }
    }
    return { top: 0 }
  },
})

setupRouterGuards(router)
