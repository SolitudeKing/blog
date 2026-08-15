<template>
  <section class="admin-page" :aria-busy="loading || savingTopic || savingTag">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Taxonomy</p>
        <div role="heading" aria-level="1">专题与标签</div>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadAll">刷新</BaseButton>
    </header>

    <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
    <p v-else-if="loading && !topics.length && !tags.length" class="admin-page__status" role="status" aria-live="polite">
      正在加载专题与标签…
    </p>

    <div class="taxonomy-tabs" role="tablist" aria-label="专题与标签切换">
      <button
        type="button"
        class="taxonomy-tab"
        role="tab"
        :aria-selected="activeTab === 'topic'"
        :tabindex="activeTab === 'topic' ? 0 : -1"
        @click="activeTab = 'topic'"
        @keydown.left.prevent="activeTab = 'tag'"
        @keydown.right.prevent="activeTab = 'tag'"
      >
        专题 ({{ topics.length }})
      </button>
      <button
        type="button"
        class="taxonomy-tab"
        role="tab"
        :aria-selected="activeTab === 'tag'"
        :tabindex="activeTab === 'tag' ? 0 : -1"
        @click="activeTab = 'tag'"
        @keydown.left.prevent="activeTab = 'topic'"
        @keydown.right.prevent="activeTab = 'topic'"
      >
        标签 ({{ tags.length }})
      </button>
    </div>

    <section v-show="activeTab === 'topic'" class="taxonomy-panel" role="tabpanel" aria-label="专题">
      <header class="taxonomy-panel__header">
        <div role="heading" aria-level="2">专题</div>
        <BaseButton ref="topicTriggerRef" size="sm" @click="openTopicDialog()">新增专题</BaseButton>
      </header>

      <div class="taxonomy-list">
        <div v-for="topic in topics" :key="topic.id" class="taxonomy-item">
          <div>
            <strong>{{ topic.name }}</strong>
            <p>{{ topic.label }} · {{ topic.slug }}<template v-if="topic.article_count !== undefined"> · 已发布 {{ topic.article_count }} 篇</template></p>
          </div>
          <div class="taxonomy-item__actions">
            <button class="text-link" type="button" @click="openTopicDialog(topic)">编辑</button>
            <button class="text-link text-link--danger" type="button" @click="removeTopic(topic.id)">
              删除
            </button>
          </div>
        </div>
        <p v-if="!topics.length && !loading" class="admin-page__status" role="status">
          还没有任何专题，点击右上角“新增专题”开始整理。
        </p>
      </div>
    </section>

    <section v-show="activeTab === 'tag'" class="taxonomy-panel" role="tabpanel" aria-label="标签">
      <header class="taxonomy-panel__header">
        <div role="heading" aria-level="2">标签</div>
        <BaseButton ref="tagTriggerRef" size="sm" @click="openTagDialog()">新增标签</BaseButton>
      </header>

      <div class="taxonomy-list">
        <div v-for="tag in tags" :key="tag.id" class="taxonomy-item">
          <div>
            <strong>
              <span class="tag-dot" :style="{ background: safeTagColor(tag.color) }" />
              {{ tag.name }}
            </strong>
            <p>{{ tag.slug }}</p>
          </div>
          <div class="taxonomy-item__actions">
            <button class="text-link" type="button" @click="openTagDialog(tag)">编辑</button>
            <button class="text-link text-link--danger" type="button" @click="removeTag(tag.id)">删除</button>
          </div>
        </div>
        <p v-if="!tags.length && !loading" class="admin-page__status" role="status">
          还没有任何标签，点击右上角“新增标签”开始整理。
        </p>
      </div>
    </section>

    <Teleport to="body">
      <Transition name="taxonomy-dialog">
        <div
          v-if="activeDialog"
          class="taxonomy-dialog"
          role="dialog"
          aria-modal="true"
          aria-labelledby="taxonomy-dialog-title"
          @click.self="closeDialog()"
          @keydown="handleDialogKeydown"
        >
          <section ref="dialogPanelRef" class="taxonomy-dialog__panel" tabindex="-1">
            <header class="taxonomy-dialog__header">
              <div>
                <p>{{ activeDialog === 'topic' ? 'Topic' : 'Tag' }}</p>
                <div id="taxonomy-dialog-title" role="heading" aria-level="2">
                  {{ activeDialog === 'topic' ? (editingTopicId ? '编辑专题' : '新增专题') : (editingTagId ? '编辑标签' : '新增标签') }}
                </div>
              </div>
              <button
                ref="dialogCloseButtonRef"
                class="taxonomy-dialog__close"
                type="button"
                aria-label="关闭弹窗"
                :disabled="savingTopic || savingTag"
                @click="closeDialog()"
              >
                <SvgIcon name="close" />
              </button>
            </header>

            <p v-if="formError" class="taxonomy-dialog__error" role="alert">{{ formError }}</p>

            <form v-if="activeDialog === 'topic'" class="taxonomy-form" :aria-busy="savingTopic" @submit.prevent="saveTopic">
              <BaseInput v-model="topicForm.name" label="专题名称" required />
              <BaseInput v-model="topicForm.label" label="Label" hint="用于文章卡片等紧凑位置的短标签" required />
              <BaseInput v-model="topicForm.slug" label="Slug" required />
              <div class="taxonomy-form__cover">
                <div class="taxonomy-form__cover-field">
                  <BaseInput
                    v-model="topicForm.cover_url"
                    type="text"
                    inputmode="url"
                    label="封面 URL"
                    hint="可填写 /uploads/... 相对路径或 https://... 绝对地址；点击右侧上传按钮直接上传。"
                    :maxlength="500"
                  />
                  <UploadButton label="上传封面" @upload="onTopicCoverUpload" />
                </div>
                <div v-if="topicForm.cover_url" class="taxonomy-form__cover-preview">
                  <img :src="topicForm.cover_url" :alt="topicForm.label || topicForm.name || 'topic cover'" />
                  <button class="text-link text-link--danger" type="button" @click="topicForm.cover_url = ''">
                    清除封面
                  </button>
                </div>
              </div>
              <BaseInput v-model="topicSortOrder" label="排序" type="number" />
              <BaseTextarea v-model="topicForm.description" label="描述" :rows="3" />
              <div class="taxonomy-form__actions">
                <BaseButton type="submit" :loading="savingTopic">{{ editingTopicId ? '保存专题' : '新增专题' }}</BaseButton>
                <BaseButton type="button" variant="secondary" :disabled="savingTopic" @click="closeDialog()">取消</BaseButton>
              </div>
            </form>

            <form v-else class="taxonomy-form" :aria-busy="savingTag" @submit.prevent="saveTag">
              <BaseInput v-model="tagForm.name" label="名称" required />
              <BaseInput v-model="tagForm.slug" label="Slug" required />
              <BaseInput v-model="tagForm.color" label="颜色" hint="使用 #RGB 或 #RRGGBB 格式，例如 #5f8d62" :error="tagColorError" required />
              <BaseTextarea v-model="tagForm.description" label="描述" :rows="3" />
              <div class="taxonomy-form__actions">
                <BaseButton type="submit" :loading="savingTag">{{ editingTagId ? '保存标签' : '新增标签' }}</BaseButton>
                <BaseButton type="button" variant="secondary" :disabled="savingTag" @click="closeDialog()">取消</BaseButton>
              </div>
            </form>
          </section>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import UploadButton from '@/components/media/UploadButton.vue'
