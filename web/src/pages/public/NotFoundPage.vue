<template>
  <section class="not-found-page" aria-labelledby="not-found-title">
    <div class="not-found-stage mist-luminous">
      <div class="not-found-stage__copy">
        <p class="not-found-stage__code">404 · LOST IN THE MIST</p>
        <h1 id="not-found-title">这条路隐入了雾中</h1>
        <p>页面可能已经迁移、删除，或者从未在这里出现。你可以回到熟悉的入口继续漫游。</p>
        <div class="not-found-stage__actions">
          <RouterLink class="mist-button" :to="primaryTarget">{{ primaryLabel }}</RouterLink>
          <RouterLink class="mist-button mist-button--secondary" :to="secondaryTarget">
            {{ secondaryLabel }}
          </RouterLink>
        </div>
      </div>

      <svg class="not-found-stage__field" viewBox="0 0 520 360" aria-hidden="true">
        <path d="M-30 185c78-68 117 64 195-4s117-50 195 4 117 66 195-8" />
        <path d="M-30 228c78-42 117 40 195-2s117-29 195 2 117 38 195-5" />
        <circle cx="346" cy="112" r="58" />
        <path d="M104 94h132M104 126h92M104 158h112" />
      </svg>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const isAdminPath = computed(() => route.path.startsWith('/admin'))
const primaryTarget = computed(() => (isAdminPath.value ? '/admin' : '/'))
const primaryLabel = computed(() => (isAdminPath.value ? '返回创作工作台' : '返回首页'))
const secondaryTarget = computed(() => (isAdminPath.value ? '/' : '/archives'))
const secondaryLabel = computed(() => (isAdminPath.value ? '查看博客' : '浏览文章归档'))
</script>
