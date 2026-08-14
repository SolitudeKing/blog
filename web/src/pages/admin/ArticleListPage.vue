<template>
  <section class="admin-page" :aria-busy="loading || loadingMore">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Articles</p>
        <div role="heading" aria-level="1">文章管理</div>
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
        @keyup.enter="reload"
      />
      <BaseSelect
        v-model="status"
        :options="statusOptions"
        label="文章状态"
        @change="reload"
      />
      <BaseButton variant="secondary" :loading="loading" @click="reload">刷新</BaseButton>
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
      :aria-rowcount="page.has_more ? -1 : articles.length + 1"
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
        <SvgIcon class="admin-empty__icon" name="empty-article" />
      </template>
    </BaseEmpty>

    <div v-if="articles.length" class="admin-list-footer">
      <BasePagination
        mode="prevNext"
        :page="page.page"
        :has-more="page.has_more"
        :loading="loadingMore"
        @prev="goPrevPage"
        @next="goNextPage"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BasePagination from '@/components/base/BasePagination.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { deleteArticle, getManagedArticleList } from '@/api/modules/article'
import { useToast } from '@/composables/useToast'
import type { ArticleListItem } from '@/types/article'

const articles = ref<ArticleListItem[]>([])
const route = useRoute()
const keyword = ref('')
const status = ref<string>('')
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')
const toast = useToast()
let articleRequestId = 0

const page = reactive({
  page: 1,
  page_size: 20,
  count: 0,
  has_more: false,
})

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
    void reload()
  },
  { immediate: true },
)

function resetPage() {
  page.page = 1
  page.has_more = false
  page.count = 0
}

async function reload() {
  articleRequestId += 1
  articles.value = []
  resetPage()
  await loadArticles(1)
}

async function loadArticles(targetPage: number) {
  const requestId = ++articleRequestId
  if (targetPage === 1) {
    loading.value = true
  } else {
    loadingMore.value = true
  }
  error.value = ''
  try {
    const result = await getManagedArticleList({
      page: targetPage,
      page_size: page.page_size,
      keyword: keyword.value || undefined,
      status: status.value || undefined,
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
      error.value = err instanceof Error ? err.message : '加载文章失败'
    }
  } finally {
    if (requestId === articleRequestId) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

async function goNextPage() {
  if (loadingMore.value || !page.has_more) return
  await loadArticles(page.page + 1)
}

async function goPrevPage() {
  if (loadingMore.value || page.page <= 1) return
  await loadArticles(page.page - 1)
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
