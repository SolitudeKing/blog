<template>
  <section class="admin-page article-editor" :aria-busy="saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Editor</p>
        <div role="heading" aria-level="1">{{ isEditing ? '编辑文章' : '新建文章' }}</div>
      </div>
      <RouterLink class="mist-button mist-button--secondary" to="/admin/articles">返回列表</RouterLink>
    </header>

    <form class="editor-form" :aria-busy="saving" @submit.prevent="save">
      <div class="editor-form__main">
        <BaseInput v-model="form.title" label="标题" />
        <BaseInput v-model="form.slug" label="Slug" />
        <BaseTextarea v-model="form.summary" label="摘要" :rows="4" />
        <div class="article-cover">
          <div class="article-cover__field">
            <BaseInput
              v-model="form.cover_url"
              type="text"
              inputmode="url"
              label="文章预览图 URL"
              hint="可填写媒体库中的 /uploads/... 路径或 https://... 图片地址；点击右侧上传按钮直接上传。留空时首页使用默认预览图。"
              :maxlength="500"
            />
            <UploadButton class="article-cover__upload" @upload="onCoverUpload" />
          </div>
          <div v-if="form.cover_url" class="article-cover__preview">
            <img :src="form.cover_url" :alt="form.title || 'cover preview'" />
            <button class="text-link text-link--danger" type="button" @click="form.cover_url = ''">
              清除预览图
            </button>
          </div>
        </div>
        <BaseTextarea v-model="form.content_md" label="Markdown 正文" :rows="18" />
      </div>

      <aside class="editor-panel">
        <div class="editor-autosave">
          <strong role="status" aria-live="polite">{{ autosaveStatus }}</strong>
          <p v-if="draftAvailable">发现本地草稿，可恢复到编辑器。</p>
          <div v-if="draftAvailable" class="editor-autosave__actions">
            <BaseButton type="button" variant="secondary" @click="restoreDraft">恢复草稿</BaseButton>
            <button class="text-link text-link--danger" type="button" @click="discardDraft">丢弃</button>
          </div>
        </div>

        <BaseSelect
          v-model="form.status"
          :options="statusOptions"
          label="状态"
        />

        <BaseSelect
          v-model="topicSelection"
          :options="topicOptions"
          label="专题"
        />

        <div class="mist-field">
          <span id="article-tags-label" class="mist-field__label">标签</span>
          <div class="tag-checks" role="group" aria-labelledby="article-tags-label">
            <label v-for="tag in tags" :key="tag.id" class="tag-check">
              <input v-model="form.tag_ids" type="checkbox" :value="tag.id" />
              <span class="tag-dot" :style="{ background: safeTagColor(tag.color) }" />
              {{ tag.name }}
            </label>
          </div>
        </div>

        <div class="editor-panel__actions">
          <BaseButton type="submit" :loading="saving">{{ isEditing ? '保存修改' : '创建文章' }}</BaseButton>
          <BaseButton type="button" variant="secondary" :disabled="saving" @click="publish">保存并发布</BaseButton>
        </div>
        <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>

        <div v-if="isEditing" class="editor-versions">
          <div class="dashboard-panel__header">
            <div role="heading" aria-level="2">版本记录</div>
            <button class="text-link" type="button" @click="loadVersions">刷新</button>
          </div>
          <div v-if="versions.length" class="editor-version-list">
            <article v-for="version in versions" :key="version.id" class="editor-version">
              <div>
                <strong>{{ version.title }}</strong>
                <p>{{ formatTime(version.created_at) }}</p>
              </div>
              <BaseButton type="button" variant="secondary" @click="applyVersion(version)">套用</BaseButton>
            </article>
          </div>
          <p v-else class="editor-version-empty">暂无版本记录</p>
        </div>
      </aside>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import UploadButton from '@/components/media/UploadButton.vue'
import { createArticle, getArticleVersions, getManagedArticleInfo, updateArticle } from '@/api/modules/article'
import { getTagList, getTopicList } from '@/api/modules/taxonomy'
import { useToast } from '@/composables/useToast'
import type { ArticleSavePayload, ArticleVersionItem } from '@/types/article'
import type { AssetItem } from '@/types/asset'
import type { TagItem, TopicItem } from '@/types/taxonomy'

const route = useRoute()
const router = useRouter()
const saving = ref(false)
const error = ref('')
const isEditing = computed(() => route.name === 'admin-article-edit')
const topics = ref<TopicItem[]>([])
const tags = ref<TagItem[]>([])
const versions = ref<ArticleVersionItem[]>([])
const autosaveStatus = ref('自动保存待命')
const draftAvailable = ref(false)
const hydrated = ref(false)
let autosaveTimer: number | undefined

interface AutoSavedDraft {
  payload: ArticleSavePayload
  saved_at: string
}

const form = reactive<ArticleSavePayload>({
  title: '',
  slug: '',
  summary: '',
  cover_url: '',
  content_md: '',
  topic_id: 0,
  tag_ids: [],
  status: 'draft',
})

const toast = useToast()

const statusOptions: Array<{ label: string; value: ArticleSavePayload['status'] }> = [
  { label: '草稿', value: 'draft' },
  { label: '发布', value: 'published' },
  { label: '私有', value: 'private' },
  { label: '归档', value: 'archived' },
]

