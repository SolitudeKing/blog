<template>
  <div class="public-layout creamy-page">
    <BlogNavbar />
    <main class="public-layout__main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import BlogNavbar from '@/components/blog/BlogNavbar.vue'
</script>
