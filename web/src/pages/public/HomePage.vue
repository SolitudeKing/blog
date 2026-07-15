<template>
  <div class="home-view">
    <section class="home-stage" aria-labelledby="home-title">
      <article class="home-stage__profile mist-luminous">
        <div class="home-stage__avatar" role="img" :aria-label="`${author} 的头像`">
          <span aria-hidden="true">{{ authorInitial }}</span>
          <svg aria-hidden="true" viewBox="0 0 220 220">
            <path d="M18 160c38-22 57 22 95 0s57 22 95 0" />
            <path d="M26 180c34-18 51 18 85 0s51 18 85 0" />
          </svg>
        </div>

        <div class="home-stage__identity">
          <span class="home-kicker">Blog keeper · Solitude</span>
          <h1 id="home-title">
            <span>你好，我是</span>
            {{ author }}
            <small>@{{ authorHandle }}</small>
          </h1>
          <p class="home-stage__motto">“{{ essay }}”</p>
          <p class="home-stage__status">
            <span aria-hidden="true" />
            {{ activeNotice?.title || '持续记录技术、设计与生活' }}
          </p>

          <nav class="home-socials" aria-label="作者链接">
            <a
              v-for="entry in socialEntries"
              :key="entry.key"
              :href="entry.url"
              target="_blank"
              rel="noreferrer"
            >
              <svg aria-hidden="true" viewBox="0 0 24 24">
                <path d="M7 17 17 7M9 7h8v8" />
              </svg>
              {{ entry.label }}
            </a>
            <a href="/rss.xml">
              <svg aria-hidden="true" viewBox="0 0 24 24">
                <circle cx="6" cy="18" r="1" />
                <path d="M5 11a8 8 0 0 1 8 8M5 5a14 14 0 0 1 14 14" />
              </svg>
              RSS
            </a>
          </nav>
        </div>
      </article>

      <aside class="home-stage__intro" aria-labelledby="home-intro-title">
        <span class="home-kicker">{{ siteName }}</span>
        <h2 id="home-intro-title">一份持续更新的博客，也是公开的思考现场</h2>
        <p>
          在这里记录工程实践、设计系统与构建过程。文章保留可复用的方法，也保留问题发生时的真实判断。
        </p>

        <dl class="home-metrics">
          <div>
            <dt>文章</dt>
            <dd>{{ articles.length }}</dd>
          </div>
          <div>
            <dt>分类</dt>
            <dd>{{ categories.length }}</dd>
          </div>
          <div>
            <dt>标签</dt>
            <dd>{{ tags.length }}</dd>
          </div>
        </dl>

        <div class="home-stage__actions">
          <a class="mist-button" href="#latest-posts">查看最近发布</a>
          <RouterLink class="mist-button mist-button--secondary" to="/archives">
            浏览全部归档
          </RouterLink>
        </div>
      </aside>
    </section>

    <section id="latest-posts" class="home-section" aria-labelledby="latest-title">
      <header class="home-section__heading">
        <div>
          <span class="home-kicker">Latest posts</span>
          <h2 id="latest-title">最近发布的博客</h2>
        </div>
        <RouterLink class="home-arrow-link" to="/archives">
          查看全部归档
          <svg aria-hidden="true" viewBox="0 0 24 24">
            <path d="M5 12h14M13 6l6 6-6 6" />
          </svg>
        </RouterLink>
      </header>

      <div class="home-post-state" :aria-busy="loading || loadingMore">
        <div v-if="featuredArticle" class="home-catalog">
          <ArticleCard :article="featuredArticle" :index="1" variant="featured" />
          <div v-if="sideArticles.length" class="home-catalog__side" aria-label="更多近期文章">
            <ArticleCard
              v-for="(article, index) in sideArticles"
              :key="article.id"
              :article="article"
              :index="index + 2"
              :offset="index === 1"
              variant="compact"
            />
          </div>
        </div>

        <div v-if="remainingArticles.length" class="home-stream" aria-label="更多文章">
          <ArticleCard
            v-for="(article, index) in remainingArticles"
            :key="article.id"
            :article="article"
            :index="index + 4"
            variant="stream"
          />
        </div>

        <div v-else-if="loading && !articles.length" class="page-state">
          <BaseSkeleton variant="card" :count="3" />
        </div>
        <div v-else-if="error && !articles.length" class="page-state page-state--error" role="alert">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" @click="reload">重试</BaseButton>
        </div>
        <BaseEmpty
          v-else-if="!articles.length"
          title="暂时还没有发布文章"
          description="第一篇文章正在潮汐之外酝酿。"
        >
          <template #icon>
            <svg viewBox="0 0 24 24">
              <path d="M5 4h10l4 4v12H5V4Zm10 0v5h4M8 13h8M8 16h6" />
            </svg>
          </template>
        </BaseEmpty>

        <div v-if="articles.length && error" class="page-state page-state--compact page-state--error" role="alert">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" size="sm" @click="loadMore">重试加载</BaseButton>
        </div>

        <div v-if="articles.length && !error" class="home-section__footer">
          <BaseButton v-if="page.has_more" variant="secondary" :loading="loadingMore" @click="loadMore">
            加载更多
          </BaseButton>
          <span v-else class="home-section__end">已经读到潮汐尽头</span>
        </div>
      </div>
    </section>

    <section v-if="topicLinks.length" class="home-section home-section--tight" aria-labelledby="topics-title">
      <div class="home-topics">
        <div class="home-topics__intro">
          <span class="home-kicker">Topics</span>
          <h2 id="topics-title">从这些专题进入</h2>
        </div>
        <nav class="home-topics__links" aria-label="文章专题">
          <RouterLink
            v-for="topic in topicLinks"
            :key="topic.key"
            :to="{ path: '/search', query: { q: topic.name } }"
          >
            <strong>{{ topic.name }}</strong>
            <span>{{ topic.description }}</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <section v-if="activeNotice" class="home-section" aria-labelledby="notice-title">
      <div class="home-notice mist-glass--subtle">
        <div>
          <span class="home-kicker">Site notice</span>
          <h2 id="notice-title">{{ activeNotice.title }}</h2>
          <p>{{ activeNotice.content }}</p>
        </div>
        <RouterLink class="mist-button mist-button--secondary" to="/archives">
          继续阅读
        </RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import { getArticleList } from '@/api/modules/article'