import {
  createTag,
  createTopic,
  deleteTag,
  deleteTopic,
  getTagList,
  getTopicList,
  updateTag,
  updateTopic,
} from '@/api/modules/taxonomy'
import type { TagItem, TagPayload, TopicItem, TopicPayload } from '@/types/taxonomy'
import type { AssetItem } from '@/types/asset'

const topics = ref<TopicItem[]>([])
const tags = ref<TagItem[]>([])
const loading = ref(false)
const savingTopic = ref(false)
const savingTag = ref(false)
const error = ref('')
const formError = ref('')
const editingTopicId = ref<number | null>(null)
const editingTagId = ref<number | null>(null)
const activeDialog = ref<'topic' | 'tag' | null>(null)
const activeTab = ref<'topic' | 'tag'>('topic')
const topicTriggerRef = ref<InstanceType<typeof BaseButton> | null>(null)
const tagTriggerRef = ref<InstanceType<typeof BaseButton> | null>(null)
const dialogCloseButtonRef = ref<HTMLButtonElement | null>(null)
const dialogPanelRef = ref<HTMLElement | null>(null)

let lockedScrollY = 0
let previousBodyStyle: Partial<Record<'position' | 'top' | 'left' | 'right' | 'width' | 'overflow', string>> = {}

const topicForm = reactive<TopicPayload>({
  name: '',
  label: '',
  slug: '',
  description: '',
  cover_url: '',
  sort_order: 0,
})

