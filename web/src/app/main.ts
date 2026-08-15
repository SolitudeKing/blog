import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { initTheme } from '@/composables/useTheme'
import { router } from '@/router'
import '@/styles/index.scss'

// 由 router 的 scrollBehavior 统一控制滚动位置，避免浏览器在路由接管前
// 凭历史记录把视口恢复到不期望的位置（例如 `/#latest-posts` 残留 hash）。
if ('scrollRestoration' in window.history) {
  window.history.scrollRestoration = 'manual'
}

initTheme()

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.mount('#app')
