<template>
  <section class="admin-page" :aria-busy="loading || savingTopic || savingTag">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Taxonomy</p>
        <h1>专题与标签</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadAll">刷新</BaseButton>
    </header>

    <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
    <p v-else-if="loading && !topics.length && !tags.length" class="admin-page__status" role="status" aria-live="polite">
      正在加载专题与标签…
    </p>

    <div class="taxonomy-grid">
      <section class="taxonomy-panel">
        <h2>专题</h2>
        <form class="taxonomy-form" :aria-busy="savingTopic" @submit.prevent="saveTopic">
          <BaseInput v-model="topicForm.name" label="专题名称" required />
          <BaseInput
            v-model="topicForm.label"
            label="Label"
            hint="用于文章卡片等紧凑位置的短标签"
            required
          />
          <BaseInput v-model="topicForm.slug" label="Slug" required />
          <BaseInput
            v-model="topicForm.cover_url"
            label="封面 URL"
            hint="可填写 /uploads/... 相对路径或 https://... 绝对地址"
          />
          <BaseInput v-model="topicSortOrder" label="排序" type="number" />
          <BaseTextarea v-model="topicForm.description" label="描述" :rows="3" />
          <div class="taxonomy-form__actions">
            <BaseButton type="submit" :loading="savingTopic">{{ editingTopicId ? '保存专题' : '新增专题' }}</BaseButton>
            <BaseButton v-if="editingTopicId" type="button" variant="secondary" @click="resetTopicForm">
              取消
            </BaseButton>
          </div>
        </form>

        <div class="taxonomy-list">
          <div v-for="topic in topics" :key="topic.id" class="taxonomy-item">
            <div>
              <strong>{{ topic.name }}</strong>
              <p>{{ topic.label }} · {{ topic.slug }}<template v-if="topic.article_count !== undefined"> · 已发布 {{ topic.article_count }} 篇</template></p>
            </div>
            <div class="taxonomy-item__actions">
              <button class="text-link" type="button" @click="editTopic(topic)">编辑</button>
              <button class="text-link text-link--danger" type="button" @click="removeTopic(topic.id)">
                删除
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="taxonomy-panel">
        <h2>标签</h2>
        <form class="taxonomy-form" :aria-busy="savingTag" @submit.prevent="saveTag">
          <BaseInput v-model="tagForm.name" label="名称" />
          <BaseInput v-model="tagForm.slug" label="Slug" />
          <BaseInput
            v-model="tagForm.color"
            label="颜色"
            hint="使用 #RGB 或 #RRGGBB 格式，例如 #5f8d62"
            :error="tagColorError"
          />
          <BaseTextarea v-model="tagForm.description" label="描述" :rows="3" />
          <div class="taxonomy-form__actions">
            <BaseButton type="submit" :loading="savingTag">{{ editingTagId ? '保存标签' : '新增标签' }}</BaseButton>
            <BaseButton v-if="editingTagId" type="button" variant="secondary" @click="resetTagForm">取消</BaseButton>
          </div>
        </form>

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
              <button class="text-link" type="button" @click="editTag(tag)">编辑</button>
              <button class="text-link text-link--danger" type="button" @click="removeTag(tag.id)">删除</button>
            </div>
          </div>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
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

const topics = ref<TopicItem[]>([])
const tags = ref<TagItem[]>([])
const loading = ref(false)
const savingTopic = ref(false)
const savingTag = ref(false)
const error = ref('')
const editingTopicId = ref<number | null>(null)
const editingTagId = ref<number | null>(null)

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

async function saveTopic() {
  savingTopic.value = true
  error.value = ''
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
    resetTopicForm()
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存专题失败'
  } finally {
    savingTopic.value = false
  }
}

function editTopic(topic: TopicItem) {
  editingTopicId.value = topic.id
  topicForm.name = topic.name
  topicForm.label = topic.label
  topicForm.slug = topic.slug
  topicForm.description = topic.description
  topicForm.cover_url = topic.cover_url
  topicForm.sort_order = topic.sort_order
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
  error.value = ''
  if (tagColorError.value) {
    error.value = tagColorError.value
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
    resetTagForm()
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存标签失败'
  } finally {
    savingTag.value = false
  }
}

function editTag(tag: TagItem) {
  editingTagId.value = tag.id
  tagForm.name = tag.name
  tagForm.slug = tag.slug
  tagForm.description = tag.description
  tagForm.color = tag.color || '#5f8d62'
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
