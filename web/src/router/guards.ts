import type { Router } from 'vue-router'
import { AUTH_UNAUTHORIZED_EVENT } from '@/api/session'
import { useAuthStore } from '@/stores/auth'

const ADMIN_HOME = '/admin'

export function resolveAdminRedirect(value: unknown) {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) {
    return ADMIN_HOME
  }

  try {
    const target = new URL(value, 'https://solitude.local')
    const isAdminPath = /^\/admin(?:\/|$)/.test(target.pathname)
    const isLoginPath = /^\/admin\/login(?:\/|$)/.test(target.pathname)
    if (target.origin !== 'https://solitude.local' || !isAdminPath || isLoginPath) {
      return ADMIN_HOME
    }
    return `${target.pathname}${target.search}${target.hash}`
  } catch {
    return ADMIN_HOME
  }
}

export function setupRouterGuards(router: Router) {
  router.beforeEach(async (to) => {
    const auth = useAuthStore()
    const isLoginRoute = to.name === 'admin-login'
    const needsAuthentication = to.matched.some((record) => record.meta.requiresAuth)

    if (isLoginRoute) {
      if (!auth.hasSession) {
        return true
      }

      try {
        if (await auth.restoreSession()) {
          return resolveAdminRedirect(to.query.redirect)
        }
      } catch {
        // A transient API failure should not make the login page unreachable.
      }
      return true
    }

    if (!needsAuthentication) {
      return true
    }

    if (!auth.hasSession) {
      return {
        name: 'admin-login',
        query: { redirect: to.fullPath },
      }
    }

    try {
      if (await auth.restoreSession()) {
        return true
      }
    } catch {
      // Keep the shell reachable on transient non-authentication failures.
      return true
    }

    return {
      name: 'admin-login',
      query: { redirect: to.fullPath },
    }
  })

  window.addEventListener(AUTH_UNAUTHORIZED_EVENT, () => {
    const auth = useAuthStore()
    auth.clearSession()

    const currentRoute = router.currentRoute.value
    const isProtectedRoute = currentRoute.matched.some((record) => record.meta.requiresAuth)
    if (!isProtectedRoute || currentRoute.name === 'admin-login') {
      return
    }

    void router.replace({
      name: 'admin-login',
      query: { redirect: resolveAdminRedirect(currentRoute.fullPath) },
    })
  })
}
