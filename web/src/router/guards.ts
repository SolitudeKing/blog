import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export function setupRouterGuards(router: Router) {
  router.beforeEach((to) => {
    const auth = useAuthStore()
    if (to.meta.requiresAuth && !auth.accessToken) {
      return {
        name: 'admin-login',
        query: { redirect: to.fullPath },
      }
    }
    return true
  })
}

