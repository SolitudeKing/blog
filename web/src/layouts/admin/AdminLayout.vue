<template>
  <div
    class="admin-layout mist-page mist-page--admin"
    :class="{
      'is-shell-collapsed': shellCollapsed && !isNarrow,
      'is-drawer-open': drawerOpen && isNarrow,
    }"
  >
    <a class="admin-layout__skip-link" href="#admin-main">跳到主内容</a>

    <Teleport to="body" :disabled="!isNarrow">
      <aside
        id="admin-sidebar"
        ref="sidebarPanel"
        class="admin-layout__sidebar"
        :class="{ 'is-open': drawerOpen && isNarrow }"
        :role="isNarrow ? 'dialog' : undefined"
        :aria-modal="isNarrow && drawerOpen ? true : undefined"
        :aria-labelledby="isNarrow ? 'admin-nav-title' : undefined"
        :aria-label="isNarrow ? undefined : '后台导航'"
        :aria-hidden="isNarrow ? !drawerOpen : undefined"
        :inert="isNarrow && !drawerOpen"
      >
        <div class="admin-layout__sidebar-header">
          <RouterLink
            class="admin-layout__brand"
            to="/admin"
            aria-label="Solitude 创作工作台首页"
            :tabindex="navigationTabIndex"
            @click="onNavigationClick"
          >
            <span class="admin-layout__brand-mark" aria-hidden="true">
              <SvgIcon name="brand-waves" />
            </span>
            <span class="admin-layout__brand-copy">
              <strong id="admin-nav-title">Solitude</strong>
              <small>创作工作台</small>
            </span>
          </RouterLink>

          <button
            v-if="!isNarrow"
            class="admin-layout__sidebar-toggle"
            type="button"
            aria-controls="admin-sidebar"
            :aria-expanded="!shellCollapsed"
            :aria-label="navigationToggleLabel"
            :title="navigationToggleLabel"
            @click="toggleNavigation"
          >
            <SvgIcon :name="shellCollapsed ? 'sidebar-expand' : 'sidebar-collapse'" />
          </button>

          <button
            v-else
            ref="closeButton"
            class="admin-layout__close-button"
            type="button"
            :tabindex="navigationTabIndex"
            aria-label="关闭导航菜单"
            @click="closeDrawer()"
          >
            <SvgIcon name="close" />
          </button>
        </div>

        <nav class="admin-layout__navigation" aria-label="后台主导航">
          <section v-for="group in navigationGroups" :key="group.label" class="admin-layout__nav-group">
            <p class="admin-layout__sidebar-title">{{ group.label }}</p>
            <RouterLink
              v-for="item in group.items"
              :key="item.to"
              class="admin-layout__nav-item"
              :class="{ 'is-active': isNavigationActive(item.to) }"
              :to="item.to"
              :aria-current="isNavigationActive(item.to) ? 'page' : undefined"
              :aria-label="shellCollapsed && !isNarrow ? item.label : undefined"
              :title="shellCollapsed && !isNarrow ? item.label : undefined"
              :tabindex="navigationTabIndex"
              @click="onNavigationClick"
            >
              <SvgIcon :name="item.icon" />
              <span class="admin-layout__nav-label">{{ item.label }}</span>
            </RouterLink>
          </section>
        </nav>

        <div class="admin-layout__sidebar-footer">
          <RouterLink
            class="admin-layout__view-site"
            to="/"
            target="_blank"
            rel="noopener"
            :aria-label="shellCollapsed && !isNarrow ? '在新窗口查看博客' : undefined"
            :title="shellCollapsed && !isNarrow ? '查看博客' : undefined"
            :tabindex="navigationTabIndex"
            @click="onNavigationClick"
          >
            <SvgIcon name="external-link" />
            <span class="admin-layout__view-site-label">查看博客</span>
          </RouterLink>

          <div v-if="auth.user" class="admin-layout__account">
            <span class="admin-layout__user-avatar" aria-hidden="true">
              {{ (auth.user.username || 'A').slice(0, 1).toUpperCase() }}
            </span>
            <span class="admin-layout__account-copy">
              <strong>{{ auth.user.username }}</strong>
              <small>站点管理员</small>
            </span>
            <button
              class="admin-layout__account-action"
              type="button"
              :tabindex="navigationTabIndex"
              aria-label="退出登录"
              title="退出登录"
              @click="onLogout"
            >
              <SvgIcon name="logout" />
            </button>
          </div>
        </div>
      </aside>
    </Teleport>

    <Transition name="admin-scrim">
      <Teleport v-if="isNarrow && drawerOpen" to="body">
        <div
          class="admin-layout__scrim"
          aria-hidden="true"
          @click="closeDrawer()"
        />
      </Teleport>
    </Transition>

    <header class="admin-layout__topbar" :inert="isNarrow && drawerOpen">
      <div class="admin-layout__topbar-title">
        <button
          v-if="isNarrow"
          ref="menuButton"
          class="admin-layout__menu-button"
          type="button"
          aria-controls="admin-sidebar"
          :aria-expanded="drawerOpen"
          :aria-label="navigationToggleLabel"
          :title="navigationToggleLabel"
          @click="toggleNavigation"
        >
            <SvgIcon name="sidebar" />
        </button>
        <div>
          <span>{{ currentPage.kicker }}</span>
          <strong>{{ currentPage.title }}</strong>
        </div>
      </div>

      <div class="admin-layout__topbar-actions">
        <span class="admin-layout__health" :class="{ 'is-offline': !isOnline }" role="status">
          <i aria-hidden="true" />{{ isOnline ? '网络正常' : '网络离线' }}
        </span>
        <BaseThemeControls class="admin-layout__theme-controls" size="sm" />
      </div>
    </header>

    <main
      id="admin-main"
      ref="mainContent"
      class="admin-layout__main"
      tabindex="-1"
      :inert="isNarrow && drawerOpen"
    >
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseThemeControls from '@/components/base/BaseThemeControls.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import type { SvgIconName } from '@/config/svgIcons'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const SHELL_STORAGE_KEY = 'blog:admin-shell'
const SHELL_MEDIA_QUERY = '(max-width: 959px)'
const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const shellCollapsed = ref(false)
const drawerOpen = ref(false)
const isNarrow = ref(false)
const menuButton = ref<HTMLButtonElement | null>(null)
const closeButton = ref<HTMLButtonElement | null>(null)
const sidebarPanel = ref<HTMLElement | null>(null)
const mainContent = ref<HTMLElement | null>(null)
const isOnline = ref(true)

