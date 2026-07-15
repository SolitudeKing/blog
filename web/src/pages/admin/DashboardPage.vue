<template>
  <section class="admin-page dashboard-page" :aria-busy="loading">
    <header class="dashboard-welcome">
      <div>
        <p class="admin-page__eyebrow">{{ todayLabel }}</p>
        <h1>创作工作台</h1>
        <p>从最近一次编辑继续，让内容发布保持清晰而从容。</p>
      </div>
      <div class="dashboard-welcome__actions">
        <BaseButton variant="secondary" :loading="loading" @click="loadSummary">
          刷新数据
        </BaseButton>
        <RouterLink class="mist-button mist-button--lg" to="/admin/articles/new">
          <svg aria-hidden="true" focusable="false" viewBox="0 0 24 24">
            <path d="M12 5v14M5 12h14" />
          </svg>
          新建文章
        </RouterLink>
      </div>
    </header>

    <section
      v-if="loading"
      class="dashboard-state"
      role="status"
      aria-live="polite"
      aria-labelledby="dashboard-loading-title"
    >
      <span class="dashboard-state__indicator" aria-hidden="true" />
      <div>
        <h2 id="dashboard-loading-title">正在整理工作台</h2>
        <p>正在读取文章、媒体与站点统计。</p>
      </div>
    </section>

    <section
      v-else-if="error"
      class="dashboard-state dashboard-state--error"
      role="alert"
      aria-labelledby="dashboard-error-title"
    >
      <div>
        <h2 id="dashboard-error-title">工作台暂时无法载入</h2>
        <p>{{ error }}</p>
      </div>
      <BaseButton variant="secondary" @click="loadSummary">重新加载</BaseButton>
    </section>

    <template v-else-if="summary">
      <section class="dashboard-focus mist-luminous" aria-labelledby="dashboard-focus-title">
        <article class="dashboard-focus__primary">
          <div class="dashboard-focus__eyebrow">
            <span class="admin-page__eyebrow">
              {{ currentDraft ? 'Current draft · 当前草稿' : 'Recently edited · 最近编辑' }}
            </span>
            <span v-if="focusArticle" class="status-pill" :class="`status-pill--${focusArticle.status}`">
              {{ statusText(focusArticle.status) }}
            </span>
          </div>

          <template v-if="focusArticle">
            <h2 id="dashboard-focus-title">{{ focusArticle.title }}</h2>
            <p>
              {{ currentDraft ? '这篇草稿仍在等待完成，继续上一次的写作节奏。' : '这是最近更新的内容，可继续编辑或查看发布结果。' }}
            </p>
            <dl class="dashboard-focus__meta">
              <div>
                <dt>最近编辑</dt>
                <dd>{{ formatTime(focusArticle.updated_at) }}</dd>
              </div>
              <div>
                <dt>内容状态</dt>
                <dd>{{ statusText(focusArticle.status) }}</dd>
              </div>
            </dl>
            <div class="dashboard-focus__actions">
              <RouterLink class="mist-button" :to="`/admin/articles/${focusArticle.id}`">
                {{ currentDraft ? '继续写作' : '打开编辑器' }}
                <svg aria-hidden="true" focusable="false" viewBox="0 0 24 24">
                  <path d="M5 12h14M13 6l6 6-6 6" />
                </svg>
              </RouterLink>
              <RouterLink
                v-if="focusArticle.status === 'published'"
                class="mist-button mist-button--ghost"
                :to="`/articles/${focusArticle.slug}`"
              >
                查看文章
              </RouterLink>
            </div>
          </template>

          <template v-else>
            <h2 id="dashboard-focus-title">把第一篇文章写进今天</h2>
            <p>工作台已经准备好，从一个标题或一段 Markdown 开始。</p>
            <div class="dashboard-focus__actions">
              <RouterLink class="mist-button" to="/admin/articles/new">开始写作</RouterLink>
            </div>
          </template>
        </article>

        <aside class="dashboard-briefing" aria-labelledby="dashboard-briefing-title">
          <header>
            <div>
              <span class="admin-page__eyebrow">Publishing</span>
              <h2 id="dashboard-briefing-title">当前公告</h2>
            </div>
            <strong>{{ summary.notice_counts.enabled.toString().padStart(2, '0') }}</strong>
          </header>

          <article v-if="summary.active_notice" class="dashboard-briefing__notice">
            <strong>{{ summary.active_notice.title }}</strong>
            <p>{{ summary.active_notice.content }}</p>
            <span>{{ formatNoticeRange(summary.active_notice) }}</span>
          </article>
          <div v-else class="dashboard-briefing__empty">
            <strong>当前没有启用公告</strong>
            <p>需要向访客传递消息时，可在公告管理中发布。</p>
          </div>

          <footer>
            <RouterLink class="text-link" to="/admin/notices">管理公告</RouterLink>
            <span>更新于 {{ formatTime(summary.generated_at) }}</span>
          </footer>
        </aside>
      </section>

      <section
        class="dashboard-metrics"
        tabindex="0"
        role="region"
        aria-labelledby="dashboard-metrics-title"
      >
        <h2 id="dashboard-metrics-title" class="sr-only">站点数据概览</h2>
        <dl>
          <div
            v-for="metric in metrics"
            :key="metric.label"
            class="dashboard-metric"
            :class="{ 'dashboard-metric--featured': metric.featured }"
          >
            <dt>{{ metric.label }}</dt>
            <dd>
              <strong>{{ formatNumber(metric.value) }}</strong>
              <small>{{ metric.detail }}</small>
            </dd>
          </div>
        </dl>
      </section>

      <section class="dashboard-content-grid" aria-label="近期内容与站点数据">
        <section class="dashboard-stream" aria-labelledby="dashboard-recent-title">
          <header class="dashboard-section-header">
            <div>
              <span class="admin-page__eyebrow">Content stream</span>
              <h2 id="dashboard-recent-title">最近文章</h2>
            </div>
            <RouterLink class="text-link" to="/admin/articles">查看全部</RouterLink>
          </header>

          <ol v-if="summary.recent_articles.length" class="dashboard-article-list">
            <li v-for="article in summary.recent_articles" :key="article.id">
              <RouterLink :to="`/admin/articles/${article.id}`">
                <div>
                  <strong>{{ article.title }}</strong>
                  <time :datetime="article.updated_at">{{ formatTime(article.updated_at) }}</time>
                </div>
                <span class="status-pill" :class="`status-pill--${article.status}`">
                  {{ statusText(article.status) }}
                </span>
                <svg aria-hidden="true" focusable="false" viewBox="0 0 24 24">
                  <path d="M5 12h14M13 6l6 6-6 6" />
                </svg>
              </RouterLink>
            </li>
          </ol>
          <div v-else class="dashboard-inline-empty">
            <strong>还没有文章</strong>
            <p>创建第一篇文章后，最近编辑记录会出现在这里。</p>
          </div>
        </section>

        <aside class="dashboard-side" aria-label="热门文章与分类分布">
          <section aria-labelledby="dashboard-top-title">
            <header class="dashboard-section-header">
              <div>
                <span class="admin-page__eyebrow">Readers</span>
                <h2 id="dashboard-top-title">热门文章</h2>
              </div>
            </header>
            <ol v-if="summary.top_articles.length" class="dashboard-ranked-list">
              <li v-for="(article, index) in summary.top_articles" :key="article.id">
                <RouterLink :to="`/admin/articles/${article.id}`">
                  <span>{{ String(index + 1).padStart(2, '0') }}</span>
                  <strong>{{ article.title }}</strong>
                  <small>{{ formatNumber(article.view_count) }} 阅读</small>
                </RouterLink>
              </li>
            </ol>
            <div v-else class="dashboard-inline-empty dashboard-inline-empty--compact">
              <p>暂无阅读数据</p>
            </div>
          </section>

          <section aria-labelledby="dashboard-category-title">
            <header class="dashboard-section-header">
              <div>
                <span class="admin-page__eyebrow">Taxonomy</span>
                <h2 id="dashboard-category-title">分类分布</h2>
              </div>
              <RouterLink class="text-link" to="/admin/taxonomy">管理</RouterLink>
            </header>
            <div v-if="summary.category_stats.length" class="dashboard-bars">
              <div v-for="item in summary.category_stats" :key="item.id" class="dashboard-bar">
                <div>
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.article_count }}</span>
                </div>
                <i :style="{ width: categoryPercent(item.article_count) + '%' }" aria-hidden="true" />
              </div>
            </div>
            <div v-else class="dashboard-inline-empty dashboard-inline-empty--compact">
              <p>暂无分类统计</p>
            </div>
          </section>
        </aside>
      </section>
    </template>
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

