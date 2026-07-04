<template>
  <label class="cui-field">
    <span class="cui-field__label">{{ label }}</span>
    <textarea
      class="cui-input cui-textarea"
      :value="modelValue"
      :rows="rows"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
  </label>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: string
    label: string
    rows?: number
  }>(),
  {
    rows: 6,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
