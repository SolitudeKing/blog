import type { RouteRecordRaw } from 'vue-router'
import PublicLayout from '@/layouts/public/PublicLayout.vue'

export const publicRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: PublicLayout,
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/pages/public/HomePage.vue'),
      },
      {
        path: 'articles/:slug',
        name: 'article-detail',
        component: () => import('@/pages/public/ArticleDetailPage.vue'),
      },
      {
        path: 'archives',
        name: 'archives',
        component: () => import('@/pages/public/ArchivesPage.vue'),
      },
      {
        path: 'search',
        name: 'search',
        component: () => import('@/pages/public/SearchPage.vue'),
      },
    ],
  },
]
