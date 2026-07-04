<template>
  <button class="cui-button" :class="classes" :type="type" :disabled="disabled || loading">
    <span v-if="loading" class="cui-button__spinner" aria-hidden="true" />
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost'
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit' | 'reset'
  }>(),
  {
    variant: 'primary',
    loading: false,
    disabled: false,
    type: 'button',
  },
)

const classes = computed(() => [`cui-button--${props.variant}`, { 'is-loading': props.loading }])
</script>
