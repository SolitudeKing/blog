<template>
  <header class="blog-navbar">
    <RouterLink class="blog-navbar__brand" to="/" aria-label="Solitude Blog home">
      <span class="blog-navbar__brand-mark" aria-hidden="true">S</span>
      <span class="blog-navbar__brand-text">Solitude Blog</span>
    </RouterLink>

    <nav class="blog-navbar__nav" aria-label="Primary">
      <RouterLink v-for="item in navItems" :key="item.to" :to="item.to">
        {{ item.label }}
      </RouterLink>
    </nav>
  </header>
</template>

<script setup lang="ts">
const navItems = [
  { label: '首页', to: '/' },
  { label: '归档', to: '/archives' },
  { label: '搜索', to: '/search' },
]
</script>
