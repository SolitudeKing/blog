<template>
  <div class="mist-theme-controls" :class="`mist-theme-controls--${size}`">
    <button
      class="mist-theme-control"
      type="button"
      :aria-label="modeActionLabel"
      :title="modeActionLabel"
      @click="cycleMode"
    >
      <SvgIcon class="mist-theme-control__icon" :name="mode === 'light' ? 'sun' : 'moon'" />
      <span class="mist-theme-control__copy">
        <span>模式</span>
        <strong>{{ modeLabel }}</strong>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import SvgIcon from '@/components/base/SvgIcon.vue'
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
