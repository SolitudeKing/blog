<template>
  <div
    class="mist-field"
    :class="{
      'mist-field--disabled': disabled,
      'mist-field--error': !!error,
    }"
  >
    <label class="mist-field__label" :for="controlId">
      {{ label }}<span v-if="required" class="mist-field__required" aria-hidden="true"> *</span>
    </label>
    <textarea
      v-bind="attrs"
      :id="controlId"
      class="mist-input mist-textarea"
      :value="modelValue"
      :rows="rows"
      :name="name"
      :placeholder="placeholder"
      :disabled="disabled"
      :required="required"
      :readonly="readonly"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="describedBy"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
    <span v-if="hint" :id="hintId" class="mist-field__hint">{{ hint }}</span>
    <span v-if="error" :id="errorId" class="mist-field__error" role="alert">{{ error }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs, useId } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue: string
    id?: string
    name?: string
    label: string
    rows?: number
    placeholder?: string
    hint?: string
    error?: string
    disabled?: boolean
    required?: boolean
    readonly?: boolean
  }>(),
  {
    rows: 6,
    disabled: false,
    required: false,
    readonly: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const attrs = useAttrs()
const generatedId = useId()
const controlId = computed(() => props.id || `mist-textarea-${generatedId}`)
const hintId = computed(() => `${controlId.value}-hint`)
const errorId = computed(() => `${controlId.value}-error`)
const describedBy = computed(() => {
  const ids = [props.hint ? hintId.value : '', props.error ? errorId.value : ''].filter(Boolean)
  return ids.length ? ids.join(' ') : undefined
})
</script>
