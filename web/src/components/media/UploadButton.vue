<template>
  <div
    class="upload-button"
    :class="{
      'is-selected': hasSelection,
      'is-error': !!errorMessage,
      'is-dragover': dragOver,
      'is-disabled': disabled,
    }"
  >
    <input
      ref="fileInput"
      type="file"
      :accept="accept"
      class="upload-button__input"
      :disabled="disabled || uploading"
      @change="onFileChange"
      @click.stop
    />

    <div v-if="!hasSelection" class="upload-button__zone" @click="openPicker">
      <div class="upload-button__zone-icon" aria-hidden="true">
        <SvgIcon name="arrow-up-right" :size="20" />
      </div>
      <div class="upload-button__zone-copy">
        <strong class="upload-button__zone-title">{{ label }}</strong>
        <span class="upload-button__zone-hint">{{ hintText }}</span>
      </div>
    </div>

    <div v-else class="upload-button__preview">
      <div class="upload-button__preview-thumb" aria-hidden="true">
        <img v-if="thumbnailUrl" :src="thumbnailUrl" alt="" />
        <SvgIcon v-else name="document" :size="20" />
      </div>
      <div class="upload-button__preview-meta">
        <strong class="upload-button__preview-name">{{ displayName }}</strong>
        <span class="upload-button__preview-size">{{ sizeText }}</span>
      </div>
      <div class="upload-button__preview-actions">
        <BaseButton
          type="button"
          variant="secondary"
          size="sm"
          :disabled="uploading"
          @click="openPicker"
        >
          重新选择
        </BaseButton>
        <button
          type="button"
          class="text-link text-link--danger"
          :disabled="uploading"
          @click="clearSelection"
        >
          清除
        </button>
      </div>
    </div>

    <p v-if="errorMessage" class="upload-button__error" role="alert" aria-live="polite">
      {{ errorMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { uploadAsset } from '@/api/modules/asset'
import { useToast } from '@/composables/useToast'
import type { AssetItem } from '@/types/asset'

// 与后端 server/internal/service/asset_service.go::maxAssetUploadSize 对齐。
const MAX_FILE_SIZE_BYTES = 10 * 1024 * 1024

const props = withDefaults(
  defineProps<{
    label?: string
    hint?: string
    accept?: string
    disabled?: boolean
    /** 外部同步 URL；非空时组件进入"已选择"预览态。 */
    url?: string
  }>(),
  {
    label: '上传图片',
    hint: '支持 JPG / PNG / GIF / WebP / SVG，单个文件 ≤ 10MB',
    accept: 'image/png,image/jpeg,image/gif,image/webp,image/svg+xml',
    disabled: false,
    url: '',
  },
)

const emit = defineEmits<{
  upload: [asset: AssetItem]
  error: [message: string]
  clear: []
  'update:url': [value: string]
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const errorMessage = ref('')
const localFileName = ref('')
const localFileSize = ref(0)
const localPreviewUrl = ref('')
const dragOver = ref(false)
const toast = useToast()

const hintText = computed(() => props.hint)
const hasSelection = computed(() => Boolean(props.url || localFileName.value))
const displayName = computed(() => localFileName.value || deriveNameFromUrl(props.url))
const sizeText = computed(() => formatSize(localFileSize.value))
const thumbnailUrl = computed(() => localPreviewUrl.value || props.url)

function deriveNameFromUrl(url: string): string {
  if (!url) return ''
  try {
    const path = new URL(url, 'http://placeholder').pathname
    const segments = path.split('/').filter(Boolean)
    return segments[segments.length - 1] ?? url
  } catch {
    return url
  }
}

function formatSize(bytes: number): string {
  if (bytes <= 0) return ''
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function openPicker() {
  if (props.disabled || uploading.value) return
  errorMessage.value = ''
  fileInput.value?.click()
}

function resetInput() {
  if (fileInput.value) {
    // 重置 value 让同一文件可重复选择（change 不会再次触发）
    fileInput.value.value = ''
  }
}

function clearSelection() {
  errorMessage.value = ''
  localFileName.value = ''
  localFileSize.value = 0
  if (localPreviewUrl.value) {
    URL.revokeObjectURL(localPreviewUrl.value)
    localPreviewUrl.value = ''
  }
  emit('update:url', '')
  emit('clear')
  resetInput()
}

function formatError(err: unknown): string {
  if (err && typeof err === 'object' && 'message' in err) {
    return String((err as { message: string }).message)
  }
  return '上传失败，请稍后重试。'
}

async function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > MAX_FILE_SIZE_BYTES) {
    const message = `文件过大：${formatSize(file.size)}，上限 10 MB。`
    errorMessage.value = message
    toast.error(message)
    emit('error', message)
    resetInput()
    return
  }

  // 本地预览：仅在 image/* 类型时使用 objectURL，避免其它类型拿到无法渲染的 URL
  if (localPreviewUrl.value) URL.revokeObjectURL(localPreviewUrl.value)
  if (file.type.startsWith('image/')) {
    localPreviewUrl.value = URL.createObjectURL(file)
  } else {
    localPreviewUrl.value = ''
  }
  localFileName.value = file.name
  localFileSize.value = file.size

  uploading.value = true
  errorMessage.value = ''
  try {
    const asset = await uploadAsset(file)
    toast.success(`已上传：${asset.display_name || asset.url}`)
    emit('update:url', asset.url)
    emit('upload', asset)
  } catch (err) {
    const message = formatError(err)
    errorMessage.value = message
    toast.error(message)
    emit('error', message)
    // 失败时回退本地预览，但保留选中状态便于重试
  } finally {
    uploading.value = false
    resetInput()
  }
}
</script>

<style scoped lang="scss">
.upload-button {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.upload-button__input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.upload-button__zone {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  border: var(--border-thin) dashed var(--border-muted);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
  cursor: pointer;
  transition: border-color 120ms ease, background-color 120ms ease, transform 120ms ease;

  &:hover {
    border-color: var(--border-strong);
    background: var(--bg-elevated);
    transform: translateY(-1px);
  }

  &:focus-visible {
    outline: none;
    border-color: var(--border-focus);
    box-shadow: var(--focus-ring);
  }
}

.upload-button.is-dragover .upload-button__zone {
  border-color: var(--accent);
  background: var(--bg-elevated);
}

.upload-button__zone-icon {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  background: var(--bg-elevated);
  color: var(--accent);
  border: var(--border-thin) solid var(--border-muted);
}

.upload-button__zone-copy {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.upload-button__zone-title {
  font-weight: 600;
  color: var(--text-primary);
}

.upload-button__zone-hint {
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.upload-button__preview {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: var(--border-thick) solid var(--border-muted);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.upload-button__preview-thumb {
  display: grid;
  place-items: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-sm);
  background: var(--surface-soft);
  border: var(--border-thin) solid var(--border-muted);
  overflow: hidden;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.upload-button__preview-meta {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.upload-button__preview-name {
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-button__preview-size {
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.upload-button__preview-actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.upload-button__error {
  margin: 0;
  color: var(--danger);
  font-size: var(--text-sm);
}

.upload-button.is-error .upload-button__zone,
.upload-button.is-error .upload-button__preview {
  border-color: var(--danger);
  background: var(--danger-soft);
}

.upload-button.is-disabled .upload-button__zone {
  cursor: not-allowed;
  opacity: 0.6;
  pointer-events: none;
}
</style>
