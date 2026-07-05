<template>
  <section class="admin-page article-editor">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Editor</p>
        <h1>{{ isEditing ? '编辑文章' : '新建文章' }}</h1>
      </div>
      <RouterLink class="cui-button cui-button--secondary" to="/admin/articles">返回列表</RouterLink>
    </header>

    <form class="editor-form" @submit.prevent="save">
      <div class="editor-form__main">
        <BaseInput v-model="form.title" label="标题" />
        <BaseInput v-model="form.slug" label="Slug" />
        <BaseTextarea v-model="form.summary" label="摘要" :rows="4" />
        <BaseTextarea v-model="form.content_md" label="Markdown 正文" :rows="18" />
      </div>

      <aside class="editor-panel">
        <label class="cui-field">
          <span class="cui-field__label">状态</span>
          <select v-model="form.status" class="cui-input">
            <option value="draft">草稿</option>
            <option value="published">发布</option>
            <option value="private">私有</option>
            <option value="archived">归档</option>
          </select>
        </label>

        <label class="cui-field">
          <span class="cui-field__label">分类</span>
          <select v-model.number="form.category_id" class="cui-input">
            <option :value="0">默认分类</option>
            <option v-for="category in categories" :key="category.id" :value="category.id">
              {{ category.name }}
            </option>
          </select>
        </label>

        <div class="cui-field">
          <span class="cui-field__label">标签</span>
          <div class="tag-checks">
            <label v-for="tag in tags" :key="tag.id" class="tag-check">
              <input v-model="form.tag_ids" type="checkbox" :value="tag.id" />
              <span class="tag-dot" :style="{ background: tag.color || 'var(--accent)' }" />
              {{ tag.name }}
            </label>
          </div>
        </div>

        <div class="editor-panel__actions">
          <BaseButton type="submit" :loading="saving">{{ isEditing ? '保存修改' : '创建文章' }}</BaseButton>
          <BaseButton type="button" variant="secondary" :disabled="saving" @click="publish">保存并发布</BaseButton>
        </div>
        <p v-if="error" class="admin-page__error">{{ error }}</p>

        <div v-if="isEditing" class="editor-versions">
          <div class="dashboard-panel__header">
            <h2>版本记录</h2>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseTextarea from '@/components/base/BaseTextarea.vue'
import { createArticle, getArticleVersions, getManagedArticleInfo, updateArticle } from '@/api/modules/article'
import { getCategoryList, getTagList } from '@/api/modules/taxonomy'
import type { ArticleSavePayload, ArticleVersionItem } from '@/types/article'
import type { CategoryItem, TagItem } from '@/types/taxonomy'

const route = useRoute()
const router = useRouter()
const saving = ref(false)
const error = ref('')
const isEditing = computed(() => route.name === 'admin-article-edit')
const categories = ref<CategoryItem[]>([])
const tags = ref<TagItem[]>([])
const versions = ref<ArticleVersionItem[]>([])

const form = reactive<ArticleSavePayload>({
  title: '',
  slug: '',
  summary: '',
  content_md: '',
  category_id: 0,
  tag_ids: [],
  status: 'draft',
})

onMounted(async () => {
  await loadTaxonomy()
  if (!isEditing.value) {
    return
  }
  try {
    const article = await getManagedArticleInfo(String(route.params.id))
    form.title = article.title
    form.slug = article.slug
    form.summary = article.summary
    form.content_md = article.content_md
    form.status = article.status
    form.category_id = article.category_id
    form.tag_ids = article.tag_ids
    await loadVersions()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载文章失败'
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

async function loadTaxonomy() {
  try {
    const [categoryItems, tagItems] = await Promise.all([getCategoryList(), getTagList()])
    categories.value = categoryItems
    tags.value = tagItems
    if (!isEditing.value && form.category_id === 0 && categoryItems.length > 0) {
      form.category_id = categoryItems[0].id
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载分类标签失败'
  }
}

async function save() {
  await submit(form.status)
}

async function publish() {
  await submit('published')
}

async function submit(status: ArticleSavePayload['status']) {
  saving.value = true
  error.value = ''
  try {
    const payload: ArticleSavePayload = {
      ...form,
      status,
      category_id: Number(form.category_id) || 0,
      tag_ids: form.tag_ids.map(Number),
    }
    if (isEditing.value) {
      await updateArticle(String(route.params.id), payload)
    } else {
      await createArticle(payload)
    }
    await router.push('/admin/articles')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存文章失败'
  } finally {
    saving.value = false
  }
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

function toSlug(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-|-$/g, '')
}
</script>
