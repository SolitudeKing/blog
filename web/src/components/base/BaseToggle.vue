<template>
  <button
    class="mist-toggle"
    :class="[`mist-toggle--${size}`, { 'is-on': modelValue }]"
    type="button"
    :name="name"
    :disabled="disabled"
    :aria-pressed="modelValue"
    :aria-disabled="disabled ? 'true' : undefined"
    :aria-label="ariaLabel || (modelValue ? labelOn : labelOff) || '切换选项'"
    :data-state="disabled ? 'disabled' : modelValue ? 'on' : 'off'"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span class="mist-toggle__track">
      <span class="mist-toggle__thumb" />
    </span>
    <span v-if="labelOn || labelOff" class="mist-toggle__label">
      {{ modelValue ? labelOn : labelOff }}
    </span>
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: boolean
    labelOn?: string
    labelOff?: string
    size?: 'sm' | 'md' | 'lg'
    name?: string
    disabled?: boolean
    ariaLabel?: string
  }>(),
  {
    labelOn: '',
    labelOff: '',
    size: 'md',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()
</script>
