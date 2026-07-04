<template>
  <section class="admin-page">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Notices</p>
        <h1>公告管理</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadNotices()">刷新</BaseButton>
    </header>

    <div class="notice-grid">
      <form class="notice-panel" @submit.prevent="saveNotice">
        <h2>{{ editingId ? '编辑公告' : '新增公告' }}</h2>
        <BaseInput v-model="form.title" label="标题" />
        <BaseTextarea v-model="form.content" label="内容" :rows="5" />

        <div class="notice-form__row">
          <BaseInput v-model="sortOrder" label="排序" type="number" />
          <label class="notice-switch">
            <input v-model="form.enabled" type="checkbox" />
            <span>启用公告</span>
          </label>
        </div>

        <div class="notice-form__row">
          <label class="cui-field">
            <span class="cui-field__label">生效时间</span>
            <input v-model="startsAtInput" class="cui-input" type="datetime-local" />
          </label>
          <label class="cui-field">
            <span class="cui-field__label">失效时间</span>
            <input v-model="endsAtInput" class="cui-input" type="datetime-local" />
          </label>
        </div>

        <div class="notice-form__actions">
          <BaseButton type="submit" :loading="saving">{{ editingId ? '保存修改' : '创建公告' }}</BaseButton>
          <BaseButton v-if="editingId" type="button" variant="secondary" @click="resetForm">取消</BaseButton>
        </div>
        <p v-if="error" class="admin-page__error">{{ error }}</p>
      </form>

      <section class="notice-list">
        <div class="admin-toolbar notice-toolbar">
          <input v-model="keyword" class="cui-input" placeholder="搜索公告" @keyup.enter="loadNotices()" />
          <select v-model="enabledFilter" class="cui-input" @change="loadNotices()">
            <option value="">全部状态</option>
            <option value="true">启用</option>
            <option value="false">停用</option>
          </select>
          <BaseButton variant="secondary" :loading="loading" @click="loadNotices()">筛选</BaseButton>
        </div>

        <div class="admin-table">
          <div class="admin-table__row admin-table__row--head notice-table__row">
            <span>标题</span>
            <span>状态</span>
            <span>有效期</span>
            <span>操作</span>
          </div>
          <div v-for="notice in notices" :key="notice.id" class="admin-table__row notice-table__row">
            <div>
              <strong>{{ notice.title }}</strong>
              <p>{{ notice.content }}</p>
            </div>
            <span class="status-pill" :class="notice.enabled ? 'status-pill--published' : 'status-pill--archived'">
              {{ notice.enabled ? '启用' : '停用' }}
            </span>
            <span>{{ formatRange(notice) }}</span>
            <div class="admin-table__actions">
              <button class="text-link" type="button" @click="editNotice(notice)">编辑</button>
              <button class="text-link text-link--danger" type="button" @click="removeNotice(notice.id)">删除</button>
            </div>
          </div>
        </div>

        <div v-if="!loading && notices.length === 0" class="empty-state">暂无公告</div>

        <div v-if="notices.length && page.has_more" class="notice-list__footer">
          <BaseButton variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import { createNotice, deleteNotice, getManagedNoticeList, updateNotice } from '@/api/modules/notice'
import type { CursorPage } from '@/api/types'
import type { NoticeItem, NoticePayload } from '@/types/notice'

const notices = ref<NoticeItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const saving = ref(false)
const error = ref('')
const keyword = ref('')
const enabledFilter = ref('')
const editingId = ref<number | null>(null)

const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 20,
  has_more: false,
})

const form = reactive<NoticePayload>({
  title: '',
  content: '',
  enabled: true,
  sort_order: 0,
  starts_at: null,
  ends_at: null,
})

const sortOrder = computed({
  get: () => String(form.sort_order),
  set: (value: string) => {
    form.sort_order = Number(value) || 0
  },
})

const startsAtInput = computed({
  get: () => toLocalInputValue(form.starts_at),
  set: (value: string) => {
    form.starts_at = fromLocalInputValue(value)
  },
})

const endsAtInput = computed({
  get: () => toLocalInputValue(form.ends_at),
  set: (value: string) => {
    form.ends_at = fromLocalInputValue(value)
  },
})

onMounted(() => {
  loadNotices()
})

async function loadNotices(cursor = '') {
  loading.value = true
  error.value = ''
  try {
    const result = await getManagedNoticeList({
      cursor: cursor || undefined,
      keyword: keyword.value || undefined,
      enabled: enabledFilter.value ? enabledFilter.value === 'true' : undefined,
      limit: page.limit,
    })
    notices.value = cursor ? [...notices.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载公告失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!page.next_cursor) {
    return
  }
  loadingMore.value = true
  try {
    await loadNotices(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}

async function saveNotice() {
  saving.value = true
  error.value = ''
  try {
    const payload = { ...form }
    if (editingId.value) {
      await updateNotice(editingId.value, payload)
    } else {
      await createNotice(payload)
    }
    resetForm()
    await loadNotices()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存公告失败'
  } finally {
    saving.value = false
  }
}

function editNotice(notice: NoticeItem) {
  editingId.value = notice.id
  form.title = notice.title
  form.content = notice.content
  form.enabled = notice.enabled
  form.sort_order = notice.sort_order
  form.starts_at = notice.starts_at
  form.ends_at = notice.ends_at
}

async function removeNotice(id: number) {
  if (!window.confirm('确认删除这条公告？')) {
    return
  }
  try {
    await deleteNotice(id)
    notices.value = notices.value.filter((notice) => notice.id !== id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除公告失败'
  }
}

function resetForm() {
  editingId.value = null
  form.title = ''
  form.content = ''
  form.enabled = true
  form.sort_order = 0
  form.starts_at = null
  form.ends_at = null
}

function formatRange(notice: NoticeItem) {
  if (!notice.starts_at && !notice.ends_at) {
    return '长期有效'
  }
  return `${formatTime(notice.starts_at) || '立即'} - ${formatTime(notice.ends_at) || '不限'}`
}

function formatTime(value: string | null) {
  if (!value) {
    return ''
  }
  return new Date(value).toLocaleString()
}

function toLocalInputValue(value: string | null) {
  if (!value) {
    return ''
  }
  const date = new Date(value)
  const offset = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

function fromLocalInputValue(value: string) {
  if (!value) {
    return null
  }
  return new Date(value).toISOString()
}
</script>
