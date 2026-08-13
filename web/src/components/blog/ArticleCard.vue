<template>
  <article
    v-if="variant === 'featured'"
    :id="domId"
    class="home-featured-story"
    :aria-labelledby="headingId"
  >
    <div class="home-featured-story__media" aria-hidden="true">
      <span class="home-featured-story__index">{{ topicLabel }} / Top</span>
      <img :class="{ 'is-default': usingDefaultCover }" :src="coverImageSrc" alt="" @error="useDefaultCover" />
    </div>

    <div class="home-featured-story__body">
      <div role="heading" aria-level="3">
        <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
      </div>
      <p v-if="article.summary">{{ article.summary }}</p>
      <div class="home-meta-row">
        <time :datetime="articleDate">{{ displayDate }}</time>
        <span>{{ article.view_count }} 次浏览</span>
      </div>
    </div>
  </article>

  <article
    v-else-if="variant === 'rail'"
    :id="domId"
    class="home-story-rail__item"
    :aria-labelledby="headingId"
  >
    <span class="home-story-rail__number">
      {{ articleNumber }} / {{ topicLabel }}<template v-if="primaryTag"> · {{ primaryTag }}</template>
    </span>
    <div role="heading" aria-level="3">
      <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
    </div>
    <p v-if="article.summary">{{ article.summary }}</p>
    <div class="home-meta-row">
      <time :datetime="articleDate">{{ displayDate }}</time>
      <span>{{ article.view_count }} 次浏览</span>
    </div>
  </article>

  <article
    v-else
    :id="domId"
    class="home-post-card home-post-card--stream mist-card mist-card--hoverable"
    :aria-labelledby="headingId"
  >
    <div class="home-post-card__top">
      <span class="home-post-card__tag">{{ topicLabel }}</span>
      <time :datetime="articleDate">{{ displayDate }}</time>
    </div>

    <div class="home-post-card__body">
      <span class="home-post-card__index">POST / {{ articleNumber }}</span>
      <div role="heading" aria-level="3">
        <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
      </div>
      <p v-if="article.summary">{{ article.summary }}</p>
    </div>

    <footer class="home-post-card__footer">
      <span>{{ article.view_count }} 次浏览</span>
      <span>{{ article.tags.slice(0, 2).join(' · ') || '随笔' }}</span>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import defaultArticlePreview from '@/assets/images/default-article-preview.svg'
import type { ArticleListItem } from '@/types/article'

const props = withDefaults(
  defineProps<{
    article: ArticleListItem
    domId?: string
    index?: number
    variant?: 'featured' | 'rail' | 'stream'
  }>(),
  {
    domId: undefined,
    index: 1,
    variant: 'rail',
  },
)

const headingId = computed(() => `${props.domId || `article-${props.article.id}`}-title`)
const articlePath = computed(() => `/articles/${props.article.slug}`)
const articleDate = computed(() => props.article.published_at || props.article.created_at)
const articleNumber = computed(() => String(props.article.id || props.index).padStart(3, '0'))
const topicLabel = computed(() => props.article.topic?.label || props.article.topic?.name || 'NODES')
const primaryTag = computed(() => props.article.tags[0] || '')
const coverFailed = ref(false)
const coverURL = computed(() => props.article.cover_url?.trim() || '')
const usingDefaultCover = computed(() => coverFailed.value || !coverURL.value)
const coverImageSrc = computed(() => (usingDefaultCover.value ? defaultArticlePreview : coverURL.value))

watch(coverURL, () => {
  coverFailed.value = false
})

function useDefaultCover() {
  coverFailed.value = true
}

const displayDate = computed(() => {
  if (!articleDate.value) {
    return '未发布'
  }
  const date = new Date(articleDate.value)
  if (Number.isNaN(date.getTime())) {
    return '日期未知'
  }
  const parts = new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(date)
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))
  return `${values.year}.${values.month}.${values.day}`
})
</script>
