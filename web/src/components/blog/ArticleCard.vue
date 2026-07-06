<template>
  <article class="article-card" :id="domId">
    <div class="article-card__topline">
      <span>{{ article.category || 'Notes' }}</span>
      <time :datetime="article.published_at || article.created_at">{{ displayDate }}</time>
    </div>

    <RouterLink class="article-card__title" :to="`/articles/${article.slug}`">
      {{ article.title }}
    </RouterLink>

    <p v-if="article.summary" class="article-card__summary">{{ article.summary }}</p>

    <div class="article-card__meta">
      <span v-for="tag in article.tags" :key="tag" class="article-card__tag">{{ tag }}</span>
      <span>{{ article.view_count }} views</span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ArticleListItem } from '@/types/article'

const props = defineProps<{
  article: ArticleListItem
  domId?: string
}>()

const displayDate = computed(() => {
  const raw = props.article.published_at || props.article.created_at
  if (!raw) {
    return '未发布'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(new Date(raw))
})
</script>