const todayLabel = new Intl.DateTimeFormat('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'long',
}).format(new Date())

const currentDraft = computed(() =>
  summary.value?.recent_articles.find((article) => article.status === 'draft') ?? null,
)
const focusArticle = computed(() => currentDraft.value ?? summary.value?.recent_articles[0] ?? null)
const metrics = computed(() => {
  const data = summary.value
  if (!data) {
    return []
  }
  return [
    {
      label: '文章总数',
      value: data.article_counts.total,
      detail: `已发布 ${data.article_counts.published}`,
    },
    {
      label: '草稿箱',
      value: data.article_counts.draft,
      detail: `私有 ${data.article_counts.private} · 归档 ${data.article_counts.archived}`,
    },
    {
      label: '累计阅读',
      value: data.total_views,
      detail: '全部已发布文章',
      featured: true,
    },
    {
      label: '分类与标签',
      value: data.taxonomy_counts.categories + data.taxonomy_counts.tags,
      detail: `分类 ${data.taxonomy_counts.categories} · 标签 ${data.taxonomy_counts.tags}`,
    },
    {
      label: '媒体资源',
      value: data.asset_count,
      detail: '已上传文件',
    },
  ]
})

onMounted(() => {
  void loadSummary()
})

async function loadSummary() {
  loading.value = true
  error.value = ''
  summary.value = null
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

function formatNumber(value: number) {
  return new Intl.NumberFormat('zh-CN', { notation: 'compact', maximumFractionDigits: 1 }).format(value)
}

function formatTime(value: string) {
  return new Date(value).toLocaleString('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatNoticeRange(notice: DashboardNoticeItem) {
  if (!notice.starts_at && !notice.ends_at) {
    return '长期有效'
  }
  return `${notice.starts_at ? formatTime(notice.starts_at) : '立即'} - ${notice.ends_at ? formatTime(notice.ends_at) : '不限'}`
}

function categoryPercent(value: number) {
  const max = Math.max(...(summary.value?.category_stats.map((item) => item.article_count) ?? [0]))
  if (!max) {
    return 0
  }
  return Math.max(8, Math.round((value / max) * 100))
}
</script>
