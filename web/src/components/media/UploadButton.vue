<template>
  <div class="upload-button">
    <input
      ref="fileInput"
      type="file"
      :accept="accept"
      class="upload-button__input"
      :disabled="disabled || uploading"
      @change="onFileChange"
    />
    <BaseButton
      type="button"
      :variant="variant"
      :loading="uploading"
      :disabled="disabled"
      @click="openPicker"
    >
      <SvgIcon name="arrow-up-right" :size="16" />
      <span>{{ uploading ? '上传中…' : label }}</span>
    </BaseButton>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import BaseButton from '@/components/base/BaseButton.vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
import { uploadAsset } from '@/api/modules/asset'
import { useToast } from '@/composables/useToast'
import type { AssetItem } from '@/types/asset'

// 与后端 server/internal/service/asset_service.go 中 maxAssetUploadSize 对齐：
// 10 MB 是写入磁盘前的硬上限，超出直接拒绝以避免大请求体占用带宽。
const MAX_FILE_SIZE_BYTES = 10 * 1024 * 1024

const props = withDefaults(
  defineProps<{
    label?: string
    accept?: string
    disabled?: boolean
    variant?: 'primary' | 'secondary' | 'ghost'
  }>(),
  {
    label: '上传图片',
    accept: 'image/*',
    disabled: false,
    variant: 'secondary',
  },
)

const emit = defineEmits<{
  upload: [asset: AssetItem]
  error: [message: string]
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const toast = useToast()

function openPicker() {
  if (props.disabled || uploading.value) {
    return
  }
  fileInput.value?.click()
}

function resetInput() {
  if (fileInput.value) {
    // 重置 value 以便同一文件可重复选择（change 事件不会重复触发）
    fileInput.value.value = ''
  }
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
  if (!file) {
    return
  }

  if (file.size > MAX_FILE_SIZE_BYTES) {
    const message = `文件过大：${(file.size / 1024 / 1024).toFixed(1)} MB，上限 10 MB。`
    toast.error(message)
    emit('error', message)
    resetInput()
    return
  }

  uploading.value = true
  try {
    const asset = await uploadAsset(file)
    toast.success(`已上传：${asset.display_name || asset.url}`)
    emit('upload', asset)
  } catch (err) {
    const message = formatError(err)
    toast.error(message)
    emit('error', message)
  } finally {
    uploading.value = false
    resetInput()
  }
}
</script>

<style scoped lang="scss">
.upload-button {
  display: inline-flex;
  align-items: center;
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
</style>