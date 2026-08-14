<template>
  <section class="archives-page" aria-labelledby="archive-title" :aria-busy="loading || loadingMore">
    <header class="archive-hero">
      <div class="archive-hero__copy">
        <p class="archive-kicker">Archive · {{ archiveRange }}</p>
        <div id="archive-title" role="heading" aria-level="1">所有足迹，都有刻度</div>
        <p>
          从最近一次发布向过去回望。这里按年份与月份整理已经公开的文章，让每一段记录都能被重新抵达。
        </p>
      </div>

      <div class="archive-hero__datum mist-luminous" role="group" :aria-label="archiveSummary">
        <strong class="archive-hero__number">{{ articles.length }}</strong>
        <span class="archive-hero__label">
          {{ page.has_more ? '篇已加载文章' : '篇公开文章' }} · {{ yearEntries.length }} 个年份
        </span>
        <SvgIcon class="archive-hero__wave" name="archive-wave" />
      </div>
    </header>

    <nav v-if="yearEntries.length" class="year-rail" aria-label="按年份跳转">
      <div class="year-rail__inner">
        <span class="archive-kicker">选择年份</span>
        <a
          v-for="year in yearEntries"
          :key="year.year"
          :href="`#year-${year.year}`"
          :aria-current="activeYear === year.year ? 'location' : undefined"
          @click="focusYear($event, year.year)"
        >
          {{ year.year }}
        </a>
      </div>
    </nav>

    <div class="archive-stream">
      <div
        v-if="loading && !articles.length"
        class="page-state"
        aria-busy="true"
        aria-label="正在加载归档"
      >
        <BaseSkeleton variant="card" :count="2" />
      </div>

      <div v-else-if="error && !articles.length" class="page-state page-state--error" role="alert">
        <p>{{ error }}</p>
        <BaseButton variant="secondary" @click="reload">重试</BaseButton>
      </div>

      <BaseEmpty
        v-else-if="!articles.length && !loading && !error"
        title="还没有归档内容"
        description="发布文章后会按年/月自动汇总到这里。"
      >
        <template #icon>
          <SvgIcon name="book-open" />
        </template>
      </BaseEmpty>

      <div v-if="error && articles.length" class="archive-load-error" role="alert">
        <span>{{ error }}</span>
        <BaseButton variant="secondary" size="sm" @click="retryCurrentPage">重试</BaseButton>
      </div>

      <section
        v-for="year in yearEntries"
        :id="`year-${year.year}`"
        :key="year.year"
        class="archive-year"
        :aria-labelledby="`archive-year-label-${year.year}`"
        tabindex="-1"
      >
        <div class="archive-year__label">
          <div :id="`archive-year-label-${year.year}`" role="heading" aria-level="2">
            {{ year.year }}
            <span class="sr-only">，{{ year.total }} 篇文章</span>
          </div>
        </div>

        <div class="archive-year__months">
          <section
            v-for="month in year.months"
            :key="month.key"
            class="archive-month"
            :aria-labelledby="`archive-month-${month.key}`"
          >
            <div :id="`archive-month-${month.key}`" role="heading" aria-level="3">{{ month.label }}</div>
            <ol class="archive-list">
              <li v-for="entry in month.entries" :key="entry.id" class="archive-item">
                <time :datetime="entry.datetime">{{ entry.displayDate }}</time>
                <RouterLink :to="`/articles/${entry.slug}`">{{ entry.title }}</RouterLink>
                <span v-if="entry.topic" class="archive-tag">{{ entry.topic }}</span>
              </li>
            </ol>
          </section>
        </div>
      </section>

      <div v-if="!error" class="archive-footer">
        <BasePagination
          mode="prevNext"
          :page="page.page"
          :has-more="page.has_more"
          :loading="loadingMore"
          @prev="goPrevPage"
          @next="goNextPage"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BasePagination from '@/components/base/BasePagination.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { getArticleList } from '@/api/modules/article'
import type { ArticleListItem } from '@/types/article'

interface ArchiveEntry {
  id: number
  slug: string
  title: string
  topic: string
  year: number
  month: number
  day: string
  datetime: string
  displayDate: string
}

interface ArchiveMonth {
  key: string
  label: string
  entries: ArchiveEntry[]
}

interface ArchiveYear {
  year: number
  total: number
  months: ArchiveMonth[]
}

const chineseMonths = ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月']
const englishMonthFormatter = new Intl.DateTimeFormat('en-US', { month: 'short' })
const articles = ref<ArchiveEntry[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const activeYear = ref<number | null>(null)
const page = reactive({
  page: 1,
  page_size: 50,
  count: 0,
  has_more: false,
})

let yearObserver: IntersectionObserver | null = null
let initialHashHandled = false

function toEntry(item: ArticleListItem): ArchiveEntry | null {
  const raw = item.published_at || item.created_at
  if (!raw) {
    return null
  }
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return null
  }
  const month = date.getMonth() + 1
  const day = String(date.getDate()).padStart(2, '0')
  return {
    id: item.id,
    slug: item.slug,
    title: item.title,
    topic: item.topic?.label || item.topic?.name || '',
    year: date.getFullYear(),
    month,
    day,
    datetime: raw,
    displayDate: `${String(month).padStart(2, '0')}.${day}`,
  }
}

