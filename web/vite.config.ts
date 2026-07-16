import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/healthz': 'http://localhost:8080',
      '/rss.xml': 'http://localhost:8080',
      '/sitemap.xml': 'http://localhost:8080',
      '/auth': 'http://localhost:8080',
      '/user': 'http://localhost:8080',
      '/dashboard': 'http://localhost:8080',
      '^/search/article(?:\\?|$)': 'http://localhost:8080',
      '/asset': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
      '/setting': 'http://localhost:8080',
      '^/article(?:/|\\?|$)': 'http://localhost:8080',
      // API 前缀必须以路径边界结束，避免把前端页面 /topics/:slug 误代理到后端。
      '^/topic(?:/|\\?|$)': 'http://localhost:8080',
      '^/tag(?:/|\\?|$)': 'http://localhost:8080',
      '/notice': 'http://localhost:8080',
    },
  },
})
