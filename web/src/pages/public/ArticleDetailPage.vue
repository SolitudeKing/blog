<template>
  <section class="article-detail-page">
    <article v-if="article" class="article-detail">
      <header class="article-detail__header">
        <p class="article-detail__meta">{{ article.category }} · {{ article.view_count }} views</p>
        <h1>{{ article.title }}</h1>
        <p v-if="article.summary" class="article-detail__summary">{{ article.summary }}</p>
      </header>

      <div class="article-detail__layout">
        <div class="markdown-body" v-html="rendered.html" />
        <aside v-if="rendered.toc.length" class="article-toc">
          <strong>目录</strong>
          <a
            v-for="item in rendered.toc"
            :key="item.id"
            :href="`#${item.id}`"
            :style="{ paddingLeft: `${(item.level - 2) * 12}px` }"
          >
            {{ item.text }}
          </a>
        </aside>
      </div>
    </article>

    <div v-else-if="loading" class="page-state">文章加载中</div>
    <div v-else class="page-state page-state--error">
      <h1>{{ error || '文章不存在' }}</h1>
      <RouterLink class="cui-button cui-button--secondary" to="/">返回首页</RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { getArticleDetail } from '@/api/modules/article'
import type { ArticleDetail } from '@/types/article'
import { renderMarkdown } from '@/utils/markdown'

const route = useRoute()
const article = ref<ArticleDetail | null>(null)
const loading = ref(false)
const error = ref('')

const rendered = computed(() => renderMarkdown(article.value?.content_md ?? ''))

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    article.value = await getArticleDetail(String(route.params.slug))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    loading.value = false
  }
})
</script>
