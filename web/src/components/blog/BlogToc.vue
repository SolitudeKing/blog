<template>
  <aside v-if="items.length" class="blog-toc" :aria-label="title">
    <div class="blog-toc__header">
      <span class="blog-toc__dot" aria-hidden="true"></span>
      <strong>{{ title }}</strong>
    </div>
    <nav class="blog-toc__nav">
      <a
        v-for="item in items"
        :key="item.id"
        :href="item.href"
        :style="{ paddingLeft: `${Math.max(item.level ?? 0, 0) * 12}px` }"
      >
        <span>{{ item.label }}</span>
        <small v-if="item.meta">{{ item.meta }}</small>
      </a>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import type { BlogTocItem } from '@/types/toc'

withDefaults(
  defineProps<{
    title?: string
    items: BlogTocItem[]
  }>(),
  {
    title: '目录',
  },
)
</script>
