<template>
  <section class="home-page">
    <div class="home-page__shell">
      <aside class="home-profile">
        <div class="home-profile__identity">
          <div class="home-profile__avatar" aria-hidden="true">{{ authorInitial }}</div>
          <div>
            <p class="home-page__eyebrow">{{ setting.lobby?.author ?? 'Solitude King' }}</p>
            <h1>{{ setting.lobby?.site_name ?? 'Solitude Blog' }}</h1>
          </div>
        </div>

        <p class="home-profile__essay">{{ setting.lobby?.essay ?? 'Keep writing, keep shipping.' }}</p>

        <div class="home-profile__metrics" aria-label="站点概览">
          <div>
            <strong>{{ articles.length }}</strong>
            <span>文章</span>
          </div>
          <div>
            <strong>{{ categories.length }}</strong>
            <span>分类</span>
          </div>
          <div>
            <strong>{{ tags.length }}</strong>
            <span>标签</span>
          </div>
        </div>

        <div class="home-profile__progress">
          <div>
            <span>今年进度</span>
            <strong>{{ yearProgress }}%</strong>
          </div>
          <i :style="{ width: `${yearProgress}%` }"></i>
        </div>

        <div class="home-profile__links">
          <RouterLink to="/archives">归档</RouterLink>
          <RouterLink to="/search">搜索</RouterLink>
        </div>
      </aside>

      <main class="home-feed" aria-label="最新文章">
        <aside v-if="activeNotice" class="notice-banner">
          <strong>{{ activeNotice.title }}</strong>
          <p>{{ activeNotice.content }}</p>
        </aside>

        <div class="home-feed__header">
          <div>
            <p class="home-page__eyebrow">Latest</p>
            <h2>最新文章</h2>
          </div>
          <span>{{ articles.length }} posts</span>
        </div>

        <div v-if="articles.length" class="home-page__list" aria-label="文章列表">
          <ArticleCard v-for="article in articles" :key="article.id" :article="article" />
        </div>

        <div v-else-if="loading" class="page-state">文章加载中</div>
        <div v-else-if="error" class="page-state page-state--error">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" @click="reload">重试</BaseButton>
        </div>
        <div v-else class="page-state">暂时还没有发布文章</div>

        <div v-if="articles.length" class="home-page__footer">
          <BaseButton v-if="page.has_more" variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
          <span v-else class="home-page__end">已经到底了</span>
        </div>
      </main>

      <aside class="home-aside">
        <section class="home-widget home-widget--quote">
          <div class="home-widget__header">
            <span></span>
            <span></span>
            <span></span>
            <strong>每日摘句</strong>
          </div>
          <p>{{ quote }}</p>
        </section>

        <section class="home-widget">
          <div class="home-widget__title">
            <strong>分类</strong>
            <span>{{ categories.length }}</span>
          </div>
          <div class="home-taxonomy-list">
            <span v-for="category in categories" :key="category.id">{{ category.name }}</span>
            <span v-if="!categories.length">Notes</span>
          </div>
        </section>

        <section class="home-widget">
          <div class="home-widget__title">
            <strong>标签云</strong>
            <span>{{ tags.length }}</span>
          </div>
          <div class="home-tag-cloud">
            <span v-for="tag in tags" :key="tag.id" :style="{ '--tag-color': tag.color || 'var(--accent)' }">
              {{ tag.name }}
            </span>
            <span v-if="!tags.length">Markdown</span>
          </div>
        </section>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { getArticleList } from '@/api/modules/article'
import { getActiveNotice } from '@/api/modules/notice'
import { getCategoryList, getTagList } from '@/api/modules/taxonomy'
import { useSettingStore } from '@/stores/setting'
import type { CursorPage } from '@/api/types'
import type { ArticleListItem } from '@/types/article'
import type { NoticeItem } from '@/types/notice'
import type { CategoryItem, TagItem } from '@/types/taxonomy'

const setting = useSettingStore()
const articles = ref<ArticleListItem[]>([])
const activeNotice = ref<NoticeItem | null>(null)
const categories = ref<CategoryItem[]>([])
const tags = ref<TagItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 20,
  has_more: false,
})

const quote = computed(() => setting.lobby?.essay || '遇事不决，可问春风；春风不语，即随本心。')
const authorInitial = computed(() => (setting.lobby?.author || 'S').trim().slice(0, 1).toUpperCase())
const yearProgress = computed(() => {
  const now = new Date()
  const start = new Date(now.getFullYear(), 0, 1).getTime()
  const end = new Date(now.getFullYear() + 1, 0, 1).getTime()
  return Math.min(100, Math.max(0, Math.round(((now.getTime() - start) / (end - start)) * 100)))
})

onMounted(async () => {
  await setting.loadLobby()
  await Promise.all([loadNotice(), loadTaxonomy()])
  await loadArticles()
})

async function loadNotice() {
  try {
    activeNotice.value = await getActiveNotice()
  } catch {
    activeNotice.value = null
  }
}

async function loadTaxonomy() {
  try {
    const [categoryItems, tagItems] = await Promise.all([getCategoryList(), getTagList()])
    categories.value = categoryItems
    tags.value = tagItems
  } catch {
    categories.value = []
    tags.value = []
  }
}

async function reload() {
  articles.value = []
  page.cursor = ''
  page.next_cursor = ''
  page.has_more = false
  await loadArticles()
}

async function loadArticles(cursor = '') {
  loading.value = true
  error.value = ''
  try {
    const result = await getArticleList({ cursor: cursor || undefined, limit: page.limit })
    articles.value = cursor ? [...articles.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!page.next_cursor) {
    return
  }
  loadingMore.value = true
  try {
    await loadArticles(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}

</script>
