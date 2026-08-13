<template>
  <header class="blog-navbar">
    <div class="blog-navbar__inner">
      <RouterLink class="blog-navbar__brand" to="/" :aria-label="`${siteName}，返回首页`">
        <span class="blog-navbar__brand-mark" aria-hidden="true">
          <SvgIcon name="brand-waves" />
        </span>
        <span class="blog-navbar__brand-copy">
          <strong class="blog-navbar__brand-text">{{ siteName }}</strong>
          <span class="blog-navbar__brand-byline">{{ author }}</span>
        </span>
      </RouterLink>

      <nav class="blog-navbar__nav" aria-label="主导航">
        <RouterLink v-for="item in navItems" :key="item.to" :to="item.to">
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="blog-navbar__actions">
        <BaseThemeControls class="blog-navbar__preferences" size="sm" />

        <button
          ref="menuButtonRef"
          class="navbar-hamburger"
          type="button"
          aria-label="打开主导航"
          aria-controls="public-navigation-drawer"
          :aria-expanded="drawerOpen"
          @click="openDrawer"
        >
          <span aria-hidden="true" />
        </button>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="navbar-drawer">
        <div
          v-if="drawerOpen"
          id="public-navigation-drawer"
          class="navbar-drawer"
          role="dialog"
          aria-modal="true"
          aria-labelledby="public-navigation-title"
          @click.self="closeDrawer()"
          @keydown="handleDrawerKeydown"
        >
          <section ref="drawerPanelRef" class="navbar-drawer__panel" tabindex="-1">
            <header class="navbar-drawer__header">
              <div>
                <p>{{ siteName }}</p>
                <div id="public-navigation-title" role="heading" aria-level="2">站点导航</div>
              </div>
              <button
                ref="closeButtonRef"
                class="navbar-drawer__close"
                type="button"
                aria-label="关闭主导航"
                @click="closeDrawer()"
              >
                <SvgIcon name="close" />
              </button>
            </header>

            <nav class="navbar-drawer__nav" aria-label="移动端主导航">
              <RouterLink v-for="item in navItems" :key="item.to" :to="item.to" @click="closeDrawer()">
                {{ item.label }}
              </RouterLink>
            </nav>

            <BaseThemeControls class="navbar-drawer__preferences" size="md" />
          </section>
        </div>
      </Transition>
    </Teleport>
  </header>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import BaseThemeControls from '@/components/base/BaseThemeControls.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { useSettingStore } from '@/stores/setting'

const navItems = [
  { label: '首页', to: '/' },
  { label: '归档', to: '/archives' },
  { label: '搜索', to: '/search' },
  { label: '关于', to: '/about' },
] as const

const drawerOpen = ref(false)
const menuButtonRef = ref<HTMLButtonElement | null>(null)
const closeButtonRef = ref<HTMLButtonElement | null>(null)
const drawerPanelRef = ref<HTMLElement | null>(null)
const route = useRoute()
const setting = useSettingStore()

const siteName = computed(() => setting.lobby?.site_name?.trim() || 'Solitude Blog')
const author = computed(() => setting.lobby?.author?.trim() || 'Solitude King')

let lockedScrollY = 0
let desktopMedia: MediaQueryList | null = null
let previousBodyStyle: Partial<Record<'position' | 'top' | 'left' | 'right' | 'width' | 'overflow', string>> = {}

function lockPage() {
  const { body } = document
  lockedScrollY = window.scrollY
  previousBodyStyle = {
    position: body.style.position,
    top: body.style.top,
    left: body.style.left,
    right: body.style.right,
    width: body.style.width,
    overflow: body.style.overflow,
  }
  body.style.position = 'fixed'
  body.style.top = `-${lockedScrollY}px`
  body.style.left = '0'
  body.style.right = '0'
  body.style.width = '100%'
  body.style.overflow = 'hidden'
  const app = document.getElementById('app')
  if (app) {
    app.inert = true
  }
}

function unlockPage() {
  const { body } = document
  body.style.position = previousBodyStyle.position ?? ''
  body.style.top = previousBodyStyle.top ?? ''
  body.style.left = previousBodyStyle.left ?? ''
  body.style.right = previousBodyStyle.right ?? ''
  body.style.width = previousBodyStyle.width ?? ''
  body.style.overflow = previousBodyStyle.overflow ?? ''
  const app = document.getElementById('app')
  if (app) {
    app.inert = false
  }
  window.scrollTo(0, lockedScrollY)
}

async function openDrawer() {
  if (drawerOpen.value) {
    return
  }
  lockPage()
  drawerOpen.value = true
  await nextTick()
  closeButtonRef.value?.focus()
}

async function closeDrawer(restoreFocus = true) {
  if (!drawerOpen.value) {
    return
  }
  drawerOpen.value = false
  unlockPage()
  if (restoreFocus) {
    await nextTick()
    menuButtonRef.value?.focus({ preventScroll: true })
  }
}

function handleDrawerKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    void closeDrawer()
    return
  }
  if (event.key !== 'Tab' || !drawerPanelRef.value) {
    return
  }
  const focusable = Array.from(
    drawerPanelRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => !element.hasAttribute('hidden'))
  if (!focusable.length) {
    event.preventDefault()
    drawerPanelRef.value.focus()
    return
  }
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

function handleDesktopChange() {
  if (desktopMedia?.matches) {
    void closeDrawer(false)
  }
}

watch(
  () => route.fullPath,
  () => {
    void closeDrawer(false)
  },
)

onMounted(() => {
  desktopMedia = window.matchMedia('(min-width: 768px)')
  desktopMedia.addEventListener('change', handleDesktopChange)
})

onBeforeUnmount(() => {
  desktopMedia?.removeEventListener('change', handleDesktopChange)
  if (drawerOpen.value) {
    drawerOpen.value = false
    unlockPage()
  }
})
</script>
