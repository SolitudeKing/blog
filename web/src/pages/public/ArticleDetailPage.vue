<template>
  <div class="article-detail-page-wrapper">
    <div class="reading-progress" aria-hidden="true">
      <div class="reading-progress__bar" :style="{ '--progress': `${progress}%` }" />
    </div>

    <section class="article-detail-page">
      <RouterLink class="article-back-channel" to="/">
        <SvgIcon name="arrow-left" />
        <span>
          <small>Back to stream</small>
          <strong>返回文章列表</strong>
        </span>
      </RouterLink>

      <article v-if="article" ref="articleRef" class="article-detail">
        <header class="article-detail__header">
          <div class="article-detail__headline">
            <p class="article-detail__meta">
              <span v-if="article.topic"><strong>{{ topicLabel }}</strong></span>
              <span>{{ formattedDate }}</span>
            </p>
            <div role="heading" aria-level="1">{{ article.title }}</div>
            <p v-if="article.summary" class="article-detail__summary">{{ article.summary }}</p>

            <div class="article-detail__toolbar">
              <button class="cta-pill" type="button" @click="copyLink">
                <SvgIcon name="link" />
                {{ copyLabel }}
              </button>
            </div>
          </div>

          <aside class="article-detail__facts" aria-label="文章信息">
            <dl>
              <div>
                <dt>阅读时间</dt>
                <dd>{{ readingMinutes }} 分钟</dd>
              </div>
              <div>
                <dt>浏览</dt>
                <dd>{{ article.view_count }}</dd>
              </div>
              <div>
                <dt>标签</dt>
                <dd>{{ article.tags.slice(0, 2).join(' · ') || '随笔' }}</dd>
              </div>
            </dl>
          </aside>
        </header>

        <figure class="article-detail__cover">
          <SvgIcon name="article-cover" />
          <figcaption>{{ topicLabel }} · {{ formattedDate }}</figcaption>
        </figure>

        <div class="article-detail__layout">
          <div class="markdown-body" v-html="rendered.html" />
          <div class="article-detail__sidebar">
            <BlogToc
              id-prefix="desktop-article-toc"
              title="文章目录"
              :items="tocItems"
              :active-id="activeTocId"
              @navigate="handleTocNavigate"
            />
          </div>
        </div>

        <footer class="article-detail__author">
          <span class="article-detail__author-avatar" aria-hidden="true">{{ authorInitial }}</span>
          <div>
            <span>Written by</span>
            <div role="heading" aria-level="2">{{ author }}</div>
            <p>{{ authorEssay }}</p>
          </div>
        </footer>

        <nav class="article-detail__nav" aria-label="文章导航">
          <RouterLink v-if="prev" :to="`/articles/${prev.slug}`">
            <small>上一篇</small>
            <strong>{{ prev.title }}</strong>
          </RouterLink>
          <div v-else class="article-detail__nav-empty">没有更早的文章了</div>

          <RouterLink v-if="next" :to="`/articles/${next.slug}`">
            <small>下一篇</small>
            <strong>{{ next.title }}</strong>
          </RouterLink>
          <div v-else class="article-detail__nav-empty">已是最新</div>
        </nav>
      </article>

      <div v-else-if="loading" class="page-state" aria-busy="true" aria-label="正在加载文章">
        <BaseSkeleton variant="card" :count="3" />
      </div>
      <div v-else-if="error" role="alert">
        <BaseEmpty title="文章加载失败" :description="error" cta-text="返回首页" cta-to="/">
          <template #icon>
            <SvgIcon name="info" />
          </template>
        </BaseEmpty>
      </div>
      <BaseEmpty
        v-else
        title="文章不存在"
        description="链接可能已失效，去首页看看其他文章吧。"
        cta-text="返回首页"
        cta-to="/"
      >
        <template #icon>
          <SvgIcon name="document" />
        </template>
      </BaseEmpty>

      <button
        v-if="article && tocItems.length"
        ref="tocTriggerRef"
        class="toc-fab"
        type="button"
        aria-label="打开目录"
        aria-controls="article-toc-drawer"
        :aria-expanded="tocDrawerOpen"
        @click="openTocDrawer"
      >
        <SvgIcon class="toc-fab__icon" name="list" />
      </button>

      <Teleport to="body">
        <Transition name="toc-drawer">
          <div
            v-if="tocDrawerOpen"
            id="article-toc-drawer"
            class="toc-drawer"
            role="dialog"
            aria-modal="true"
            aria-labelledby="mobile-article-toc-title"
            @click.self="closeTocDrawer()"
            @keydown="handleTocKeydown"
          >
            <section ref="tocPanelRef" class="toc-drawer__panel" tabindex="-1">
              <div class="toc-drawer__handle" aria-hidden="true" />
              <header class="toc-drawer__header">
                <div id="mobile-article-toc-title" role="heading" aria-level="2">文章目录</div>
                <button
                  ref="tocCloseRef"
                  class="toc-drawer__close"
                  type="button"
                  aria-label="关闭文章目录"
                  @click="closeTocDrawer()"
                >
                  <SvgIcon name="close" />
                </button>
              </header>
              <BlogToc
                id-prefix="mobile-article-toc"
                title="目录链接"
                :items="tocItems"
                :active-id="activeTocId"
                @navigate="handleMobileTocNavigate"
              />
            </section>
          </div>
        </Transition>
      </Teleport>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import BlogToc from '@/components/blog/BlogToc.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { getArticleDetail, getArticleList } from '@/api/modules/article'
