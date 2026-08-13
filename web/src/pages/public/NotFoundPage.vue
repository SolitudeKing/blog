<template>
  <section class="not-found-page" aria-labelledby="not-found-title">
    <div class="not-found-stage mist-luminous">
      <div class="not-found-stage__copy">
        <p class="not-found-stage__code">404 · LOST IN THE MIST</p>
        <div id="not-found-title" role="heading" aria-level="1">这条路隐入了雾中</div>
        <p>页面可能已经迁移、删除，或者从未在这里出现。你可以回到熟悉的入口继续漫游。</p>
        <div class="not-found-stage__actions">
          <RouterLink class="mist-button" :to="primaryTarget">{{ primaryLabel }}</RouterLink>
          <RouterLink class="mist-button mist-button--secondary" :to="secondaryTarget">
            {{ secondaryLabel }}
          </RouterLink>
        </div>
      </div>

      <SvgIcon class="not-found-stage__field" name="not-found-field" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import SvgIcon from '@/components/base/SvgIcon.vue'

const route = useRoute()
const isAdminPath = computed(() => route.path.startsWith('/admin'))
const primaryTarget = computed(() => (isAdminPath.value ? '/admin' : '/'))
const primaryLabel = computed(() => (isAdminPath.value ? '返回创作工作台' : '返回首页'))
const secondaryTarget = computed(() => (isAdminPath.value ? '/' : '/archives'))
const secondaryLabel = computed(() => (isAdminPath.value ? '查看博客' : '浏览文章归档'))
</script>
