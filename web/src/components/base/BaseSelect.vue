<template>
  <div class="mist-field" :class="{ 'mist-field--disabled': disabled, 'mist-field--error': !!error }">
    <label v-if="label" class="mist-field__label" :for="controlId">
      {{ label }}<span v-if="required" class="mist-field__required" aria-hidden="true"> *</span>
    </label>
    <span class="mist-select">
      <select
        v-bind="attrs"
        :id="controlId"
        class="mist-select__native"
        :value="modelValue"
        :name="name"
        :disabled="disabled"
        :required="required"
        :aria-label="label ? undefined : ariaLabel"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="describedBy"
        @change="onChange"
      >
        <option v-if="placeholder" value="" :disabled="required">{{ placeholder }}</option>
        <option v-for="opt in options" :key="String(opt.value)" :value="opt.value" :disabled="opt.disabled">
          {{ opt.label }}
        </option>
      </select>
      <SvgIcon class="mist-select__chevron" name="chevron-down" />
    </span>
    <span v-if="hint" :id="hintId" class="mist-field__hint">{{ hint }}</span>
    <span v-if="error" :id="errorId" class="mist-field__error" role="alert">{{ error }}</span>
  </div>
</template>

<script setup lang="ts" generic="T extends string | number">
import { computed, useAttrs, useId } from 'vue'
import SvgIcon from '@/components/base/SvgIcon.vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue: T | ''
    options: Array<{ label: string; value: T; disabled?: boolean }>
    id?: string
    name?: string
    label?: string
    ariaLabel?: string
    hint?: string
    error?: string
    disabled?: boolean
    required?: boolean
    placeholder?: string
  }>(),
  {
    ariaLabel: '选择选项',
    disabled: false,
    required: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: T | '']
  change: [value: T | '']
}>()

const attrs = useAttrs()
const generatedId = useId()
const controlId = computed(() => props.id || `mist-select-${generatedId}`)
const hintId = computed(() => `${controlId.value}-hint`)
const errorId = computed(() => `${controlId.value}-error`)
const describedBy = computed(() => {
  const ids = [props.hint ? hintId.value : '', props.error ? errorId.value : ''].filter(Boolean)
  return ids.length ? ids.join(' ') : undefined
})

function onChange(event: Event) {
  const target = event.target as HTMLSelectElement
  const option = target.options[target.selectedIndex] as HTMLOptionElement & { _value?: T | '' }
  const value = Object.prototype.hasOwnProperty.call(option, '_value') ? option._value : option.value
  const typedValue = value as T | ''
  // 先同步 v-model，再通知父组件刷新，避免筛选请求读到上一次选项。
  emit('update:modelValue', typedValue)
  emit('change', typedValue)
}
</script>