import type { ArticleDetail, ArticleListItem } from '@/types/article'
import type { BlogTocItem } from '@/types/toc'
import { renderMarkdown } from '@/utils/markdown'
import { useToast } from '@/composables/useToast'
import { useSettingStore } from '@/stores/setting'

const route = useRoute()
const article = ref<ArticleDetail | null>(null)
const allArticles = ref<ArticleListItem[]>([])
const loading = ref(false)
const error = ref('')
const progress = ref(0)
const activeTocId = ref('')
const tocDrawerOpen = ref(false)
const copyLabel = ref('复制链接')
const articleRef = ref<HTMLElement | null>(null)
const tocTriggerRef = ref<HTMLButtonElement | null>(null)
const tocCloseRef = ref<HTMLButtonElement | null>(null)
const tocPanelRef = ref<HTMLElement | null>(null)
const toast = useToast()
const setting = useSettingStore()
let articleRequestId = 0
let copyTimer: ReturnType<typeof setTimeout> | undefined
let articleResizeObserver: ResizeObserver | undefined
let lockedScrollY = 0
let previousBodyStyle: Partial<Record<'position' | 'top' | 'left' | 'right' | 'width' | 'overflow', string>> = {}

const rendered = computed(() => renderMarkdown(article.value?.content_md ?? ''))
const tocItems = computed<BlogTocItem[]>(() =>
  rendered.value.toc.map((item) => ({
    id: item.id,
    label: item.text,
    href: `#${item.id}`,
    level: item.level - 2,
  })),
)

const formattedDate = computed(() => {
  if (!article.value) {
    return ''
  }
  const raw = article.value.published_at || article.value.created_at
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(raw))
})

const readingMinutes = computed(() => {
  if (!article.value) {
    return 0
  }
  const words = article.value.content_md?.length ?? 0
  return Math.max(1, Math.round(words / 400))
})

const author = computed(() => setting.lobby?.author ?? 'Solitude King')
const authorInitial = computed(() => author.value.trim().slice(0, 1).toUpperCase())
const authorEssay = computed(() =>
  setting.lobby?.essay?.trim() || '持续记录工程实践、设计判断与缓慢生长的想法。',
)
const topicLabel = computed(() => article.value?.topic?.label || article.value?.topic?.name || 'NODES')

const prev = computed(() => {
  const slug = article.value?.slug
  if (!slug) {
    return null
  }
  const idx = allArticles.value.findIndex((a) => a.slug === slug)
  if (idx < 0 || idx === allArticles.value.length - 1) {
    return null
  }
  return allArticles.value[idx + 1]
})

const next = computed(() => {
  const slug = article.value?.slug
  if (!slug) {
    return null
  }
  const idx = allArticles.value.findIndex((a) => a.slug === slug)
  if (idx <= 0) {
    return null
  }
  return allArticles.value[idx - 1]
})

