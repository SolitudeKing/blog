<template>
  <div class="home-view">
    <section class="home-stage" aria-labelledby="home-title">
      <article class="home-stage__profile mist-luminous">
        <div class="home-stage__avatar">
          <img :src="authorAvatarSrc" :alt="`${author} 的头像`" @error="useDefaultAuthorAvatar" />
          <span v-if="usingDefaultAuthorAvatar" aria-hidden="true">{{ authorInitial }}</span>
        </div>

        <div class="home-stage__identity">
          <span class="home-kicker">Blog keeper · Solitude</span>
          <div id="home-title" role="heading" aria-level="1">
            <span>你好，我是</span>
            {{ author }}
            <small>@{{ authorHandle }}</small>
          </div>
          <p class="home-stage__motto">“{{ essay }}”</p>
          <p class="home-stage__status">
            <span aria-hidden="true" />
            {{ activeNotice?.title || '持续记录技术、设计与生活' }}
          </p>

          <nav class="home-socials" aria-label="作者链接">
            <a
              v-for="entry in socialEntries"
              :key="entry.key"
              :href="entry.href"
              :target="entry.external ? '_blank' : undefined"
              :rel="entry.external ? 'noopener noreferrer' : undefined"
            >
              <SvgIcon name="arrow-up-right" />
              {{ entry.label }}
            </a>
            <a href="/rss.xml">
              <SvgIcon name="rss" />
              RSS
            </a>
          </nav>
        </div>
      </article>

      <aside class="home-stage__intro" aria-labelledby="home-intro-title">
        <span class="home-kicker">{{ siteName }}</span>
        <div id="home-intro-title" role="heading" aria-level="2">一份持续更新的博客，也是公开的思考现场</div>
        <p>
          在这里记录工程实践、设计系统与构建过程。文章保留可复用的方法，也保留问题发生时的真实判断。
        </p>

        <dl class="home-metrics">
          <div>
            <dt>文章</dt>
            <dd :aria-label="articleTotal === null ? '文章总数暂不可用' : `${articleTotal} 篇文章`">
              {{ articleTotal ?? '—' }}
            </dd>
          </div>
          <div>
            <dt>专题</dt>
            <dd>{{ topicCatalog.length }}</dd>
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
          <div id="latest-title" role="heading" aria-level="2">最近发布的博客</div>
        </div>
        <RouterLink class="home-arrow-link" to="/archives">
          查看全部归档
          <SvgIcon name="arrow-right" />
        </RouterLink>
      </header>

      <div class="home-post-state" :aria-busy="loading || loadingMore">
        <template v-if="featuredArticle">
          <div class="home-latest-flow">
            <ArticleCard :article="featuredArticle" :index="1" variant="featured" />
            <div v-if="railArticles.length" class="home-story-rail" aria-label="更多近期博客">
              <ArticleCard
                v-for="(article, index) in railArticles"
                :key="article.id"
                :article="article"
                :index="index + 2"
                variant="rail"
              />
            </div>
          </div>

          <div v-if="remainingArticles.length" class="home-stream" aria-label="更多文章">
            <ArticleCard
              v-for="(article, index) in remainingArticles"
              :key="article.id"
              :article="article"
              :index="index + 5"
              variant="stream"
            />
          </div>
        </template>
        <div v-else-if="loading" class="home-latest-skeleton">
          <BaseSkeleton variant="card" height="480px" :count="1" gap="0" label="正在加载主文章" />
          <BaseSkeleton variant="rect" height="120px" :count="3" gap="32px" label="正在加载近期文章" />
        </div>
        <div v-else-if="error && !articles.length" class="page-state page-state--error" role="alert">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" @click="reload">重试</BaseButton>
        </div>
        <BaseEmpty
          v-else-if="!articles.length"
          title="暂时还没有发布文章"
          :description="currentThemeElements.home_latest_empty_description"
        >
          <template #icon>
            <SvgIcon name="document-lines" />
          </template>
        </BaseEmpty>

        <div v-if="articles.length && error" class="page-state page-state--compact page-state--error" role="alert">
          <p>{{ error }}</p>
          <BaseButton variant="secondary" size="sm" @click="reload">重试加载</BaseButton>
        </div>

        <div v-if="articles.length && !error" class="home-section__footer">
          <BasePagination
            mode="prevNext"
            :page="page.page"
            :has-more="page.has_more"
            :loading="loadingMore"
            @prev="goPrevPage"
            @next="goNextPage"
          />
          <span v-if="!page.has_more" class="home-section__end">
            {{ currentThemeElements.home_latest_end_text }}
          </span>
        </div>
      </div>
    </section>

    <section v-if="topicLinks.length" class="home-section home-section--tight" aria-labelledby="topics-title">
      <div class="home-topics">
        <div class="home-topics__intro">
          <span class="home-kicker">Topics</span>
          <div id="topics-title" role="heading" aria-level="2">从这些专题进入</div>
        </div>
        <nav class="home-topics__links" aria-label="文章专题">
          <RouterLink
            v-for="topic in topicLinks"
            :key="topic.key"
            :to="`/topics/${topic.slug}`"
          >
            <strong>{{ topic.name }}</strong>
            <small>{{ topic.label }}</small>
            <span>{{ topic.description }}</span>
          </RouterLink>
        </nav>
      </div>
    </section>

    <section v-if="activeNotice" class="home-section" aria-labelledby="notice-title">
      <div class="home-notice mist-glass--subtle">
        <div>
          <span class="home-kicker">Site notice</span>
          <div id="notice-title" role="heading" aria-level="2">{{ activeNotice.title }}</div>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import defaultAuthorAvatar from '@/assets/images/default-author-avatar.svg'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BasePagination from '@/components/base/BasePagination.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { getArticleList } from '@/api/modules/article'
