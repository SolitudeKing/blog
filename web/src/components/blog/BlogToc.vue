<template>
  <aside v-if="items.length" class="blog-toc" :aria-labelledby="headingId">
    <div class="blog-toc__header">
      <span class="blog-toc__dot" aria-hidden="true"></span>
      <div :id="headingId" role="heading" aria-level="2">{{ title }}</div>
    </div>
    <nav class="blog-toc__nav" :aria-label="title">
      <a
        v-for="item in items"
        :key="item.id"
        :href="item.href"
        :style="{ '--toc-level': Math.max(item.level ?? 0, 0) }"
        :aria-current="activeId === item.id ? 'location' : undefined"
        @click="$emit('navigate', item)"
      >
        <span>{{ item.label }}</span>
        <small v-if="item.meta">{{ item.meta }}</small>
      </a>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BlogTocItem } from '@/types/toc'

const props = withDefaults(
  defineProps<{
    title?: string
    items: BlogTocItem[]
    idPrefix?: string
    activeId?: string
  }>(),
  {
    title: '目录',
    idPrefix: 'blog-toc',
    activeId: '',
  },
)

defineEmits<{
  navigate: [item: BlogTocItem]
}>()

const headingId = computed(() => `${props.idPrefix}-heading`)
</script>
