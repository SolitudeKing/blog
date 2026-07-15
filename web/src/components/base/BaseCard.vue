<template>
  <article class="mist-card" :class="classes">
    <header v-if="title || $slots.header || $slots.actions" class="mist-card__header">
      <slot name="header">
        <div class="mist-card__heading">
          <h3 v-if="title">{{ title }}</h3>
          <p v-if="subtitle">{{ subtitle }}</p>
        </div>
      </slot>
      <div v-if="$slots.actions" class="mist-card__actions">
        <slot name="actions" />
      </div>
    </header>
    <div class="mist-card__body" :class="`mist-card__body--${padding}`">
      <slot />
    </div>
    <footer v-if="$slots.footer" class="mist-card__footer">
      <slot name="footer" />
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    title?: string
    subtitle?: string
    elevation?: 'flat' | 'xs' | 'sm' | 'md'
    hoverable?: boolean
    bordered?: boolean
    padding?: 'sm' | 'md' | 'lg'
    accent?: boolean
  }>(),
  {
    elevation: 'xs',
    hoverable: false,
    bordered: true,
    padding: 'md',
    accent: false,
  },
)

const classes = computed(() => [
  `mist-card--${props.elevation}`,
  `mist-card--pad-${props.padding}`,
  {
    'mist-card--hoverable': props.hoverable,
    'mist-card--bordered': props.bordered,
    'mist-card--accent': props.accent,
  },
])
</script>