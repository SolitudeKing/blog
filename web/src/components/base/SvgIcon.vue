<template>
  <svg
    v-bind="attrs"
    class="svg-icon"
    :viewBox="icon.viewBox"
    :fill="icon.fill"
    :stroke="icon.stroke"
    :stroke-linecap="icon.strokeLinecap"
    :stroke-linejoin="icon.strokeLinejoin"
    :stroke-width="icon.strokeWidth"
    :width="normalizedSize"
    :height="normalizedSize"
    :role="label ? 'img' : undefined"
    :aria-label="label || undefined"
    :aria-hidden="label ? undefined : 'true'"
    focusable="false"
    v-html="icon.content"
  />
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { svgIcons, type SvgIconName } from '@/config/svgIcons'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  name: SvgIconName
  size?: number | string
  label?: string
}>()

const attrs = useAttrs()
const icon = computed(() => svgIcons[props.name])
const normalizedSize = computed(() => {
  if (props.size === undefined) {
    return undefined
  }
  return typeof props.size === 'number' ? `${props.size}px` : props.size
})
</script>

<style scoped>
.svg-icon {
  display: block;
  max-width: 100%;
  flex: 0 0 auto;
}
</style>
