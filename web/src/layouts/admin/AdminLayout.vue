<template>
  <div class="admin-layout">
    <aside class="admin-layout__sidebar">
      <RouterLink to="/admin">仪表盘</RouterLink>
      <RouterLink to="/admin/articles">文章</RouterLink>
      <RouterLink to="/admin/taxonomy">分类标签</RouterLink>
      <RouterLink to="/admin/notices">公告</RouterLink>
      <RouterLink to="/admin/settings">站点设置</RouterLink>
    </aside>
    <main class="admin-layout__main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts"></script>
