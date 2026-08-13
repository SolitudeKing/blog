<template>
  <section class="admin-page" :aria-busy="loading || uploading || saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Media</p>
        <div role="heading" aria-level="1">媒体库</div>
      </div>
      <BaseButton variant="secondary" :loading="loading" @click="loadAssets()">刷新</BaseButton>
    </header>

    <div class="media-layout">
      <aside class="media-panel" :aria-busy="uploading || saving">
        <div role="heading" aria-level="2">上传资源</div>
        <label class="mist-field">
          <span class="mist-field__label">文件</span>
          <input class="mist-input" type="file" accept="image/*" @change="selectFile" />
        </label>
        <BaseInput v-model="uploadName" label="显示名称" />
        <BaseButton :loading="uploading" :disabled="!selectedFile" @click="upload">上传</BaseButton>

        <div v-if="editingAsset" class="media-edit">
          <div role="heading" aria-level="2">资源信息</div>
          <BaseInput v-model="editForm.display_name" label="显示名称" />
          <BaseInput v-model="editForm.alt_text" label="Alt 文本" />
          <div class="media-actions">
            <BaseButton :loading="saving" @click="saveAsset">保存</BaseButton>
            <BaseButton variant="secondary" @click="cancelEdit">取消</BaseButton>
          </div>
        </div>

        <p v-if="error" class="admin-page__error" role="alert" aria-live="assertive">{{ error }}</p>
      </aside>

      <section class="media-main" :aria-busy="loading">
        <div class="admin-toolbar media-toolbar" role="search" aria-label="筛选媒体资源">
          <input
            v-model="keyword"
            class="mist-input"
            type="search"
            aria-label="搜索文件名或替代文本"
            placeholder="搜索文件名或 Alt"
            @keyup.enter="loadAssets()"
          />
          <BaseSelect
            v-model="mime"
            :options="mimeOptions"
            label="资源类型"
            @change="loadAssets()"
          />
          <BaseButton variant="secondary" :loading="loading" @click="loadAssets()">筛选</BaseButton>
        </div>

        <p v-if="loading && !assets.length" class="admin-page__status" role="status" aria-live="polite">
          正在加载媒体资源…
        </p>

        <div v-if="assets.length" class="media-grid">
          <article v-for="asset in assets" :key="asset.id" class="media-card">
            <div class="media-card__preview">
              <img v-if="asset.mime_type.startsWith('image/')" :src="asset.url" :alt="asset.alt_text || asset.display_name" />
              <span v-else>{{ asset.ext || 'file' }}</span>
            </div>
            <div class="media-card__body">
              <strong>{{ asset.display_name }}</strong>
              <p>{{ asset.mime_type }} · {{ formatSize(asset.size) }}</p>
              <p v-if="asset.width && asset.height">{{ asset.width }} x {{ asset.height }}</p>
            </div>
            <div class="media-card__actions">
              <button class="text-link" type="button" @click="copyURL(asset.url)">复制 URL</button>
              <button class="text-link" type="button" @click="editAsset(asset)">编辑</button>
              <button class="text-link text-link--danger" type="button" @click="removeAsset(asset)">删除</button>
            </div>
          </article>
        </div>
        <BaseEmpty
          v-else-if="!loading"
          title="暂无媒体资源"
          description="从左侧上传第一张图片，文章配图不用愁。"
        >
          <template #icon>
            <SvgIcon class="admin-empty__icon" name="empty-image" />
          </template>
        </BaseEmpty>

        <div v-if="assets.length && page.has_more" class="media-footer">
          <BaseButton variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { deleteAsset, getAssetList, updateAsset, uploadAsset } from '@/api/modules/asset'
import { useToast } from '@/composables/useToast'
import type { CursorPage } from '@/api/types'
import type { AssetItem, AssetUpdatePayload } from '@/types/asset'

const assets = ref<AssetItem[]>([])
const selectedFile = ref<File | null>(null)
const uploadName = ref('')
const keyword = ref('')
const mime = ref<string>('')
const loading = ref(false)
const loadingMore = ref(false)
const uploading = ref(false)
const saving = ref(false)
const error = ref('')
const editingAsset = ref<AssetItem | null>(null)
const toast = useToast()

const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 40,
  has_more: false,
})

const editForm = reactive<AssetUpdatePayload>({
  display_name: '',
  alt_text: '',
})

const mimeOptions = [
  { label: '全部类型', value: '' },
  { label: '图片', value: 'image/' },
  { label: 'SVG', value: 'image/svg+xml' },
]

onMounted(() => {
  loadAssets()
})

async function loadAssets(cursor = '') {
  loading.value = true
  error.value = ''
  try {
    const result = await getAssetList({
      cursor: cursor || undefined,
      keyword: keyword.value || undefined,
      mime: mime.value || undefined,
      limit: page.limit,
    })
    assets.value = cursor ? [...assets.value, ...result.data] : result.data
    Object.assign(page, result.page)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载媒体资源失败'
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
    await loadAssets(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}

function selectFile(event: Event) {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
  if (selectedFile.value && !uploadName.value) {
    uploadName.value = selectedFile.value.name
  }
}

async function upload() {
  if (!selectedFile.value) {
    return
  }
  uploading.value = true
  error.value = ''
  try {
    await uploadAsset(selectedFile.value, uploadName.value)
    selectedFile.value = null
    uploadName.value = ''
    toast.success('上传完成')
    await loadAssets()
  } catch (err) {
    const message = err instanceof Error ? err.message : '上传媒体资源失败'
    error.value = message
    toast.error(message)
  } finally {
    uploading.value = false
  }
}

function editAsset(asset: AssetItem) {
  editingAsset.value = asset
  editForm.display_name = asset.display_name
  editForm.alt_text = asset.alt_text
}

async function saveAsset() {
  if (!editingAsset.value) {
    return
  }
  saving.value = true
  error.value = ''
  try {
    const saved = await updateAsset(editingAsset.value.id, editForm)
    const index = assets.value.findIndex((asset) => asset.id === saved.id)
    if (index >= 0) {
      assets.value[index] = saved
    }
    editingAsset.value = null
    toast.success('资源信息已保存')
  } catch (err) {
    const message = err instanceof Error ? err.message : '保存资源信息失败'
    error.value = message
    toast.error(message)
  } finally {
    saving.value = false
  }
}

function cancelEdit() {
  editingAsset.value = null
}

async function removeAsset(asset: AssetItem) {
  if (!window.confirm('确认删除这个媒体资源？')) {
    return
  }
  try {
    await deleteAsset(asset.id)
    assets.value = assets.value.filter((item) => item.id !== asset.id)
    toast.success('资源已删除')
  } catch (err) {
    const message = err instanceof Error ? err.message : '删除媒体资源失败'
    error.value = message
    toast.error(message)
  }
}

async function copyURL(url: string) {
  try {
    await navigator.clipboard.writeText(new URL(url, window.location.origin).toString())
    toast.success('URL 已复制')
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

function formatSize(size: number) {
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}
</script>
