<template>
  <section class="search-page">
    <header class="search-page__header">
      <p class="home-page__eyebrow">Search</p>
      <h1>搜索文章</h1>
    </header>

    <form class="search-box" @submit.prevent="submitSearch">
      <input v-model.trim="keyword" class="cui-input" placeholder="输入标题、标签或正文关键词" />
      <BaseButton :loading="loading">搜索</BaseButton>
    </form>

    <div v-if="searched && results.length" class="search-results">
      <RouterLink v-for="item in results" :key="item.id" class="search-result" :to="`/articles/${item.slug}`">
        <div>
          <h2>{{ item.title }}</h2>
          <p>
            <template v-for="(part, index) in highlight(item.snippet)" :key="`${item.id}-${index}`">
              <mark v-if="part.hit">{{ part.text }}</mark>
              <span v-else>{{ part.text }}</span>
            </template>
          </p>
        </div>
        <div class="search-result__meta">
          <span>{{ item.category }}</span>
          <span v-for="field in item.matched_fields" :key="field">{{ field }}</span>
        </div>
      </RouterLink>
    </div>

    <div v-else-if="loading" class="page-state">正在搜索</div>
    <div v-else-if="error" class="page-state page-state--error">{{ error }}</div>
    <div v-else-if="searched" class="page-state">没有找到匹配文章</div>
    <div v-else class="page-state">输入关键词开始搜索</div>

    <div v-if="results.length && page.has_more" class="home-page__footer">
      <BaseButton variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import { searchArticles } from '@/api/modules/article'
import type { CursorPage } from '@/api/types'
import type { ArticleSearchItem } from '@/types/article'

interface HighlightPart {
  text: string
  hit: boolean
}

const route = useRoute()
const router = useRouter()
const keyword = ref('')
const results = ref<ArticleSearchItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const searched = ref(false)
const error = ref('')
const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 20,
  has_more: false,
})

onMounted(async () => {
  const query = typeof route.query.q === 'string' ? route.query.q : ''
  keyword.value = query
  if (query) {
    await runSearch()
  }
})

async function submitSearch() {
  await router.replace({ path: '/search', query: keyword.value ? { q: keyword.value } : {} })
  await runSearch()
}

async function runSearch(cursor = '') {
  searched.value = true
  if (!keyword.value) {
    results.value = []
    return
  }
  loading.value = !cursor
  error.value = ''
  try {
    const result = await searchArticles({ keyword: keyword.value, cursor: cursor || undefined, limit: page.limit })
    results.value = cursor ? [...results.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '搜索失败'
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
    await runSearch(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}

function highlight(value: string): HighlightPart[] {
  if (!keyword.value) {
    return [{ text: value, hit: false }]
  }
  const lowerValue = value.toLowerCase()
  const lowerKeyword = keyword.value.toLowerCase()
  const parts: HighlightPart[] = []
  let cursor = 0
  let index = lowerValue.indexOf(lowerKeyword)
  while (index >= 0) {
    if (index > cursor) {
      parts.push({ text: value.slice(cursor, index), hit: false })
    }
    parts.push({ text: value.slice(index, index + keyword.value.length), hit: true })
    cursor = index + keyword.value.length
    index = lowerValue.indexOf(lowerKeyword, cursor)
  }
  if (cursor < value.length) {
    parts.push({ text: value.slice(cursor), hit: false })
  }
  return parts.length ? parts : [{ text: value, hit: false }]
}
</script>
