<template>
  <div class="public-layout mist-page">
    <a class="skip-link" href="#main-content">跳到主内容</a>
    <BlogNavbar />
    <main id="main-content" ref="mainRef" class="public-layout__main" tabindex="-1">
      <RouterView />
    </main>
    <footer class="public-footer">
      <div class="public-footer__inner">
        <div class="public-footer__line" aria-hidden="true" />
        <div class="public-footer__grid">
          <section class="public-footer__identity" aria-labelledby="footer-brand-name">
            <RouterLink class="public-footer__brand" to="/" :aria-label="`${siteName}，返回首页`">
              <span class="public-footer__brand-mark" aria-hidden="true">
                <svg viewBox="0 0 24 24">
                  <path d="M3 8.5c3-3 6 3 9 0s6 3 9 0M3 15.5c3-3 6 3 9 0s6 3 9 0" />
                </svg>
              </span>
              <span class="public-footer__brand-copy">
                <strong id="footer-brand-name">{{ siteName }}</strong>
                <span>{{ author }}</span>
              </span>
            </RouterLink>
            <p class="public-footer__note">{{ essay }}</p>
          </section>

          <section aria-labelledby="footer-roam-title">
            <h2 id="footer-roam-title" class="public-footer__title">漫游</h2>
            <nav class="public-footer__links" aria-label="页脚导航">
              <RouterLink v-for="item in footerNavigation" :key="item.to" :to="item.to">
                {{ item.label }}
              </RouterLink>
            </nav>
          </section>

          <section aria-labelledby="footer-social-title">
            <h2 id="footer-social-title" class="public-footer__title">社交</h2>
            <nav class="public-footer__links" aria-label="社交链接">
              <a
                v-for="item in socialItems"
                :key="item.key"
                :href="item.url"
                target="_blank"
                rel="noopener noreferrer"
              >
                {{ item.label }}
              </a>
              <span v-if="!socialItems.length" class="public-footer__empty">暂未公开社交链接</span>
              <a href="/rss.xml">RSS</a>
            </nav>
          </section>
        </div>

        <p class="public-footer__copyright">
          © 2020–{{ year }} {{ author }} · {{ siteName }}
        </p>
      </div>
    </footer>

    <button
      class="public-backtop"
      :class="{ 'is-visible': backtopVisible }"
      type="button"
      aria-label="返回页面顶部"
      :aria-hidden="backtopVisible ? undefined : 'true'"
      :tabindex="backtopVisible ? 0 : -1"
      @click="backToTop"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="m6 14 6-6 6 6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import BlogNavbar from '@/components/blog/BlogNavbar.vue'
import { useSettingStore } from '@/stores/setting'

const BACKTOP_THRESHOLD = 720
const footerNavigation = [
  { label: '首页', to: '/' },
  { label: '全部文章', to: '/archives' },
  { label: '搜索', to: '/search' },
  { label: '关于我', to: '/about' },
] as const
const socialLabelMap: Record<string, string> = {
  gitee: 'Gitee',
  github: 'GitHub',
  bilibili: 'Bilibili',
  douyin: 'Douyin',
}

const setting = useSettingStore()
const route = useRoute()
const mainRef = ref<HTMLElement | null>(null)
const backtopVisible = ref(false)
let backtopFocusTimer: number | null = null

const siteName = computed(() => setting.lobby?.site_name?.trim() || 'Solitude Blog')
const author = computed(() => setting.lobby?.author?.trim() || 'Solitude King')
const essay = computed(
  () =>
    setting.lobby?.essay?.trim() ||
    '关于设计、代码与缓慢生活的长期笔记。保持好奇，也保持边界。',
)
const social = computed(() => setting.lobby?.social_links ?? {})
const socialItems = computed(() =>
  Object.entries(social.value)
    .filter(([, url]) => url.trim().length > 0)
    .map(([key, url]) => ({
      key,
      url,
      label: socialLabelMap[key] ?? key,
    })),
)
const year = new Date().getFullYear()

function focusMain() {
  mainRef.value?.focus({ preventScroll: true })
}

function updateBacktop() {
  backtopVisible.value = window.scrollY > BACKTOP_THRESHOLD
}

function backToTop() {
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  window.scrollTo({ top: 0, behavior: reducedMotion ? 'auto' : 'smooth' })

  if (backtopFocusTimer !== null) {
    window.clearTimeout(backtopFocusTimer)
  }
  if (reducedMotion) {
    focusMain()
    return
  }
  backtopFocusTimer = window.setTimeout(() => {
    backtopFocusTimer = null
    focusMain()
  }, 420)
}

watch(
  () => route.fullPath,
  async () => {
    await nextTick()
    focusMain()
  },
)

onMounted(() => {
  window.addEventListener('scroll', updateBacktop, { passive: true })
  updateBacktop()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateBacktop)
  if (backtopFocusTimer !== null) {
    window.clearTimeout(backtopFocusTimer)
  }
})
</script>
