<template>
  <component
    :is="wrapperTag"
    :class="['mist-skeleton', `mist-skeleton--${variant}`, { 'mist-skeleton--animated': animated }]"
    role="status"
    aria-busy="true"
    :aria-label="label"
  >
    <span
      v-for="i in count"
      :key="i"
      class="mist-skeleton__item"
      :style="itemStyle"
      :aria-hidden="true"
    />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    variant?: 'text' | 'rect' | 'circle' | 'card' | 'line'
    width?: string
    height?: string
    count?: number
    gap?: string
    animated?: boolean
    tag?: string
    label?: string
  }>(),
  {
    variant: 'text',
    width: '',
    height: '',
    count: 1,
    gap: '8px',
    animated: true,
    tag: 'div',
    label: '正在加载',
  },
)

const wrapperTag = computed(() => props.tag)

const itemStyle = computed(() => ({
  width: props.width || (props.variant === 'circle' ? props.height || '40px' : '100%'),
  height: props.height || (props.variant === 'circle' ? props.width || '40px' : ''),
  marginBottom: props.gap,
}))
</script>
