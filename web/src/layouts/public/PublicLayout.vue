<template>
  <div class="public-layout creamy-page">
    <header class="public-layout__header">
      <RouterLink class="public-layout__brand" to="/">Solitude Blog</RouterLink>
      <nav class="public-layout__nav" aria-label="Primary">
        <RouterLink to="/">首页</RouterLink>
        <RouterLink to="/archives">归档</RouterLink>
        <RouterLink to="/search">搜索</RouterLink>
      </nav>
    </header>
    <main class="public-layout__main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts"></script>
