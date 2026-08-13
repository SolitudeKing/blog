<template>
  <section class="search-page" aria-labelledby="search-title">
    <header class="search-hero">
      <div class="search-hero__copy">
        <p class="search-kicker">Search the current</p>
        <div id="search-title" role="heading" aria-level="1">打捞一段想法</div>
        <p>输入一个词，沿着标题、摘要、正文、专题与标签寻找。也可以从常用航标开始，看看它会把你带去哪里。</p>
      </div>

      <aside class="search-hero__aside" :aria-label="searchOverview">
        <div class="search-stat">
          <strong>{{ searched ? results.length : '—' }}</strong>
          <span>{{ searched ? (page.has_more ? '条已加载' : '篇结果') : '等待搜索' }}</span>
        </div>
        <div class="search-stat">
          <strong>{{ searched ? matchedFieldCount : '—' }}</strong>
          <span>个命中字段</span>
        </div>
      </aside>
    </header>

    <section class="search-workbench mist-glass--strong" aria-labelledby="search-workbench-title">
      <div id="search-workbench-title" class="sr-only" role="heading" aria-level="2">文章搜索工作台</div>
      <form class="search-form" role="search" @submit.prevent="submitSearch">
        <label class="sr-only" for="article-search">搜索文章标题、摘要、正文、专题或标签</label>
        <div class="search-input-wrap">
          <SvgIcon class="search-input-wrap__icon" name="search" />
          <input
            id="article-search"
            ref="inputRef"
            v-model.trim="keyword"
            class="mist-input search-form__input"
            type="search"
            name="q"
            autocomplete="off"
            placeholder="例如：设计系统、写作、Vue……"
            aria-describedby="search-hint search-status"
            aria-controls="search-results"
          />
          <button
            v-if="keyword"
            class="search-clear"
            type="button"
            aria-label="清空搜索"
            @click="clearSearch"
          >
            <SvgIcon name="close" />
          </button>
        </div>
        <BaseButton class="search-submit" type="submit" :loading="loading">搜索</BaseButton>
        <p id="search-hint" class="sr-only">提交搜索后，查询会写入地址栏，搜索结果可以分享。</p>
        <p id="search-status" class="sr-only" aria-live="polite">{{ searchStatus }}</p>
      </form>

      <div class="search-suggestions" role="group" aria-label="常用搜索词">
        <span class="search-suggestions__label">试试这些航标</span>
        <button
          v-for="suggestion in suggestions"
          :key="suggestion"
          class="search-chip"
          type="button"
          :aria-pressed="normalizedKeyword === suggestion.toLocaleLowerCase('zh-CN')"
          @click="applySuggestion(suggestion)"
        >
          {{ suggestion }}
        </button>
      </div>
    </section>

    <div class="search-content" :class="{ 'search-content--single': !fieldMap.length }">
      <section
        id="search-results"
        class="search-results"
        aria-labelledby="results-title"
        :aria-busy="loading || loadingMore"
      >
        <header v-if="searched" class="search-results__header">
          <div id="results-title" class="search-results__count" role="heading" aria-level="2">
            {{ page.has_more ? '已加载' : '找到' }} <strong>{{ results.length }}</strong> 篇文章
          </div>
          <span>按相关度排序</span>
        </header>
        <div v-else id="results-title" class="sr-only" role="heading" aria-level="2">搜索结果</div>

        <div v-if="loading && !results.length" class="page-state" aria-busy="true" aria-label="正在搜索">
          <BaseSkeleton variant="card" :count="2" />
        </div>

        <div v-else-if="error && !results.length" class="page-state page-state--error" role="alert">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" @click="retrySearch">重试</BaseButton>
        </div>

        <ol v-else-if="searched && results.length" class="search-result-list">
          <li v-for="(item, index) in results" :key="item.id" class="search-result">
            <span class="search-result__index" aria-hidden="true">{{ formatResultIndex(index) }}</span>
            <article>
              <div role="heading" aria-level="3">
                <RouterLink :to="`/articles/${item.slug}`">{{ item.title }}</RouterLink>
              </div>
              <p>
                <template v-for="(part, partIndex) in highlight(item.snippet)" :key="`${item.id}-${partIndex}`">
                  <mark v-if="part.hit">{{ part.text }}</mark>
                  <span v-else>{{ part.text }}</span>
                </template>
              </p>
              <div class="search-result__meta">
                <time :datetime="item.published_at || item.created_at">{{ formatDate(item) }}</time>
                <span v-if="item.topic" class="search-result__tag">{{ item.topic.label || item.topic.name }}</span>
                <span v-for="field in item.matched_fields" :key="field">命中 {{ displayField(field) }}</span>
              </div>
            </article>
          </li>
        </ol>

        <BaseEmpty
          v-else-if="searched"
          title="这片水域还没有记录"
          description="可以尝试缩短关键词，或从“设计”“代码”“写作”这些航标重新出发。"
        >
          <template #icon>
            <SvgIcon name="search-minus" />
          </template>
        </BaseEmpty>

        <div v-else class="page-state search-page__prompt">
          <SvgIcon name="search" />
          <p>输入关键词开始搜索</p>
        </div>

        <div v-if="error && results.length" class="search-load-error" role="alert">
          <span>{{ error }}</span>
          <BaseButton variant="secondary" size="sm" @click="loadMore">重试加载</BaseButton>
        </div>

        <div v-if="results.length && page.has_more && !error" class="search-footer">
          <BaseButton variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
        </div>
      </section>

      <aside v-if="fieldMap.length" class="search-aside mist-glass" aria-labelledby="search-map-title">
        <div id="search-map-title" role="heading" aria-level="2">命中海图</div>
        <div class="search-map">
          <div v-for="field in fieldMap" :key="field.label" class="search-map__row">
            <span>{{ field.label }}</span>
            <span class="search-map__track" aria-hidden="true">
              <span class="search-map__fill" :style="{ '--map-width': `${field.width}%` }" />
            </span>
            <span>{{ field.count }}</span>
          </div>
        </div>
        <p>同一篇文章可能同时命中多个字段。</p>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { searchArticles } from '@/api/modules/article'