let shellMedia: MediaQueryList | null = null
let drawerTrigger: HTMLElement | null = null
let bodyLock:
  | {
      scrollY: number
      overflow: string
      position: string
      top: string
      width: string
    }
  | null = null

const navigationTabIndex = computed(() => (isNarrow.value && !drawerOpen.value ? -1 : undefined))
const navigationToggleLabel = computed(() => {
  if (isNarrow.value) {
    return drawerOpen.value ? '关闭导航菜单' : '打开导航菜单'
  }
  return shellCollapsed.value ? '展开后台侧栏' : '折叠后台侧栏'
})
const pageTitles: Record<string, { kicker: string; title: string }> = {
  'admin-dashboard': { kicker: 'Dashboard', title: '内容概览' },
  'admin-articles': { kicker: 'Articles', title: '文章管理' },
  'admin-article-new': { kicker: 'Editor', title: '新建文章' },
  'admin-article-edit': { kicker: 'Editor', title: '编辑文章' },
  'admin-taxonomy': { kicker: 'Taxonomy', title: '专题与标签' },
  'admin-media': { kicker: 'Media', title: '媒体库' },
  'admin-notices': { kicker: 'Notices', title: '公告管理' },
  'admin-settings': { kicker: 'Settings', title: '站点设置' },
}
const currentPage = computed(() =>
  pageTitles[String(route.name ?? '')] ?? { kicker: 'Admin', title: '创作工作台' },
)
const navigationGroups: Array<{
  label: string
  items: Array<{ to: string; label: string; icon: SvgIconName }>
}> = [
  {
    label: '工作台',
    items: [{ to: '/admin', label: '仪表盘', icon: 'dashboard' }],
  },
  {
    label: '内容',
    items: [
      { to: '/admin/articles', label: '文章管理', icon: 'article' },
      { to: '/admin/taxonomy', label: '专题与标签', icon: 'topic-grid' },
      { to: '/admin/media', label: '媒体库', icon: 'media' },
      { to: '/admin/notices', label: '公告管理', icon: 'empty-notice' },
    ],
  },
  {
    label: '系统',
    items: [{ to: '/admin/settings', label: '站点设置', icon: 'settings' }],
  },
]

onMounted(() => {
  shellMedia = window.matchMedia(SHELL_MEDIA_QUERY)
  syncViewport(shellMedia)
  shellMedia.addEventListener('change', syncViewport)
  isOnline.value = window.navigator.onLine
  window.addEventListener('online', syncOnlineState)
  window.addEventListener('offline', syncOnlineState)

  const storedPreference = readShellPreference()
  if (storedPreference === 'collapsed' || storedPreference === 'expanded') {
    shellCollapsed.value = storedPreference === 'collapsed'
  } else {
    shellCollapsed.value = false
  }
})

