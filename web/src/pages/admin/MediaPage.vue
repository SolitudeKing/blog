<template>
  <section class="admin-page" :aria-busy="loading || uploading || saving">
    <header class="admin-page__header">
      <div>
        <p class="admin-page__eyebrow">Media</p>
        <div role="heading" aria-level="1">媒体库</div>
      </div>
      <div class="admin-page__actions">
        <BaseButton variant="secondary" :loading="loading" @click="loadAssets()">刷新</BaseButton>
        <BaseButton variant="primary" @click="openUploadModal">上传</BaseButton>
      </div>
    </header>

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
      <article
        v-for="asset in assets"
        :key="asset.id"
        class="media-card"
        tabindex="0"
        role="button"
        :aria-label="`查看 ${asset.display_name} 资源详情`"
        @click="openDetailModal(asset)"
        @keydown.enter.prevent="openDetailModal(asset)"
        @keydown.space.prevent="openDetailModal(asset)"
      >
        <div class="media-card__preview">
          <img
            v-if="asset.mime_type.startsWith('image/')"
            :src="asset.url"
            :alt="asset.alt_text || asset.display_name"
          />
          <span v-else class="media-card__file-fallback">{{ asset.ext || 'file' }}</span>
          <div class="media-card__actions" @click.stop>
            <button
              type="button"
              class="media-card__action"
              :aria-label="`复制 ${asset.display_name} 的 URL`"
              @click="copyURL(asset.url)"
            >
              <SvgIcon name="link" :size="16" />
            </button>
            <button
              type="button"
              class="media-card__action"
              :aria-label="`编辑 ${asset.display_name}`"
              @click="openEditModal(asset)"
            >
              <SvgIcon name="settings" :size="16" />
            </button>
            <button
              type="button"
              class="media-card__action media-card__action--danger"
              :aria-label="`删除 ${asset.display_name}`"
              @click="removeAsset(asset)"
            >
              <SvgIcon name="close" :size="16" />
            </button>
          </div>
        </div>
        <div class="media-card__body">
          <strong>{{ asset.display_name }}</strong>
          <p>{{ asset.mime_type }} · {{ formatSize(asset.size) }}</p>
          <p v-if="asset.width && asset.height">{{ asset.width }} × {{ asset.height }}</p>
        </div>
      </article>
    </div>
    <BaseEmpty
      v-else-if="!loading"
      title="暂无媒体资源"
      description="点击右上角“上传”按钮，添加第一张图片。"
    >
      <template #icon>
        <SvgIcon class="admin-empty__icon" name="empty-image" />
      </template>
    </BaseEmpty>

    <div v-if="assets.length && page.has_more" class="media-footer">
      <BaseButton variant="secondary" :loading="loadingMore" @click="loadMore">加载更多</BaseButton>
    </div>

    <!-- 上传 Modal -->
    <BaseModal v-model="showUploadModal" title="上传媒体资源" eyebrow="Upload">
      <div class="media-edit-modal">
        <UploadButton :url="uploadForm.url" @upload="onUploadSuccess" @clear="resetUploadForm" />
        <BaseInput
          v-model="uploadForm.displayName"
          label="显示名称"
          hint="留空时使用文件名"
        />
        <BaseInput
          v-model="uploadForm.altText"
          label="Alt 文本"
          hint="可选；用于图片 alt 描述"
        />
      </div>
      <template #footer>
        <BaseButton variant="secondary" :disabled="uploading" @click="closeUploadModal">
          取消
        </BaseButton>
        <BaseButton
          :loading="uploading"
          :disabled="uploadForm.disabled"
          @click="confirmUpload"
        >
          上传
        </BaseButton>
      </template>
    </BaseModal>

    <!-- 编辑 Modal -->
    <BaseModal v-model="showEditModal" title="编辑资源信息" eyebrow="Edit">
      <div v-if="editingAsset" class="media-edit-modal">
        <div class="media-edit-modal__preview">
          <img
            v-if="editingAsset.mime_type.startsWith('image/')"
            :src="editingAsset.url"
            :alt="editingAsset.alt_text || editingAsset.display_name"
          />
          <span v-else class="media-card__file-fallback">{{ editingAsset.ext || 'file' }}</span>
        </div>
        <BaseInput v-model="editForm.display_name" label="显示名称" />
        <BaseInput v-model="editForm.alt_text" label="Alt 文本" />
        <p v-if="editingAsset.ref_count > 0" class="media-edit-modal__refcount">
          ⚠ 当前被 <strong>{{ editingAsset.ref_count }}</strong> 处引用，删除前请先解除引用。
        </p>
      </div>
      <template #footer>
        <BaseButton variant="secondary" :disabled="saving" @click="closeEditModal">
          取消
        </BaseButton>
        <BaseButton :loading="saving" @click="saveAsset">保存</BaseButton>
      </template>
    </BaseModal>

    <!-- 详情 Modal -->
    <BaseModal v-model="showDetailModal" title="资源详情" eyebrow="Inspect" size="lg">
      <div v-if="detailAsset" class="media-detail-modal">
        <div class="media-detail-modal__image">
          <img
            v-if="detailAsset.mime_type.startsWith('image/')"
            :src="detailAsset.url"
            :alt="detailAsset.alt_text || detailAsset.display_name"
          />
          <span v-else class="media-card__file-fallback">{{ detailAsset.ext || 'file' }}</span>
        </div>
        <dl class="media-detail-modal__meta">
          <div><dt>文件名</dt><dd>{{ detailAsset.display_name }}</dd></div>
          <div><dt>MIME</dt><dd>{{ detailAsset.mime_type }}</dd></div>
          <div v-if="detailAsset.width && detailAsset.height">
            <dt>尺寸</dt>
            <dd>{{ detailAsset.width }} × {{ detailAsset.height }}</dd>
          </div>
          <div><dt>大小</dt><dd>{{ formatSize(detailAsset.size) }}</dd></div>
          <div><dt>引用数</dt><dd>{{ detailAsset.ref_count }}</dd></div>
          <div v-if="detailAsset.sha256">
            <dt>SHA256</dt>
            <dd class="media-detail-modal__hash">{{ detailAsset.sha256 }}</dd>
          </div>
          <div>
            <dt>URL</dt>
            <dd class="media-detail-modal__url">{{ detailAsset.url }}</dd>
          </div>
        </dl>
      </div>
      <template #footer>
        <BaseButton variant="secondary" @click="copyURL(detailAsset?.url ?? '')">复制 URL</BaseButton>
        <BaseButton variant="secondary" @click="openEditFromDetail">编辑</BaseButton>
        <BaseButton variant="ghost" @click="removeAsset(detailAsset)">删除</BaseButton>
        <BaseButton variant="primary" @click="showDetailModal = false">关闭</BaseButton>
      </template>
    </BaseModal>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import BaseEmpty from '@/components/base/BaseEmpty.vue'
