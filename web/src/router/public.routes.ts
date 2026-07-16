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
        path: 'topics/:slug',
        name: 'topic-detail',
        component: () => import('@/pages/public/TopicPage.vue'),
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
      {
        path: 'about',
        name: 'about',
        component: () => import('@/pages/public/AboutPage.vue'),
      },
      {
        path: ':pathMatch(.*)*',
        name: 'not-found',
        component: () => import('@/pages/public/NotFoundPage.vue'),
      },
    ],
  },
]
