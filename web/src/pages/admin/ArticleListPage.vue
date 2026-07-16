<template>
  <section class="admin-page" :aria-busy="loading">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Articles</p>
        <h1>文章管理</h1>
      </div>
      <RouterLink class="mist-button" to="/admin/articles/new">新建文章</RouterLink>
    </header>

    <div class="admin-toolbar" role="search" aria-label="筛选文章">
      <input
        v-model="keyword"
        class="mist-input"
        type="search"
        aria-label="搜索文章标题"
        placeholder="搜索标题"
        @keyup.enter="loadArticles()"
      />
      <BaseSelect
        v-model="status"
        :options="statusOptions"
        label="文章状态"
        @change="loadArticles()"
      />
      <BaseButton variant="secondary" :loading="loading" @click="loadArticles()">刷新</BaseButton>
    </div>

    <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
    <p v-else-if="loading && !articles.length" class="admin-page__status" role="status" aria-live="polite">
      正在加载文章…
    </p>

    <div
      v-if="articles.length"
      class="admin-table"
      role="table"
      aria-label="文章列表"
      :aria-rowcount="articles.length + 1"
    >
      <div class="admin-table__row admin-table__row--head" role="row">
        <span class="admin-table__cell" role="columnheader">标题</span>
        <span class="admin-table__cell" role="columnheader">状态</span>
        <span class="admin-table__cell" role="columnheader">专题</span>
        <span class="admin-table__cell" role="columnheader">更新时间</span>
        <span class="admin-table__cell" role="columnheader">操作</span>
      </div>
      <div v-for="article in articles" :key="article.id" class="admin-table__row" role="row">
        <div class="admin-table__cell" role="cell" data-label="标题">
          <strong>{{ article.title }}</strong>
          <p>{{ article.summary || article.slug }}</p>
        </div>
        <div class="admin-table__cell" role="cell" data-label="状态">
          <span class="status-pill" :class="'status-pill--' + article.status">
            {{ statusText(article.status) }}
          </span>
        </div>
        <span class="admin-table__cell" role="cell" data-label="专题">
          {{ article.topic?.label || article.topic?.name || '未设置' }}
        </span>
        <span class="admin-table__cell" role="cell" data-label="更新时间">
          {{ formatTime(article.updated_at) }}
        </span>
        <div class="admin-table__cell admin-table__actions" role="cell" data-label="操作">
          <div class="admin-table__action-list">
            <RouterLink class="text-link" :to="'/admin/articles/' + article.id">编辑</RouterLink>
            <RouterLink
              v-if="article.status === 'published'"
              class="text-link"
              :to="'/articles/' + article.slug"
            >
              预览
            </RouterLink>
            <button class="text-link text-link--danger" type="button" @click="removeArticle(article.id)">
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <BaseEmpty
      v-else-if="!loading"
      title="还没有文章"
      description="点击右上角的“新建文章”开始写作。"
    >
      <template #icon>
        <svg class="admin-empty__icon" aria-hidden="true" focusable="false" viewBox="0 0 24 24">
          <path d="M6 3h9l4 4v14H6zM15 3v5h4M9 12h7M9 16h5" />
        </svg>
      </template>
    </BaseEmpty>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import { deleteArticle, getManagedArticleList } from '@/api/modules/article'
import { useToast } from '@/composables/useToast'
import type { ArticleListItem } from '@/types/article'

const articles = ref<ArticleListItem[]>([])
const route = useRoute()
const keyword = ref('')
const status = ref<string>('')
const loading = ref(false)
const error = ref('')
const toast = useToast()

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' },
  { label: '私有', value: 'private' },
  { label: '归档', value: 'archived' },
]

watch(
  () => route.query.keyword,
  (rawKeyword) => {
    keyword.value = typeof rawKeyword === 'string' ? rawKeyword.trim() : ''
    void loadArticles()
  },
  { immediate: true },
)

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
    toast.success('文章已删除')
  } catch (err) {
    const message = err instanceof Error ? err.message : '删除文章失败'
    error.value = message
    toast.error(message)
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
