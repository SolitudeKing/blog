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
    <input
      v-bind="attrs"
      :id="controlId"
      class="mist-input"
      :value="modelValue"
      :type="type"
      :name="name"
      :autocomplete="autocomplete"
      :placeholder="placeholder"
      :disabled="disabled"
      :required="required"
      :readonly="readonly"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="describedBy"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
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
    modelValue: string | number
    id?: string
    name?: string
    label: string
    type?: string
    autocomplete?: string
    placeholder?: string
    hint?: string
    error?: string
    disabled?: boolean
    required?: boolean
    readonly?: boolean
  }>(),
  {
    type: 'text',
    autocomplete: 'off',
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
const controlId = computed(() => props.id || `mist-input-${generatedId}`)
const hintId = computed(() => `${controlId.value}-hint`)
const errorId = computed(() => `${controlId.value}-error`)
const describedBy = computed(() => {
  const ids = [props.hint ? hintId.value : '', props.error ? errorId.value : ''].filter(Boolean)
  return ids.length ? ids.join(' ') : undefined
})
</script>