const tagForm = reactive<TagPayload>({
  name: '',
  slug: '',
  description: '',
  color: '#5f8d62',
})

const topicSortOrder = computed({
  get: () => String(topicForm.sort_order),
  set: (value: string) => {
    topicForm.sort_order = Number(value) || 0
  },
})

const tagColorError = computed(() => {
  const color = tagForm.color.trim()
  if (!color) {
    return '请输入标签颜色'
  }
  if (!isHexColor(color)) {
    return '颜色格式应为 #RGB 或 #RRGGBB'
  }
  return ''
})

onMounted(() => {
  loadAll()
})

watch(
  () => topicForm.name,
  (name) => {
    if (editingTopicId.value || topicForm.slug) {
      return
    }
    topicForm.slug = toSlug(name)
  },
)

watch(
  () => tagForm.name,
  (name) => {
    if (editingTagId.value || tagForm.slug) {
      return
    }
    tagForm.slug = toSlug(name)
  },
)

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    const [topicItems, tagItems] = await Promise.all([getTopicList(), getTagList()])
    topics.value = topicItems
    tags.value = tagItems
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载专题与标签失败'
  } finally {
    loading.value = false
  }
}

function onTopicCoverUpload(asset: AssetItem) {
  // local 驱动下 asset.url 是 Nginx 可解析的 /uploads/... 相对路径；
  // s3 驱动下为 STORAGE_S3_PUBLIC_URL/<key>。两种形态都直接写入即可。
  topicForm.cover_url = asset.url
}

async function saveTopic() {
  savingTopic.value = true
  formError.value = ''
  try {
    const payload: TopicPayload = {
      ...topicForm,
      name: topicForm.name.trim(),
      label: topicForm.label.trim(),
      slug: topicForm.slug.trim(),
      description: topicForm.description.trim(),
      cover_url: topicForm.cover_url.trim(),
    }
    if (editingTopicId.value) {
      await updateTopic(editingTopicId.value, payload)
    } else {
      await createTopic(payload)
    }
    await loadAll()
    await closeDialog(true, true)
  } catch (err) {
    formError.value = err instanceof Error ? err.message : '保存专题失败'
  } finally {
    savingTopic.value = false
  }
}

async function removeTopic(id: number) {
  if (!window.confirm('确认删除这个专题？已被文章使用的专题不能删除。')) {
    return
  }
  try {
    await deleteTopic(id)
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除专题失败'
  }
}

function resetTopicForm() {
  editingTopicId.value = null
  topicForm.name = ''
  topicForm.label = ''
  topicForm.slug = ''
  topicForm.description = ''
  topicForm.cover_url = ''
  topicForm.sort_order = 0
}

async function saveTag() {
  formError.value = ''
  if (tagColorError.value) {
    formError.value = tagColorError.value
    return
  }

  savingTag.value = true
  try {
    const payload: TagPayload = {
      ...tagForm,
      color: tagForm.color.trim(),
    }
    if (editingTagId.value) {
      await updateTag(editingTagId.value, payload)
    } else {
      await createTag(payload)
    }
    await loadAll()
    await closeDialog(true, true)
  } catch (err) {
    formError.value = err instanceof Error ? err.message : '保存标签失败'
  } finally {
    savingTag.value = false
  }
}

