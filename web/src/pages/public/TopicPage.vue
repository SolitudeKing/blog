<template>
  <NotFoundPage v-if="!activeTopic" />

  <section
    v-else
    class="topic-page"
    :aria-labelledby="topicTitleId"
    :aria-busy="loading || loadingMore"
  >
    <header class="topic-hero">
      <div class="topic-hero__copy">
        <RouterLink class="topic-hero__back" to="/">
          <SvgIcon name="arrow-left" />
          返回首页
        </RouterLink>
        <p class="topic-kicker">Topic · {{ activeTopic.label }}</p>
        <div :id="topicTitleId" role="heading" aria-level="1">{{ activeTopic.name }}</div>
        <p>{{ activeTopic.description }}</p>
      </div>

      <aside class="topic-hero__marker mist-luminous" :aria-label="topicSummary">
        <span>{{ activeTopic.label }}</span>
        <strong>{{ loading && !articles.length ? '—' : loadedCount }}</strong>
        <small>
          {{ loading && !articles.length ? '正在整理文章' : page.has_more ? '篇已加载文章' : '篇专题文章' }}
        </small>
      </aside>
    </header>

    <section class="topic-stream" aria-labelledby="topic-stream-title">
      <header class="topic-stream__heading">
        <div>
          <p class="topic-kicker">Reading stream</p>
          <div id="topic-stream-title" role="heading" aria-level="2">沿着这条线索继续阅读</div>
        </div>
        <span role="status" aria-live="polite">{{ resultStatus }}</span>
      </header>

      <div
        v-if="loading && !articles.length"
        class="topic-state"
        aria-busy="true"
        aria-label="正在加载专题文章"
      >
        <BaseSkeleton variant="card" :count="3" />
      </div>

      <div v-else-if="error && !articles.length" class="topic-state topic-state--error" role="alert">
        <p>{{ error }}</p>
        <BaseButton variant="secondary" @click="reload">重新加载</BaseButton>
      </div>

      <div v-else-if="articles.length" class="topic-article-list">
        <ArticleCard
          v-for="(article, index) in articles"
          :key="article.id"
          :article="article"
          :index="index + 1"
          variant="stream"
        />
      </div>

      <BaseEmpty
        v-else
        :title="`${activeTopic.name}暂时还没有文章`"
        description="内容正在雾中酝酿，可以先去归档看看其他记录。"
        cta-text="浏览全部归档"
        cta-to="/archives"
      >
        <template #icon>
          <SvgIcon name="document-lines" />
        </template>
      </BaseEmpty>

      <div v-if="error && articles.length" class="topic-load-error" role="alert">
        <span>{{ error }}</span>
        <BaseButton variant="secondary" size="sm" @click="reload">重试加载</BaseButton>
      </div>

      <div v-if="articles.length && !error" class="topic-stream__footer">
        <BasePagination
          mode="prevNext"
          :page="page.page"
          :has-more="page.has_more"
          :loading="loadingMore"
          @prev="goPrevPage"
          @next="goNextPage"
        />
        <span v-if="!page.has_more">已抵达 {{ activeTopic.name }} 的尽头</span>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BasePagination from '@/components/base/BasePagination.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSkeleton from '@/components/base/BaseSkeleton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import NotFoundPage from '@/pages/public/NotFoundPage.vue'
import { getArticleList } from '@/api/modules/article'
import { findTopicBySlug } from '@/config/topicCatalog'
import type { ArticleListItem } from '@/types/article'

const route = useRoute()
const articles = ref<ArticleListItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
let articleRequestId = 0

const page = reactive({
  page: 1,
  page_size: 20,
  count: 0,
  has_more: false,
})

const activeTopic = computed(() => findTopicBySlug(route.params.slug))
const topicTitleId = computed(() => `topic-${activeTopic.value?.slug ?? 'unknown'}-title`)
const loadedCount = computed(() => articles.value.length)
const topicSummary = computed(() => {
  if (!activeTopic.value) {
    return ''
  }
  if (loading.value && !articles.value.length) {
    return `${activeTopic.value.name}，正在整理文章`
  }
  const countLabel = page.has_more ? `已加载 ${loadedCount.value} 篇` : `共 ${loadedCount.value} 篇`
  return `${activeTopic.value.name}，${countLabel}`
})
const resultStatus = computed(() => {
  if (loading.value || loadingMore.value) {
    return '正在整理文章'
  }
  if (error.value && !articles.value.length) {
    return '加载失败'
  }
  return page.has_more ? `已加载 ${loadedCount.value} 篇` : `共 ${loadedCount.value} 篇`
})

function resetList() {
  articles.value = []
  loading.value = false
  loadingMore.value = false
  error.value = ''
  page.page = 1
  page.has_more = false
  page.count = 0
}

async function loadArticles(targetPage: number) {
  const topic = activeTopic.value
  if (!topic) {
    return
  }

  const requestId = ++articleRequestId
  if (targetPage === 1) {
    loading.value = true
  } else {
    loadingMore.value = true
  }
  error.value = ''

  try {
    // topic 参数使用稳定 slug，避免诗意名称变化后破坏专题筛选链接。
    const result = await getArticleList({
      topic: topic.slug,
      page: targetPage,
      page_size: page.page_size,
    })
    if (requestId !== articleRequestId) {
      return
    }
    articles.value = result.data
    page.page = result.page
    page.page_size = result.page_size
    page.count = result.count
    page.has_more = result.has_more
  } catch (err) {
    if (requestId === articleRequestId) {
      error.value = err instanceof Error ? err.message : '加载专题文章失败'
    }
  } finally {
    if (requestId === articleRequestId) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

async function reload() {
  articleRequestId += 1
  resetList()
  await loadArticles(1)
}

async function goNextPage() {
  if (loadingMore.value || !page.has_more) return
  await loadArticles(page.page + 1)
}

async function goPrevPage() {
  if (loadingMore.value || page.page <= 1) return
  await loadArticles(page.page - 1)
}

watch(
  () => route.params.slug,
  async () => {
    articleRequestId += 1
    resetList()
    if (activeTopic.value) {
      await loadArticles(1)
    }
  },
  { immediate: true },
)
</script>