import BaseInput from '@/components/base/BaseInput.vue'
import BaseModal from '@/components/base/BaseModal.vue'
import BaseSelect from '@/components/base/BaseSelect.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import UploadButton from '@/components/media/UploadButton.vue'
import { deleteAsset, getAssetList, updateAsset } from '@/api/modules/asset'
import { useToast } from '@/composables/useToast'
import type { CursorPage } from '@/api/types'
import type { AssetItem, AssetUpdatePayload } from '@/types/asset'

const assets = ref<AssetItem[]>([])
const keyword = ref('')
const mime = ref<string>('')
const loading = ref(false)
const loadingMore = ref(false)
const uploading = ref(false)
const saving = ref(false)
const toast = useToast()

const showUploadModal = ref(false)
const showEditModal = ref(false)
const showDetailModal = ref(false)
const editingAsset = ref<AssetItem | null>(null)
const detailAsset = ref<AssetItem | null>(null)

const uploadForm = reactive({
  url: '',
  displayName: '',
  altText: '',
  // 上传弹窗中"上传"按钮总是可用，URL 由 UploadButton emit 触发；
  // 这只是个占位，预留 future 文件模式的扩展入口。
  disabled: false,
})

const editForm = reactive<AssetUpdatePayload>({
  display_name: '',
  alt_text: '',
})