const yearEntries = computed<ArchiveYear[]>(() => {
  const grouped = new Map<number, Map<number, ArchiveEntry[]>>()
  for (const entry of articles.value) {
    if (!grouped.has(entry.year)) {
      grouped.set(entry.year, new Map())
    }
    const months = grouped.get(entry.year)!
    if (!months.has(entry.month)) {
      months.set(entry.month, [])
    }
    months.get(entry.month)!.push(entry)
  }
  const years: ArchiveYear[] = []
  for (const [year, months] of grouped) {
    const monthList: ArchiveMonth[] = []
    let total = 0
    for (const [month, entries] of months) {
      total += entries.length
      const monthName = englishMonthFormatter.format(new Date(year, month - 1)).toUpperCase()
      monthList.push({
        key: `${year}-${month}`,
        label: `${monthName} / ${chineseMonths[month - 1]}`,
        entries: entries.sort((a, b) => Number(b.day) - Number(a.day)),
      })
    }
    monthList.sort((a, b) => Number(b.key.split('-')[1]) - Number(a.key.split('-')[1]))
    years.push({ year, total, months: monthList })
  }
  return years.sort((a, b) => b.year - a.year)
})

const archiveRange = computed(() => {
  if (!yearEntries.value.length) {
    return 'Now'
  }
  const newest = yearEntries.value[0].year
  const oldest = yearEntries.value[yearEntries.value.length - 1].year
  return newest === oldest ? String(newest) : `${oldest}—${newest}`
})

const archiveSummary = computed(() => {
  const countLabel = page.has_more ? `已加载 ${articles.value.length} 篇文章` : `共 ${articles.value.length} 篇公开文章`
  return `${countLabel}，分布在 ${yearEntries.value.length} 个年份`
})

onMounted(async () => {
  await loadArticles(1)
})

onBeforeUnmount(() => {
  yearObserver?.disconnect()
})

async function setupYearObserver() {
  await nextTick()
  yearObserver?.disconnect()
  const sections = yearEntries.value
    .map(({ year }) => document.getElementById(`year-${year}`))
    .filter((section): section is HTMLElement => section !== null)
  if (!sections.length) {
    activeYear.value = null
    return
  }
  if (!activeYear.value || !yearEntries.value.some(({ year }) => year === activeYear.value)) {
    activeYear.value = yearEntries.value[0].year
  }
  if (!initialHashHandled) {
    const match = window.location.hash.match(/^#year-(\d{4})$/)
    if (!match) {
      initialHashHandled = true
    } else {
      const hashYear = Number(match[1])
      const hashTarget = document.getElementById(`year-${hashYear}`)
      if (hashTarget) {
        initialHashHandled = true
        activeYear.value = hashYear
        window.requestAnimationFrame(() => {
          hashTarget.scrollIntoView({ behavior: 'auto', block: 'start' })
          hashTarget.focus({ preventScroll: true })
        })
      }
    }
  }
  if (!('IntersectionObserver' in window)) {
    return
  }
  yearObserver = new IntersectionObserver(
    (entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0]
      if (!visible) {
        return
      }
      const year = Number(visible.target.id.replace('year-', ''))
      if (Number.isFinite(year)) {
        activeYear.value = year
      }
    },
    { rootMargin: '-25% 0px -65% 0px', threshold: 0 },
  )
  sections.forEach((section) => yearObserver?.observe(section))
}

function focusYear(event: MouseEvent, year: number) {
  event.preventDefault()
  const target = document.getElementById(`year-${year}`)
  if (!target) {
    return
  }
  activeYear.value = year
  const url = new URL(window.location.href)
  url.hash = `year-${year}`
  window.history.replaceState(window.history.state, '', url)
  const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
  target.scrollIntoView({ behavior: reduceMotion ? 'auto' : 'smooth', block: 'start' })
  window.setTimeout(() => target.focus({ preventScroll: true }), reduceMotion ? 0 : 360)
}

async function loadArticles(targetPage: number) {
  if (targetPage === 1) {
    loading.value = true
  } else {
    loadingMore.value = true
  }
  error.value = ''
  try {
    const result = await getArticleList({
      page: targetPage,
      page_size: page.page_size,
    })
    const entries = result.data
      .map(toEntry)
      .filter((entry): entry is ArchiveEntry => entry !== null)
    articles.value = entries
    page.page = result.page
    page.page_size = result.page_size
    page.count = result.count
    page.has_more = result.has_more
    await setupYearObserver()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载归档失败'
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function reload() {
  articles.value = []
  page.page = 1
  page.has_more = false
  page.count = 0
  await loadArticles(1)
}

async function retryCurrentPage() {
  await loadArticles(page.page)
}

async function goNextPage() {
  if (loadingMore.value || !page.has_more) return
  await loadArticles(page.page + 1)
}

async function goPrevPage() {
  if (loadingMore.value || page.page <= 1) return
  await loadArticles(page.page - 1)
}
</script>
