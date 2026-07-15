<template>
  <section class="admin-page" :aria-busy="loading || savingCategory || savingTag">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Taxonomy</p>
        <h1>分类标签</h1>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadAll">刷新</BaseButton>
    </header>

    <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
    <p v-else-if="loading && !categories.length && !tags.length" class="admin-page__status" role="status" aria-live="polite">
      正在加载分类与标签…
    </p>

    <div class="taxonomy-grid">
      <section class="taxonomy-panel">
        <h2>分类</h2>
        <form class="taxonomy-form" :aria-busy="savingCategory" @submit.prevent="saveCategory">
          <BaseInput v-model="categoryForm.name" label="名称" />
          <BaseInput v-model="categoryForm.slug" label="Slug" />
          <BaseInput v-model="categorySortOrder" label="排序" type="number" />
          <BaseTextarea v-model="categoryForm.description" label="描述" :rows="3" />
          <div class="taxonomy-form__actions">
            <BaseButton type="submit" :loading="savingCategory">{{ editingCategoryId ? '保存分类' : '新增分类' }}</BaseButton>
            <BaseButton v-if="editingCategoryId" type="button" variant="secondary" @click="resetCategoryForm">
              取消
            </BaseButton>
          </div>
        </form>

        <div class="taxonomy-list">
          <div v-for="category in categories" :key="category.id" class="taxonomy-item">
            <div>
              <strong>{{ category.name }}</strong>
              <p>{{ category.slug }}</p>
            </div>
            <div class="taxonomy-item__actions">
              <button class="text-link" type="button" @click="editCategory(category)">编辑</button>
              <button class="text-link text-link--danger" type="button" @click="removeCategory(category.id)">
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
  createCategory,
  createTag,
  deleteCategory,
  deleteTag,
  getCategoryList,
  getTagList,
  updateCategory,
  updateTag,
} from '@/api/modules/taxonomy'
import type { CategoryItem, CategoryPayload, TagItem, TagPayload } from '@/types/taxonomy'

const categories = ref<CategoryItem[]>([])
const tags = ref<TagItem[]>([])
const loading = ref(false)
const savingCategory = ref(false)
const savingTag = ref(false)
const error = ref('')
const editingCategoryId = ref<number | null>(null)
const editingTagId = ref<number | null>(null)

const categoryForm = reactive<CategoryPayload>({
  name: '',
  slug: '',
  description: '',
  sort_order: 0,
})

const tagForm = reactive<TagPayload>({
  name: '',
  slug: '',
  description: '',
  color: '#5f8d62',
})

const categorySortOrder = computed({
  get: () => String(categoryForm.sort_order),
  set: (value: string) => {
    categoryForm.sort_order = Number(value) || 0
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
  () => categoryForm.name,
  (name) => {
    if (editingCategoryId.value || categoryForm.slug) {
      return
    }
    categoryForm.slug = toSlug(name)
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
    const [categoryItems, tagItems] = await Promise.all([getCategoryList(), getTagList()])
    categories.value = categoryItems
    tags.value = tagItems
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载分类标签失败'
  } finally {
    loading.value = false
  }
}

async function saveCategory() {
  savingCategory.value = true
  error.value = ''
  try {
    if (editingCategoryId.value) {
      await updateCategory(editingCategoryId.value, categoryForm)
    } else {
      await createCategory(categoryForm)
    }
    resetCategoryForm()
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存分类失败'
  } finally {
    savingCategory.value = false
  }
}

function editCategory(category: CategoryItem) {
  editingCategoryId.value = category.id
  categoryForm.name = category.name
  categoryForm.slug = category.slug
  categoryForm.description = category.description
  categoryForm.sort_order = category.sort_order
}

async function removeCategory(id: number) {
  if (!window.confirm('确认删除这个分类？已被文章使用的分类不能删除。')) {
    return
  }
  try {
    await deleteCategory(id)
    await loadAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除分类失败'
  }
}

function resetCategoryForm() {
  editingCategoryId.value = null
  categoryForm.name = ''
  categoryForm.slug = ''
  categoryForm.description = ''
  categoryForm.sort_order = 0
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
