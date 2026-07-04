<template>
  <section class="home-page">
    <div class="home-page__intro">
      <p class="home-page__eyebrow">{{ setting.lobby?.author ?? 'Solitude King' }}</p>
      <h1>{{ setting.lobby?.site_name ?? 'Solitude Blog' }}</h1>
      <p>{{ setting.lobby?.essay ?? 'Keep writing, keep shipping.' }}</p>
    </div>

    <div class="home-page__list">
      <ArticleCard v-for="article in articles" :key="article.id" :article="article" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ArticleCard from '@/components/blog/ArticleCard.vue'
import { getArticleList } from '@/api/modules/article'
import { useSettingStore } from '@/stores/setting'
import type { ArticleListItem } from '@/types/article'

const setting = useSettingStore()
const articles = ref<ArticleListItem[]>([])

onMounted(async () => {
  await setting.loadLobby()
  const result = await getArticleList()
  articles.value = result.data
})
</script>