async function removeTag(id: number) {
  if (!window.confirm('确认删除这个标签？已被文章使用的标签不能删除。')) {
    return
  }
  try {
    await deleteTag(id)
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除标签失败'
  }
}

function resetTagForm() {
  editingTagId.value = null
  tagForm.name = ''
  tagForm.slug = ''
  tagForm.description = ''
  tagForm.color = '#5f8d62'
}

function lockPage() {
  const { body } = document
  lockedScrollY = window.scrollY
  previousBodyStyle = {
    position: body.style.position,
    top: body.style.top,
    left: body.style.left,
    right: body.style.right,
    width: body.style.width,
    overflow: body.style.overflow,
  }
  body.style.position = 'fixed'
  body.style.top = `-${lockedScrollY}px`
  body.style.left = '0'
  body.style.right = '0'
  body.style.width = '100%'
  body.style.overflow = 'hidden'
  const app = document.getElementById('app')
  if (app) {
    app.inert = true
  }
}

function unlockPage() {
  const { body } = document
  body.style.position = previousBodyStyle.position ?? ''
  body.style.top = previousBodyStyle.top ?? ''
  body.style.left = previousBodyStyle.left ?? ''
  body.style.right = previousBodyStyle.right ?? ''
  body.style.width = previousBodyStyle.width ?? ''
  body.style.overflow = previousBodyStyle.overflow ?? ''
  const app = document.getElementById('app')
  if (app) {
    app.inert = false
  }
  window.scrollTo(0, lockedScrollY)
}

async function openDialog(kind: 'topic' | 'tag') {
  formError.value = ''
  activeDialog.value = kind
  lockPage()
  await nextTick()
  dialogCloseButtonRef.value?.focus()
}

async function closeDialog(restoreFocus = true, force = false) {
  if (!activeDialog.value || (!force && (savingTopic.value || savingTag.value))) {
    return
  }
  const dialog = activeDialog.value
  activeDialog.value = null
  resetTopicForm()
  resetTagForm()
  formError.value = ''
  unlockPage()
  if (!restoreFocus) {
    return
  }
  await nextTick()
  const trigger = dialog === 'topic' ? topicTriggerRef.value : tagTriggerRef.value
  trigger?.$el.focus({ preventScroll: true })
}

function openTopicDialog(topic?: TopicItem) {
  resetTopicForm()
  if (topic) {
    editingTopicId.value = topic.id
    topicForm.name = topic.name
    topicForm.label = topic.label
    topicForm.slug = topic.slug
    topicForm.description = topic.description
    topicForm.cover_url = topic.cover_url
    topicForm.sort_order = topic.sort_order
  }
  void openDialog('topic')
}

function openTagDialog(tag?: TagItem) {
  resetTagForm()
  if (tag) {
    editingTagId.value = tag.id
    tagForm.name = tag.name
    tagForm.slug = tag.slug
    tagForm.description = tag.description
    tagForm.color = tag.color || '#5f8d62'
  }
  void openDialog('tag')
}

function handleDialogKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    void closeDialog()
    return
  }
  if (event.key !== 'Tab' || !dialogPanelRef.value) {
    return
  }
  const focusable = Array.from(
    dialogPanelRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  ).filter((element) => !element.hasAttribute('hidden'))
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (!first || !last) {
    event.preventDefault()
    dialogPanelRef.value.focus()
    return
  }
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

onBeforeUnmount(() => {
  if (activeDialog.value) {
    activeDialog.value = null
    unlockPage()
  }
})

function toSlug(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-|-$/g, '')
}

function isHexColor(value: string) {
  return /^#(?:[\da-f]{3}|[\da-f]{6})$/i.test(value)
}

function safeTagColor(value: string | null | undefined) {
  const color = value?.trim() ?? ''
  return isHexColor(color) ? color : 'var(--accent)'
}
</script>
