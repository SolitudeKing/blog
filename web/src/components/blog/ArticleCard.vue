<template>
  <article
    class="home-post-card mist-card mist-card--hoverable"
    :class="[
      `home-post-card--${variant}`,
      { 'home-post-card--offset': offset },
    ]"
    :id="domId"
    :aria-labelledby="headingId"
  >
    <div class="home-post-card__top">
      <span class="home-post-card__tag">{{ article.category || 'Notes' }}</span>
      <time :datetime="article.published_at || article.created_at">{{ displayDate }}</time>
    </div>

    <div v-if="variant === 'featured'" class="home-post-card__diagram" aria-hidden="true">
      <svg viewBox="0 0 680 220">
        <rect x="36" y="50" width="164" height="116" rx="16" />
        <rect x="258" y="28" width="164" height="70" rx="12" />
        <rect x="258" y="124" width="164" height="70" rx="12" />
        <rect x="480" y="76" width="164" height="70" rx="12" />
        <path d="M200 108h58M422 63h28c18 0 30 12 30 30v18M422 159h28c18 0 30-12 30-30v-18" />
        <circle cx="230" cy="108" r="4" />
        <path d="M68 84h98M68 108h72M68 132h84" />
      </svg>
    </div>

    <div class="home-post-card__body">
      <span class="home-post-card__index">POST / {{ paddedIndex }}</span>
      <h3>
        <RouterLink :id="headingId" :to="`/articles/${article.slug}`">
          {{ article.title }}
        </RouterLink>
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
    variant?: 'featured' | 'compact' | 'stream'
    offset?: boolean
  }>(),
  {
    domId: undefined,
    index: 1,
    variant: 'compact',
    offset: false,
  },
)

const headingId = computed(() => `${props.domId || `article-${props.article.id}`}-title`)
const paddedIndex = computed(() => String(props.index).padStart(3, '0'))

const displayDate = computed(() => {
  const raw = props.article.published_at || props.article.created_at
  if (!raw) {
    return '未发布'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: props.variant === 'featured' ? 'numeric' : undefined,
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(raw))
})
</script>
