<template>
  <section class="admin-page dashboard-page">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Dashboard</p>
        <h1>仪表盘</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadSummary">刷新</BaseButton>
    </header>

    <p v-if="error" class="admin-page__error">{{ error }}</p>

    <div class="dashboard-metrics">
      <section v-for="metric in metrics" :key="metric.label" class="dashboard-metric">
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <p>{{ metric.detail }}</p>
      </section>
    </div>

    <div class="dashboard-grid">
      <section class="dashboard-panel">
        <div class="dashboard-panel__header">
          <h2>最近文章</h2>
          <RouterLink class="text-link" to="/admin/articles">查看全部</RouterLink>
        </div>
        <div v-if="summary?.recent_articles.length" class="dashboard-list">
          <RouterLink
            v-for="article in summary.recent_articles"
            :key="article.id"
            class="dashboard-list__item"
            :to="`/admin/articles/${article.id}`"
          >
            <div>
              <strong>{{ article.title }}</strong>
              <p>{{ formatTime(article.updated_at) }}</p>
            </div>
            <span class="status-pill" :class="`status-pill--${article.status}`">{{ statusText(article.status) }}</span>
          </RouterLink>
        </div>
        <div v-else class="empty-state">暂无文章</div>
      </section>

      <section class="dashboard-panel">
        <div class="dashboard-panel__header">
          <h2>当前公告</h2>
          <RouterLink class="text-link" to="/admin/notices">管理公告</RouterLink>
        </div>
        <article v-if="summary?.active_notice" class="dashboard-notice">
          <strong>{{ summary.active_notice.title }}</strong>
          <p>{{ summary.active_notice.content }}</p>
          <span>{{ formatNoticeRange(summary.active_notice) }}</span>
        </article>
        <div v-else class="empty-state">暂无启用公告</div>
      </section>

      <section class="dashboard-panel dashboard-panel--wide">
        <div class="dashboard-panel__header">
          <h2>快捷入口</h2>
        </div>
        <div class="dashboard-actions">
          <RouterLink class="cui-button" to="/admin/articles/new">新建文章</RouterLink>
          <RouterLink class="cui-button cui-button--secondary" to="/admin/taxonomy">分类标签</RouterLink>
          <RouterLink class="cui-button cui-button--secondary" to="/admin/notices">发布公告</RouterLink>
          <RouterLink class="cui-button cui-button--secondary" to="/admin/settings">站点设置</RouterLink>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import { getDashboardSummary } from '@/api/modules/dashboard'
import type { DashboardArticleItem, DashboardNoticeItem, DashboardSummary } from '@/types/dashboard'

const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')

const metrics = computed(() => {
  const data = summary.value
  return [
    {
      label: '文章总数',
      value: data?.article_counts.total ?? '-',
      detail: `已发布 ${data?.article_counts.published ?? 0} / 草稿 ${data?.article_counts.draft ?? 0}`,
    },
    {
      label: '阅读量',
      value: data?.total_views ?? '-',
      detail: '文章累计浏览',
    },
    {
      label: '分类标签',
      value: data ? data.taxonomy_counts.categories + data.taxonomy_counts.tags : '-',
      detail: `分类 ${data?.taxonomy_counts.categories ?? 0} / 标签 ${data?.taxonomy_counts.tags ?? 0}`,
    },
    {
      label: '公告',
      value: data?.notice_counts.total ?? '-',
      detail: `启用 ${data?.notice_counts.enabled ?? 0}`,
    },
  ]
})

onMounted(() => {
  loadSummary()
})

async function loadSummary() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await getDashboardSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载仪表盘失败'
  } finally {
    loading.value = false
  }
}

function statusText(value: DashboardArticleItem['status']) {
  const map: Record<DashboardArticleItem['status'], string> = {
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

function formatNoticeRange(notice: DashboardNoticeItem) {
  if (!notice.starts_at && !notice.ends_at) {
    return '长期有效'
  }
  return `${notice.starts_at ? formatTime(notice.starts_at) : '立即'} - ${notice.ends_at ? formatTime(notice.ends_at) : '不限'}`
}
</script>
