<template>
  <section class="admin-page">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Articles</p>
        <h1>文章管理</h1>
      </div>
      <RouterLink class="cui-button" to="/admin/articles/new">新建文章</RouterLink>
    </header>

    <div class="admin-toolbar">
      <input v-model="keyword" class="cui-input" placeholder="搜索标题" @keyup.enter="loadArticles()" />
      <select v-model="status" class="cui-input" @change="loadArticles()">
        <option value="">全部状态</option>
        <option value="draft">草稿</option>
        <option value="published">已发布</option>
        <option value="private">私有</option>
        <option value="archived">归档</option>
      </select>
      <BaseButton variant="secondary" :loading="loading" @click="loadArticles()">刷新</BaseButton>
    </div>

    <p v-if="error" class="admin-page__error">{{ error }}</p>

    <div class="admin-table">
      <div class="admin-table__row admin-table__row--head">
        <span>标题</span>
        <span>状态</span>
        <span>分类</span>
        <span>更新时间</span>
        <span>操作</span>
      </div>
      <div v-for="article in articles" :key="article.id" class="admin-table__row">
        <div>
          <strong>{{ article.title }}</strong>
          <p>{{ article.summary || article.slug }}</p>
        </div>
        <span class="status-pill" :class="`status-pill--${article.status}`">{{ statusText(article.status) }}</span>
        <span>{{ article.category }}</span>
        <span>{{ formatTime(article.updated_at) }}</span>
        <div class="admin-table__actions">
          <RouterLink class="text-link" :to="`/admin/articles/${article.id}`">编辑</RouterLink>
          <RouterLink v-if="article.status === 'published'" class="text-link" :to="`/articles/${article.slug}`">
            预览
          </RouterLink>
          <button class="text-link text-link--danger" type="button" @click="removeArticle(article.id)">删除</button>
        </div>
      </div>
    </div>

    <div v-if="!loading && articles.length === 0" class="empty-state">暂无文章</div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { deleteArticle, getManagedArticleList } from '@/api/modules/article'
import type { ArticleListItem } from '@/types/article'

const articles = ref<ArticleListItem[]>([])
const keyword = ref('')
const status = ref('')
const loading = ref(false)
const error = ref('')

onMounted(() => {
  loadArticles()
})

async function loadArticles() {
  loading.value = true
  error.value = ''
  try {
    const result = await getManagedArticleList({
      keyword: keyword.value || undefined,
      status: status.value || undefined,
    })
    articles.value = result.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    loading.value = false
  }
}

async function removeArticle(id: number) {
  if (!window.confirm('确认删除这篇文章？')) {
    return
  }
  try {
    await deleteArticle(id)
    articles.value = articles.value.filter((article) => article.id !== id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除文章失败'
  }
}

function statusText(value: ArticleListItem['status']) {
  const map: Record<ArticleListItem['status'], string> = {
    draft: '草稿',
    published: '已发布',
    private: '私有',
    archived: '归档',
  }
  return map[value]
}

function formatTime(value: string) {
  return new Date(value).toLocaleString()
}
</script>
