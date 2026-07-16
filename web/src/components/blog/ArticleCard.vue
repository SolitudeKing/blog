<template>
  <article
    v-if="variant === 'featured'"
    :id="domId"
    class="home-featured-story"
    :aria-labelledby="headingId"
  >
    <div class="home-featured-story__media" aria-hidden="true">
      <span class="home-featured-story__index">{{ topicLabel }} / Top</span>
      <svg viewBox="0 0 680 220">
        <rect x="36" y="50" width="164" height="116" rx="16" stroke-width="1.4" />
        <rect x="258" y="28" width="164" height="70" rx="12" stroke-width="1.2" opacity=".72" />
        <rect x="258" y="124" width="164" height="70" rx="12" stroke-width="1.2" opacity=".72" />
        <rect x="480" y="76" width="164" height="70" rx="12" stroke-width="1.2" opacity=".48" />
        <path
          d="M200 108h58M422 63h28c18 0 30 12 30 30v18M422 159h28c18 0 30-12 30-30v-18"
          stroke-width="1.4"
        />
        <circle cx="230" cy="108" r="4" />
        <path d="M68 84h98M68 108h72M68 132h84" stroke-width="1.2" opacity=".62" />
      </svg>
    </div>

    <div class="home-featured-story__body">
      <h3>
        <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
      </h3>
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
    <h3>
      <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
    </h3>
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
      <h3>
        <RouterLink :id="headingId" :to="articlePath">{{ article.title }}</RouterLink>
      </h3>
      <p v-if="article.summary">{{ article.summary }}</p>
    </div>

    <footer class="home-post-card__footer">
      <span>{{ article.view_count }} 次浏览</span>
      <span>{{ article.tags.slice(0, 2).join(' · ') || '随笔' }}</span>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
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