onBeforeUnmount(() => {
  shellMedia?.removeEventListener('change', syncViewport)
  document.removeEventListener('keydown', onDocumentKeydown)
  window.removeEventListener('online', syncOnlineState)
  window.removeEventListener('offline', syncOnlineState)
  unlockBodyScroll()
})

watch(
  () => route.fullPath,
  async () => {
    if (drawerOpen.value) {
      closeDrawer(false)
    }
    await nextTick()
    mainContent.value?.focus({ preventScroll: true })
  },
)

function syncViewport(event: MediaQueryList | MediaQueryListEvent) {
  isNarrow.value = event.matches
  if (!event.matches && drawerOpen.value) {
    closeDrawer(false)
  }
}

function syncOnlineState() {
  isOnline.value = window.navigator.onLine
}

function toggleNavigation() {
  if (isNarrow.value) {
    if (drawerOpen.value) {
      closeDrawer()
    } else {
      openDrawer()
    }
    return
  }

  shellCollapsed.value = !shellCollapsed.value
  persistShellPreference(shellCollapsed.value)
}

function persistShellPreference(collapsed: boolean) {
  try {
    localStorage.setItem(SHELL_STORAGE_KEY, collapsed ? 'collapsed' : 'expanded')
  } catch {
    // The layout remains usable when storage is blocked or unavailable.
  }
}

function readShellPreference() {
  try {
    return localStorage.getItem(SHELL_STORAGE_KEY)
  } catch {
    return null
  }
}

function openDrawer() {
  if (!isNarrow.value || drawerOpen.value) {
    return
  }
  drawerTrigger = document.activeElement instanceof HTMLElement
    ? document.activeElement
    : menuButton.value
  drawerOpen.value = true
  lockBodyScroll()
  document.addEventListener('keydown', onDocumentKeydown)
  nextTick(() => closeButton.value?.focus())
}

function closeDrawer(restoreFocus = true) {
  if (!drawerOpen.value) {
    return
  }
  drawerOpen.value = false
  document.removeEventListener('keydown', onDocumentKeydown)
  unlockBodyScroll()

  if (!restoreFocus) {
    drawerTrigger = null
    return
  }

  const target = drawerTrigger?.isConnected ? drawerTrigger : menuButton.value
  drawerTrigger = null
  nextTick(() => target?.focus({ preventScroll: true }))
}

function onNavigationClick() {
  if (isNarrow.value) {
    closeDrawer()
  }
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (!drawerOpen.value) {
    return
  }
  if (event.key === 'Escape') {
    event.preventDefault()
    closeDrawer()
    return
  }
  if (event.key === 'Tab') {
    trapDrawerFocus(event)
  }
}

function trapDrawerFocus(event: KeyboardEvent) {
  const panel = sidebarPanel.value
  if (!panel) {
    return
  }
  const focusable = Array.from(panel.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
    .filter((element) => !element.hasAttribute('disabled') && element.getAttribute('aria-hidden') !== 'true')
  if (!focusable.length) {
    event.preventDefault()
    panel.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement
  if (event.shiftKey && (active === first || !panel.contains(active))) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && (active === last || !panel.contains(active))) {
    event.preventDefault()
    first.focus()
  }
}

function lockBodyScroll() {
  if (bodyLock) {
    return
  }
  const body = document.body
  bodyLock = {
    scrollY: window.scrollY,
    overflow: body.style.overflow,
    position: body.style.position,
    top: body.style.top,
    width: body.style.width,
  }
  body.style.overflow = 'hidden'
  body.style.position = 'fixed'
  body.style.top = '-' + bodyLock.scrollY + 'px'
  body.style.width = '100%'
}

function unlockBodyScroll() {
  if (!bodyLock) {
    return
  }
  const body = document.body
  const snapshot = bodyLock
  bodyLock = null
  body.style.overflow = snapshot.overflow
  body.style.position = snapshot.position
  body.style.top = snapshot.top
  body.style.width = snapshot.width
  window.scrollTo({ top: snapshot.scrollY, left: 0, behavior: 'auto' })
}

function isNavigationActive(to: string) {
  if (to === '/admin') {
    return route.path === to
  }
  return route.path === to || route.path.startsWith(to + '/')
}

async function onLogout() {
  if (drawerOpen.value) {
    closeDrawer(false)
  }
  auth.logout()
  toast.info('已退出登录')
  await router.push('/admin/login')
}
</script>
