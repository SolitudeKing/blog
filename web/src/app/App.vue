<template>
  <div class="app-root">
    <RouterView />
    <BaseToast />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import BaseToast from '@/components/base/BaseToast.vue'
import { useSettingStore } from '@/stores/setting'

const settingStore = useSettingStore()

onMounted(() => {
  void settingStore.loadLobby().catch(() => undefined)
})
</script>

<style lang="scss">
.app-root {
  min-height: 100vh;
  min-height: 100dvh;
}
</style>
