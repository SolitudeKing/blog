<template>
  <div class="mist-theme-controls" :class="`mist-theme-controls--${size}`">
    <button
      class="mist-theme-control"
      type="button"
      :aria-label="modeActionLabel"
      :title="modeActionLabel"
      @click="cycleMode"
    >
      <svg v-if="mode === 'light'" class="mist-theme-control__icon" viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="12" cy="12" r="3.5" fill="none" stroke="currentColor" stroke-width="1.8" />
        <path d="M12 2.5v2M12 19.5v2M2.5 12h2M19.5 12h2M5.3 5.3l1.4 1.4M17.3 17.3l1.4 1.4M18.7 5.3l-1.4 1.4M6.7 17.3l-1.4 1.4" fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="1.8" />
      </svg>
      <svg v-else class="mist-theme-control__icon" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M20.2 15.2A8.5 8.5 0 0 1 8.8 3.8 8.5 8.5 0 1 0 20.2 15.2Z" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" />
      </svg>
      <span class="mist-theme-control__copy">
        <span>模式</span>
        <strong>{{ modeLabel }}</strong>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from '@/composables/useTheme'

withDefaults(
  defineProps<{
    size?: 'sm' | 'md'
  }>(),
  {
    size: 'md',
  },
)

const { mode, cycleMode } = useTheme()

const modeLabel = computed(() => (mode.value === 'light' ? '浅色' : '深色'))
const modeActionLabel = computed(() =>
  mode.value === 'light' ? '当前为浅色模式，切换到深色模式' : '当前为深色模式，切换到浅色模式',
)
</script>
