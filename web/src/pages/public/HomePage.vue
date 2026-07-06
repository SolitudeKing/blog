<template>
  <section class="home-page">
    <div class="home-page__hero">
      <div class="home-page__intro">
        <p class="home-page__eyebrow">{{ setting.lobby?.author ?? 'Solitude King' }}</p>
        <h1>{{ setting.lobby?.site_name ?? 'Solitude Blog' }}</h1>
        <p>{{ setting.lobby?.essay ?? 'Keep writing, keep shipping.' }}</p>
      </div>
      <div class="home-page__stats" aria-label="文章概览">
        <strong>{{ articles.length }}</strong>
        <span>Published notes</span>
      </div>
    </div>

    <aside v-if="activeNotice" class="notice-banner">
      <strong>{{ activeNotice.title }}</strong>
      <p>{{ activeNotice.content }}</p>
    </aside>

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
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { getArticleList } from '@/api/modules/article'
import { getActiveNotice } from '@/api/modules/notice'
import { useSettingStore } from '@/stores/setting'
import type { CursorPage } from '@/api/types'
import type { ArticleListItem } from '@/types/article'
import type { NoticeItem } from '@/types/notice'

const setting = useSettingStore()
const articles = ref<ArticleListItem[]>([])
const activeNotice = ref<NoticeItem | null>(null)
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 20,
  has_more: false,
})

onMounted(async () => {
  await setting.loadLobby()
  await loadNotice()
  await loadArticles()
})

async function loadNotice() {
  try {
    activeNotice.value = await getActiveNotice()
  } catch {
    activeNotice.value = null
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