const page = reactive<CursorPage>({
  cursor: '',
  next_cursor: '',
  limit: 40,
  has_more: false,
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
    toast.error(err instanceof Error ? err.message : '加载媒体资源失败')
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!page.next_cursor) return
  loadingMore.value = true
  try {
    await loadAssets(page.next_cursor)
  } finally {
    loadingMore.value = false
  }
}

function openUploadModal() {
  resetUploadForm()
  showUploadModal.value = true
}

function closeUploadModal() {
  if (uploading.value) return
  showUploadModal.value = false
  resetUploadForm()
}

function resetUploadForm() {
  uploadForm.url = ''
  uploadForm.displayName = ''
  uploadForm.altText = ''
}

function onUploadSuccess(asset: { url: string }) {
  uploadForm.url = asset.url
  if (!uploadForm.displayName) {
    uploadForm.displayName = asset.url.split('/').pop() ?? ''
  }
}

async function confirmUpload() {
  if (!uploadForm.url) {
    toast.error('请先上传文件')
    return
  }
  // 当前直接采用 URL 模式（上传已经完成）；保留 confirmUpload 入口用于日后的"先选文件再确认"扩展
  uploading.value = true
  try {
    toast.success('上传完成')
    showUploadModal.value = false
    resetUploadForm()
    await loadAssets()
  } catch (err) {
    toast.error(err instanceof Error ? err.message : '上传媒体资源失败')
  } finally {
    uploading.value = false
  }
}

function openEditModal(asset: AssetItem) {
  editingAsset.value = asset
  editForm.display_name = asset.display_name
  editForm.alt_text = asset.alt_text
  showEditModal.value = true
}

function closeEditModal() {
  if (saving.value) return
  showEditModal.value = false
  editingAsset.value = null
}

async function saveAsset() {
  if (!editingAsset.value) return
  saving.value = true
  try {
    const saved = await updateAsset(editingAsset.value.id, editForm)
    const index = assets.value.findIndex((asset) => asset.id === saved.id)
    if (index >= 0) {
      assets.value[index] = saved
    }
    // 同步刷新详情 Modal 数据
    if (detailAsset.value?.id === saved.id) {
      detailAsset.value = saved
    }
    showEditModal.value = false
    editingAsset.value = null
    toast.success('资源信息已保存')
  } catch (err) {
    toast.error(err instanceof Error ? err.message : '保存资源信息失败')
  } finally {
    saving.value = false
  }
}

function openDetailModal(asset: AssetItem) {
  detailAsset.value = asset
  showDetailModal.value = true
}

function openEditFromDetail() {
  if (!detailAsset.value) return
  const target = detailAsset.value
  showDetailModal.value = false
  openEditModal(target)
}

async function removeAsset(asset: AssetItem | null) {
  if (!asset) return
  if (!window.confirm(`确认删除"${asset.display_name}"？`)) return
  try {
    await deleteAsset(asset.id)
    assets.value = assets.value.filter((item) => item.id !== asset.id)
    if (editingAsset.value?.id === asset.id) {
      showEditModal.value = false
      editingAsset.value = null
    }
    if (detailAsset.value?.id === asset.id) {
      showDetailModal.value = false
      detailAsset.value = null
    }
    toast.success('资源已删除')
  } catch (err) {
    toast.error(err instanceof Error ? err.message : '删除媒体资源失败')
  }
}

async function copyURL(url: string) {
  if (!url) return
  try {
    await navigator.clipboard.writeText(new URL(url, window.location.origin).toString())
    toast.success('URL 已复制')
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

function formatSize(size: number | undefined) {
  if (!size) return ''
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}
</script>
