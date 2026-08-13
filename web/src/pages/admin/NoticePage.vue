<template>
  <section class="admin-page" :aria-busy="loading || saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Notices</p>
        <div role="heading" aria-level="1">公告管理</div>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadNotices()">刷新</BaseButton>
    </header>

    <div class="notice-grid">
      <form class="notice-panel" :aria-busy="saving" @submit.prevent="saveNotice">
        <div role="heading" aria-level="2">{{ editingId ? '编辑公告' : '新增公告' }}</div>
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
          <label class="mist-field">
            <span class="mist-field__label">生效时间</span>
            <input v-model="startsAtInput" class="mist-input" type="datetime-local" />
          </label>
          <label class="mist-field">
            <span class="mist-field__label">失效时间</span>
            <input v-model="endsAtInput" class="mist-input" type="datetime-local" />
          </label>
        </div>

        <div class="notice-form__actions">
          <BaseButton type="submit" :loading="saving">{{ editingId ? '保存修改' : '创建公告' }}</BaseButton>
          <BaseButton v-if="editingId" type="button" variant="secondary" @click="resetForm">取消</BaseButton>
        </div>
        <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
      </form>

      <section class="notice-list">
        <div class="admin-toolbar notice-toolbar" role="search" aria-label="筛选公告">
          <input
            v-model="keyword"
            class="mist-input"
            type="search"
            aria-label="搜索公告"
            placeholder="搜索公告"
            @keyup.enter="loadNotices()"
          />
          <BaseSelect
            v-model="enabledFilter"
            :options="enabledOptions"
            label="公告状态"
            @change="loadNotices()"
          />
          <BaseButton variant="secondary" :loading="loading" @click="loadNotices()">筛选</BaseButton>
        </div>

        <p v-if="loading && !notices.length" class="admin-page__status" role="status" aria-live="polite">
          正在加载公告…
        </p>

        <div
          v-if="notices.length"
          class="admin-table"
          role="table"
          aria-label="公告列表"
          :aria-rowcount="notices.length + 1"
        >
          <div class="admin-table__row admin-table__row--head notice-table__row" role="row">
            <span class="admin-table__cell" role="columnheader">标题</span>
            <span class="admin-table__cell" role="columnheader">状态</span>
            <span class="admin-table__cell" role="columnheader">有效期</span>
            <span class="admin-table__cell" role="columnheader">操作</span>
          </div>
          <div v-for="notice in notices" :key="notice.id" class="admin-table__row notice-table__row" role="row">
            <div class="admin-table__cell" role="cell" data-label="标题">
              <strong>{{ notice.title }}</strong>
              <p>{{ notice.content }}</p>
            </div>
            <div class="admin-table__cell" role="cell" data-label="状态">
              <span
                class="status-pill"
                :class="notice.enabled ? 'status-pill--published' : 'status-pill--archived'"
              >
                {{ notice.enabled ? '启用' : '停用' }}
              </span>
            </div>
            <span class="admin-table__cell" role="cell" data-label="有效期">{{ formatRange(notice) }}</span>
            <div class="admin-table__cell admin-table__actions" role="cell" data-label="操作">
              <div class="admin-table__action-list">
                <button class="text-link" type="button" @click="editNotice(notice)">编辑</button>
                <button class="text-link text-link--danger" type="button" @click="removeNotice(notice.id)">
                  删除
                </button>
              </div>
            </div>
          </div>
        </div>

        <BaseEmpty
          v-else-if="!loading"
          title="暂无公告"
          description="创建一条公告，让访客在首页第一眼看到。"
        >
          <template #icon>
            <SvgIcon class="admin-empty__icon" name="empty-notice" />
          </template>
        </BaseEmpty>

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
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { createNotice, deleteNotice, getManagedNoticeList, updateNotice } from '@/api/modules/notice'
import { useToast } from '@/composables/useToast'
import type { CursorPage } from '@/api/types'
import type { NoticeItem, NoticePayload } from '@/types/notice'

const notices = ref<NoticeItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const saving = ref(false)
const error = ref('')
const keyword = ref('')
const enabledFilter = ref<string>('')
const editingId = ref<number | null>(null)
const toast = useToast()

const enabledOptions = [
  { label: '全部状态', value: '' },
  { label: '启用', value: 'true' },
  { label: '停用', value: 'false' },
]

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
      toast.success('公告已更新')
    } else {
      await createNotice(payload)
      toast.success('公告已创建')
    }
    resetForm()
    await loadNotices()
  } catch (err) {
    const message = err instanceof Error ? err.message : '保存公告失败'
    error.value = message
    toast.error(message)
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
    toast.success('公告已删除')
  } catch (err) {
    const message = err instanceof Error ? err.message : '删除公告失败'
    error.value = message
    toast.error(message)
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
