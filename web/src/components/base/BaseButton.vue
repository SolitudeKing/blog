<template>
  <button
    class="mist-button"
    :class="classes"
    :type="type"
    :disabled="disabled || loading"
    :aria-busy="loading ? 'true' : undefined"
    :data-state="loading ? 'loading' : disabled ? 'disabled' : undefined"
  >
    <span v-if="loading" class="mist-button__spinner" aria-hidden="true" />
    <span class="mist-button__content"><slot /></span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    size: 'md',
    loading: false,
    disabled: false,
    type: 'button',
  },
)

const classes = computed(() => [
  `mist-button--${props.variant}`,
  `mist-button--${props.size}`,
  {
    'is-loading': props.loading,
    'is-disabled': props.disabled,
  },
])
</script>