const topicSelection = computed<number>({
  get: () => form.topic_id || 0,
  set: (value) => {
    form.topic_id = Number(value) || 0
  },
})

const topicOptions = computed(() => [
  { label: '请选择专题', value: 0, disabled: true },
  ...topics.value.map((topic) => ({
    label: topic.label && topic.label !== topic.name ? `${topic.name} · ${topic.label}` : topic.name,
    value: topic.id,
  })),
])

onMounted(async () => {
  await loadTaxonomy()
  if (!isEditing.value) {
    loadLocalDraft()
    hydrated.value = true
    return
  }
  try {
    const article = await getManagedArticleInfo(String(route.params.id))
    form.title = article.title
    form.slug = article.slug
    form.summary = article.summary
    form.cover_url = article.cover_url
    form.content_md = article.content_md
    form.status = article.status
    form.topic_id = article.topic_id
    form.tag_ids = article.tag_ids
    await loadVersions()
    loadLocalDraft()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
  } finally {
    hydrated.value = true
  }
})

onBeforeUnmount(() => {
  if (autosaveTimer) {
    window.clearTimeout(autosaveTimer)
  }
})

watch(
  () => form.title,
  (title) => {
    if (isEditing.value || form.slug) {
      return
    }
    form.slug = toSlug(title)
  },
)

watch(
  form,
  () => {
    if (!hydrated.value) {
      return
    }
    scheduleAutosave()
  },
  { deep: true },
)

async function loadTaxonomy() {
  try {
    const [topicItems, tagItems] = await Promise.all([getTopicList(), getTagList()])
    topics.value = topicItems
    tags.value = tagItems
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载专题与标签失败'
  }
}

async function save() {
  await submit(form.status)
}

async function publish() {
  await submit('published')
}

async function submit(status: ArticleSavePayload['status']) {
  if (!Number(form.topic_id)) {
    error.value = '请选择文章专题'
    toast.error(error.value)
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload: ArticleSavePayload = {
      ...form,
      status,
      topic_id: Number(form.topic_id),
      tag_ids: form.tag_ids.map(Number),
    }
    if (isEditing.value) {
      await updateArticle(String(route.params.id), payload)
    } else {
      await createArticle(payload)
    }
    clearLocalDraft()
    toast.success(isEditing.value ? '文章已更新' : '文章已创建')
    await router.push('/admin/articles')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存文章失败'
    toast.error(error.value)
  } finally {
    saving.value = false
  }
}

function scheduleAutosave() {
  if (autosaveTimer) {
    window.clearTimeout(autosaveTimer)
  }
  autosaveStatus.value = '有未保存修改'
  autosaveTimer = window.setTimeout(() => {
    saveLocalDraft()
  }, 800)
}

function saveLocalDraft() {
  const draft: AutoSavedDraft = {
    payload: {
      ...form,
      topic_id: Number(form.topic_id) || 0,
      tag_ids: form.tag_ids.map(Number),
    },
    saved_at: new Date().toISOString(),
  }
  localStorage.setItem(autosaveKey(), JSON.stringify(draft))
  draftAvailable.value = true
  autosaveStatus.value = `已自动保存 ${new Date(draft.saved_at).toLocaleTimeString()}`
}

function loadLocalDraft() {
  const raw = localStorage.getItem(autosaveKey())
  if (!raw) {
    draftAvailable.value = false
    autosaveStatus.value = '自动保存待命'
    return
  }
  draftAvailable.value = true
  try {
    const draft = JSON.parse(raw) as AutoSavedDraft
    autosaveStatus.value = `本地草稿 ${new Date(draft.saved_at).toLocaleString()}`
  } catch {
    clearLocalDraft()
  }
}

function restoreDraft() {
  const raw = localStorage.getItem(autosaveKey())
  if (!raw) {
    return
  }
  const draft = JSON.parse(raw) as AutoSavedDraft
  Object.assign(form, {
    ...draft.payload,
    tag_ids: draft.payload.tag_ids.map(Number),
  })
  autosaveStatus.value = '已恢复本地草稿'
}

function discardDraft() {
  clearLocalDraft()
  autosaveStatus.value = '本地草稿已丢弃'
}

function clearLocalDraft() {
  localStorage.removeItem(autosaveKey())
  draftAvailable.value = false
}

function autosaveKey() {
  const id = isEditing.value ? String(route.params.id) : 'new'
  return `solitude:article-draft:${id}`
}

async function loadVersions() {
  if (!isEditing.value) {
    return
  }
  try {
    versions.value = await getArticleVersions(String(route.params.id))
  } catch {
    versions.value = []
  }
}

function applyVersion(version: ArticleVersionItem) {
  form.title = version.title
  form.summary = version.summary
  form.content_md = version.content_md
  form.status = version.status
}

function formatTime(value: string) {
  return new Date(value).toLocaleString()
}

function onCoverUpload(asset: AssetItem) {
  // local 驱动下 asset.url 已是 Nginx 可解析的 /uploads/... 相对路径；
  // s3 驱动下为 STORAGE_S3_PUBLIC_URL/<key>。两种形态都直接写入即可。
  form.cover_url = asset.url
}

function toSlug(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-|-$/g, '')
}

function safeTagColor(value: string | null | undefined) {
  const color = value?.trim() ?? ''
  return /^#(?:[\da-f]{3}|[\da-f]{6})$/i.test(color) ? color : 'var(--accent)'
}
</script>