function handleScroll() {
  const articleEl = articleRef.value
  if (!articleEl) {
    progress.value = 0
    return
  }
  const rect = articleEl.getBoundingClientRect()
  const articleTop = window.scrollY + rect.top
  const total = rect.height - window.innerHeight
  if (total <= 0) {
    progress.value = 100
  } else {
    const scrolled = Math.min(total, Math.max(0, window.scrollY - articleTop))
    progress.value = Math.round((scrolled / total) * 100)
  }

  const headingOffset = 112
  let currentId = tocItems.value[0]?.id ?? ''
  for (const item of tocItems.value) {
    const heading = document.getElementById(item.id)
    if (!heading || heading.getBoundingClientRect().top > headingOffset) {
      break
    }
    currentId = item.id
  }
  activeTocId.value = currentId
}

async function focusTocTarget(item: BlogTocItem) {
  await nextTick()
  window.requestAnimationFrame(() => {
    const target = document.getElementById(item.id)
    if (!target) {
      return
    }
    target.setAttribute('tabindex', '-1')
    target.focus({ preventScroll: true })
  })
}

function handleTocNavigate(item: BlogTocItem) {
  void focusTocTarget(item)
}

async function handleMobileTocNavigate(item: BlogTocItem) {
  await closeTocDrawer(false)
  await focusTocTarget(item)
}

async function copyLink() {
  if (!article.value) {
    return
  }
  const url = `${window.location.origin}/articles/${article.value.slug}`
  try {
    await navigator.clipboard.writeText(url)
    copyLabel.value = '已复制'
    toast.success('链接已复制到剪贴板')
    if (copyTimer) {
      clearTimeout(copyTimer)
    }
    copyTimer = setTimeout(() => (copyLabel.value = '复制链接'), 2000)
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

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

async function openTocDrawer() {
  if (tocDrawerOpen.value) {
    return
  }
  lockPage()
  tocDrawerOpen.value = true
  await nextTick()
  tocCloseRef.value?.focus()
}

async function closeTocDrawer(restoreFocus = true) {
  if (!tocDrawerOpen.value) {
    return
  }
  tocDrawerOpen.value = false
  unlockPage()
  if (restoreFocus) {
    await nextTick()
    tocTriggerRef.value?.focus({ preventScroll: true })
  }
}

function handleTocKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    void closeTocDrawer()
    return
  }
  if (event.key !== 'Tab' || !tocPanelRef.value) {
    return
  }
  const focusable = Array.from(
    tocPanelRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => !element.hasAttribute('hidden'))
  if (!focusable.length) {
    event.preventDefault()
    tocPanelRef.value.focus()
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

async function loadArticle(slug: string) {
  const requestId = ++articleRequestId
  loading.value = true
  error.value = ''
  article.value = null
  allArticles.value = []
  progress.value = 0
  activeTocId.value = ''
  articleResizeObserver?.disconnect()
  try {
    const [detail, listResult] = await Promise.all([
      getArticleDetail(slug),
      getArticleList({ page: 1, page_size: 200 }).catch(
        () =>
          ({
            data: [] as ArticleListItem[],
            code: 0,
            message: 'ok',
            page: 1,
            page_size: 200,
            count: 0,
            has_more: false,
          }),
      ),
    ])
    if (requestId !== articleRequestId) {
      return
    }
    article.value = detail
    allArticles.value = listResult.data
    await nextTick()
    if (articleRef.value) {
      articleResizeObserver?.observe(articleRef.value)
    }
    handleScroll()
  } catch (err) {
    if (requestId === articleRequestId) {
      error.value = err instanceof Error ? err.message : '加载文章失败'
    }
  } finally {
    if (requestId === articleRequestId) {
      loading.value = false
    }
  }
}

watch(
  () => String(route.params.slug ?? ''),
  async (slug) => {
    await closeTocDrawer(false)
    if (!slug) {
      article.value = null
      error.value = '文章地址无效'
      return
    }
    await loadArticle(slug)
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
  window.addEventListener('resize', handleScroll, { passive: true })
  if ('ResizeObserver' in window) {
    articleResizeObserver = new ResizeObserver(handleScroll)
    if (articleRef.value) {
      articleResizeObserver.observe(articleRef.value)
    }
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('resize', handleScroll)
  articleResizeObserver?.disconnect()
  articleRequestId += 1
  if (copyTimer) {
    clearTimeout(copyTimer)
  }
  if (tocDrawerOpen.value) {
    tocDrawerOpen.value = false
    unlockPage()
  }
})
</script>

<style lang="scss" scoped>
.article-detail-page-wrapper {
  position: relative;
}
</style>
