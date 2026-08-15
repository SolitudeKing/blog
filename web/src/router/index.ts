import { createRouter, createWebHistory } from 'vue-router'
import { publicRoutes } from './public.routes'
import { adminRoutes } from './admin.routes'
import { setupRouterGuards } from './guards'

export const router = createRouter({
  history: createWebHistory(),
  routes: [...publicRoutes, ...adminRoutes],
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    // SPA 启动后的首次导航（含硬刷新）：忽略 URL 上的 hash，强制回到顶部，
    // 避免 `/#latest-posts` 这类残留 hash 在刷新后把视口带回 anchor。
    if (!from.name) {
      return { top: 0 }
    }
    if (to.hash) {
      return { el: to.hash, top: 88 }
    }
    return { top: 0 }
  },
})

setupRouterGuards(router)