import { getActiveNotice } from '@/api/modules/notice'
import { getTagList, getTopicList } from '@/api/modules/taxonomy'
import { resolveLobbyThemeElements } from '@/config/themeAppearance'
import { normalizeTopicLabel, topicCatalog } from '@/config/topicCatalog'
import { useSettingStore } from '@/stores/setting'
import { createSocialLinkEntries } from '@/utils/socialLinks'
import type { ArticleListItem } from '@/types/article'
import type { NoticeItem } from '@/types/notice'
import type { TagItem, TopicItem } from '@/types/taxonomy'

interface TopicLink {
  key: string
  label: string
  name: string
  slug: string
  description: string
}

const setting = useSettingStore()
const articles = ref<ArticleListItem[]>([])
const activeNotice = ref<NoticeItem | null>(null)
const topics = ref<TopicItem[]>([])
const tags = ref<TagItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const page = reactive({
  page: 1,
  page_size: 20,
  count: 0,
  has_more: false,
})

const siteName = computed(() => setting.lobby?.site_name ?? 'Solitude Blog')
const author = computed(() => setting.lobby?.author ?? 'Solitude King')
const essay = computed(() => setting.lobby?.essay?.trim() || '把复杂的技术写清楚，也把无法量化的感受留在字里行间。')
const currentThemeElements = computed(() => resolveLobbyThemeElements(setting.lobby))
const authorInitial = computed(() => author.value.trim().slice(0, 1).toUpperCase())
const configuredAuthorAvatarURL = computed(() => setting.lobby?.author_avatar_url?.trim() || '')
const authorAvatarFailed = ref(false)
const usingDefaultAuthorAvatar = computed(() => authorAvatarFailed.value || !configuredAuthorAvatarURL.value)
const authorAvatarSrc = computed(() =>
  usingDefaultAuthorAvatar.value ? defaultAuthorAvatar : configuredAuthorAvatarURL.value,
)
const authorHandle = computed(() =>
  author.value
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '.')
    .replace(/[^a-z0-9.\u4e00-\u9fa5]/g, ''),
)
const featuredArticle = computed(() => articles.value[0] ?? null)
const railArticles = computed(() => articles.value.slice(1, 4))
const remainingArticles = computed(() => articles.value.slice(4))

const socialEntries = computed(() => createSocialLinkEntries(setting.lobby?.social_links))

watch(configuredAuthorAvatarURL, () => {
  authorAvatarFailed.value = false
})

const articleTotal = computed(() => {
  const counts = topicCatalog.map((catalogTopic) => {
    const topic = findCatalogTopic(catalogTopic)
    return topic?.article_count
  })

  // 只有三个固定专题都返回计数时才展示总数，避免把当前分页长度误当作全站总量。
  if (counts.some((count) => typeof count !== 'number')) {
    return null
  }
  return counts.reduce<number>((total, count) => total + (count ?? 0), 0)
})

const topicLinks = computed<TopicLink[]>(() => {
  return topicCatalog.map((catalogTopic) => {
    const topic = findCatalogTopic(catalogTopic)
    return {
      key: catalogTopic.label,
      label: catalogTopic.label,
      name: catalogTopic.name,
      slug: catalogTopic.slug,
      description: topic?.description?.trim() || catalogTopic.description,
    }
  })
})

onMounted(async () => {
  await Promise.all([loadNotice(), loadTaxonomy(), loadArticles(1)])
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
    const [topicItems, tagItems] = await Promise.all([getTopicList(), getTagList()])
    topics.value = topicItems
    tags.value = tagItems
  } catch {
    topics.value = []
    tags.value = []
  }
}

async function reload() {
  articles.value = []
  page.page = 1
  page.has_more = false
  page.count = 0
  await loadArticles(1)
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
    articles.value = result.data
    page.page = result.page
    page.page_size = result.page_size
    page.count = result.count
    page.has_more = result.has_more
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    if (targetPage === 1) {
      loading.value = false
    } else {
      loadingMore.value = false
    }
  }
}

async function goNextPage() {
  if (loadingMore.value || !page.has_more) return
  await loadArticles(page.page + 1)
}

async function goPrevPage() {
  if (loadingMore.value || page.page <= 1) return
  await loadArticles(page.page - 1)
}

function findCatalogTopic(catalogTopic: { slug: string; label: string }) {
  // slug 是公开链接的稳定契约；label 仅用于兼容尚未完成迁移的旧接口数据。
  return (
    topics.value.find((item) => item.slug.trim().toLowerCase() === catalogTopic.slug) ??
    topics.value.find((item) => normalizeTopicLabel(item.label) === catalogTopic.label)
  )
}

function useDefaultAuthorAvatar() {
  authorAvatarFailed.value = true
}
</script>
