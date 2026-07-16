<template>
  <div class="mist-empty" :class="{ 'mist-empty--compact': compact }">
    <div class="mist-empty__icon" aria-hidden="true">
      <slot name="icon">
        <svg viewBox="0 0 48 48" aria-hidden="true">
          <path d="M9 15.5h30v21H9z" fill="none" stroke="currentColor" stroke-linejoin="round" stroke-width="2.5" />
          <path d="m9 15.5 7-7h16l7 7M18 24h12" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" />
        </svg>
      </slot>
    </div>
    <h4 v-if="title" class="mist-empty__title">{{ title }}</h4>
    <p v-if="description" class="mist-empty__description">{{ description }}</p>
    <div v-if="$slots.default || ctaText" class="mist-empty__actions">
      <slot>
        <RouterLink v-if="ctaTo" class="mist-button" :to="ctaTo">{{ ctaText }}</RouterLink>
        <BaseButton v-else-if="ctaText" @click="$emit('cta')">{{ ctaText }}</BaseButton>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from './BaseButton.vue'

withDefaults(
  defineProps<{
    title?: string
    description?: string
    ctaText?: string
    ctaTo?: string
    compact?: boolean
  }>(),
  {
    compact: false,
  },
)

defineEmits<{
  cta: []
}>()
</script>
