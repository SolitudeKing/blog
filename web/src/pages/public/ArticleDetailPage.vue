<template>
  <article class="article-detail" v-if="article">
    <p class="article-detail__meta">{{ article.category }} · {{ article.view_count }} views</p>
    <h1>{{ article.title }}</h1>
    <pre class="article-detail__content">{{ article.content_md }}</pre>
  </article>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getArticleDetail } from '@/api/modules/article'
import type { ArticleDetail } from '@/types/article'

const route = useRoute()
const article = ref<ArticleDetail | null>(null)

onMounted(async () => {
  article.value = await getArticleDetail(String(route.params.slug))
})
</script>