import { getActiveNotice } from '@/api/modules/notice'
import { getCategoryList, getTagList } from '@/api/modules/taxonomy'
import { useSettingStore } from '@/stores/setting'
import type { CursorPage } from '@/api/types'
import type { ArticleListItem } from '@/types/article'
import type { NoticeItem } from '@/types/notice'
import type { CategoryItem, TagItem } from '@/types/taxonomy'

interface TopicLink {
  key: string
  name: string
  description: string
}

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

const siteName = computed(() => setting.lobby?.site_name ?? 'Solitude Blog')
const author = computed(() => setting.lobby?.author ?? 'Solitude King')
const essay = computed(() => setting.lobby?.essay?.trim() || '把复杂的技术写清楚，也把无法量化的感受留在字里行间。')
const authorInitial = computed(() => author.value.trim().slice(0, 1).toUpperCase())
const authorHandle = computed(() =>
  author.value
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '.')
    .replace(/[^a-z0-9.\u4e00-\u9fa5]/g, ''),
)
const featuredArticle = computed(() => articles.value[0] ?? null)
const sideArticles = computed(() => articles.value.slice(1, 3))
const remainingArticles = computed(() => articles.value.slice(3))

const socialEntries = computed(() => {
  const labelMap: Record<string, string> = {
    github: 'GitHub',
    gitee: 'Gitee',
    bilibili: 'Bilibili',
    douyin: 'Douyin',
  }
  return Object.entries(setting.lobby?.social_links ?? {})
    .filter(([, url]) => Boolean(url))
    .map(([key, url]) => ({ key, url, label: labelMap[key] ?? key }))
})

const topicLinks = computed<TopicLink[]>(() => {
  const result: TopicLink[] = categories.value.slice(0, 3).map((category) => ({
    key: `category-${category.id}`,
    name: category.name,
    description: category.description || `浏览 ${category.name} 相关内容`,
  }))
  if (result.length >= 3) {
    return result
  }
  for (const tag of tags.value) {
    if (result.length >= 3 || result.some((item) => item.name === tag.name)) {
      continue
    }
    result.push({
      key: `tag-${tag.id}`,
      name: tag.name,
      description: tag.description || `查看标记为 ${tag.name} 的文章`,
    })
  }
  return result
})

onMounted(async () => {
  await Promise.all([loadNotice(), loadTaxonomy(), loadArticles()])
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
  if (!cursor) {
    loading.value = true
  }
  error.value = ''
  try {
    const result = await getArticleList({ cursor: cursor || undefined, limit: page.limit })
    articles.value = cursor ? [...articles.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    if (!cursor) {
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
    await loadArticles(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}
</script>
