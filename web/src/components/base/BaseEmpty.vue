<template>
  <div class="mist-empty" :class="{ 'mist-empty--compact': compact }">
    <div class="mist-empty__icon" aria-hidden="true">
      <slot name="icon">
        <SvgIcon name="empty-inbox" />
      </slot>
    </div>
    <div v-if="title" class="mist-empty__title" role="heading" aria-level="4">{{ title }}</div>
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
import SvgIcon from './SvgIcon.vue'

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
