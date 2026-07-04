<template>
  <label class="cui-field">
    <span class="cui-field__label">{{ label }}</span>
    <input
      class="cui-input"
      :value="modelValue"
      :type="type"
      :autocomplete="autocomplete"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
  </label>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: string
    label: string
    type?: string
    autocomplete?: string
  }>(),
  {
    type: 'text',
    autocomplete: 'off',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

