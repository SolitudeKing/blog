import { createRouter, createWebHistory } from 'vue-router'
import { publicRoutes } from './public.routes'
import { adminRoutes } from './admin.routes'
import { setupRouterGuards } from './guards'

export const router = createRouter({
  history: createWebHistory(),
  routes: [...publicRoutes, ...adminRoutes],
  scrollBehavior() {
    return { top: 0 }
  },
})

setupRouterGuards(router)