import type { CursorPage } from '@/api/types'
import type { ArticleSearchItem } from '@/types/article'

interface HighlightPart {
  text: string
  hit: boolean
}

interface SearchMapItem {
  label: string
  count: number
  width: number
}

const suggestions = ['设计', '代码', '写作', 'Vue', '前端']
const fieldLabels: Record<string, string> = {
  title: '标题',
  summary: '摘要',
  content: '正文',
  topic: '专题',
  topic_name: '专题名称',
  topic_label: '专题 Label',
  tag: '标签',
}
const dateFormatter = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
})
const route = useRoute()
const router = useRouter()
const inputRef = ref<HTMLInputElement | null>(null)
const keyword = ref('')
const results = ref<ArticleSearchItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const searched = ref(false)
const error = ref('')
let searchRequestId = 0
const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 20,
  has_more: false,
})

const normalizedKeyword = computed(() => keyword.value.trim().toLocaleLowerCase('zh-CN'))

const matchedFieldCount = computed(() => {
  const fields = new Set(results.value.flatMap((item) => item.matched_fields))
  return fields.size
})

const fieldMap = computed<SearchMapItem[]>(() => {
  const counts = new Map<string, number>()
  for (const result of results.value) {
    for (const field of result.matched_fields) {
      counts.set(field, (counts.get(field) ?? 0) + 1)
    }
  }
  const entries = [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 6)
  const max = entries[0]?.[1] ?? 1
  return entries.map(([field, count]) => ({
    label: displayField(field),
    count,
    width: Math.max(12, Math.round((count / max) * 100)),
  }))
})

const searchOverview = computed(() => {
  if (!searched.value) {
    return '尚未开始搜索'
  }
  const resultLabel = page.has_more ? `已加载 ${results.value.length} 条结果` : `找到 ${results.value.length} 篇文章`
  return `${resultLabel}，命中 ${matchedFieldCount.value} 个内容字段`
})

const searchStatus = computed(() => {
  if (loading.value || loadingMore.value) {
    return `正在搜索“${keyword.value}”`
  }
  if (!searched.value) {
    return ''
  }
  if (error.value && !results.value.length) {
    return '搜索失败，请重试'
  }
  return page.has_more
    ? `关键词“${keyword.value}”已加载 ${results.value.length} 条结果，还有更多结果`
    : `关键词“${keyword.value}”找到 ${results.value.length} 篇文章`
})

async function submitSearch() {
  const normalized = keyword.value.trim()
  const currentQuery = typeof route.query.q === 'string' ? route.query.q : ''
  keyword.value = normalized
  if (normalized === currentQuery) {
    await runSearch('', true)
    return
  }
  await router.replace({ path: '/search', query: normalized ? { q: normalized } : {} })
}

async function applySuggestion(suggestion: string) {
  keyword.value = suggestion
  await submitSearch()
  await nextTick()
  inputRef.value?.focus({ preventScroll: true })
}

async function clearSearch() {
  keyword.value = ''
  if (route.query.q) {
    await router.replace({ path: '/search' })
  } else {
    await runSearch('', true)
  }
  await nextTick()
  inputRef.value?.focus({ preventScroll: true })
}

async function runSearch(cursor = '', replace = false) {
  const requestId = ++searchRequestId
  if (!keyword.value) {
    results.value = []
    searched.value = false
    error.value = ''
    loading.value = false
    loadingMore.value = false
    page.cursor = ''
    page.next_cursor = ''
    page.has_more = false
    return
  }
  searched.value = true
  if (replace || !cursor) {
    results.value = []
    loadingMore.value = false
    page.cursor = ''
    page.next_cursor = ''
    page.has_more = false
  }
  loading.value = !cursor
  error.value = ''
  try {
    const result = await searchArticles({ keyword: keyword.value, cursor: cursor || undefined, limit: page.limit })
    if (requestId !== searchRequestId) {
      return
    }
    results.value = cursor ? [...results.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    if (requestId === searchRequestId) {
      error.value = err instanceof Error ? err.message : '搜索失败'
    }
  } finally {
    if (requestId === searchRequestId) {
      loading.value = false
    }
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

async function retrySearch() {
  await runSearch('', true)
}

function formatResultIndex(index: number) {
  return String(index + 1).padStart(2, '0')
}

function displayField(field: string) {
  return fieldLabels[field] ?? field
}

function formatDate(item: ArticleSearchItem) {
  const raw = item.published_at || item.created_at
  if (!raw) {
    return '未发布'
  }
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '日期未知'
  }
  return dateFormatter.format(date)
}

function highlight(value: string): HighlightPart[] {
  if (!keyword.value) {
    return [{ text: value, hit: false }]
  }
  const lowerValue = value.toLocaleLowerCase('zh-CN')
  const lowerKeyword = keyword.value.toLocaleLowerCase('zh-CN')
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

watch(
  () => route.query.q,
  async (rawQuery) => {
    const query = typeof rawQuery === 'string' ? rawQuery.trim() : ''
    keyword.value = query
    await runSearch('', true)
  },
  { immediate: true },
)
</script>
