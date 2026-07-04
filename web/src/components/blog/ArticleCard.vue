<template>
  <article class="article-card">
    <RouterLink class="article-card__title" :to="`/articles/${article.slug}`">
      {{ article.title }}
    </RouterLink>
    <p class="article-card__summary">{{ article.summary }}</p>
    <div class="article-card__meta">
      <span>{{ article.category }}</span>
      <span>{{ article.view_count }} views</span>
    </div>
  </article>
</template>

<script setup lang="ts">
import type { ArticleListItem } from '@/types/article'

defineProps<{
  article: ArticleListItem
}>()
</script>

