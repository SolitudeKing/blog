import type { RouteRecordRaw } from 'vue-router'
import AdminLayout from '@/layouts/admin/AdminLayout.vue'

export const adminRoutes: RouteRecordRaw[] = [
  {
    path: '/admin/login',
    name: 'admin-login',
    component: () => import('@/pages/admin/LoginPage.vue'),
  },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'admin-dashboard',
        component: () => import('@/pages/admin/DashboardPage.vue'),
      },
      {
        path: 'articles',
        name: 'admin-articles',
        component: () => import('@/pages/admin/ArticleListPage.vue'),
      },
      {
        path: 'articles/new',
        name: 'admin-article-new',
        component: () => import('@/pages/admin/ArticleEditorPage.vue'),
      },
      {
        path: 'articles/:id',
        name: 'admin-article-edit',
        component: () => import('@/pages/admin/ArticleEditorPage.vue'),
      },
      {
        path: 'taxonomy',
        name: 'admin-taxonomy',
        component: () => import('@/pages/admin/TaxonomyPage.vue'),
      },
      {
        path: 'media',
        name: 'admin-media',
        component: () => import('@/pages/admin/MediaPage.vue'),
      },
      {
        path: 'notices',
        name: 'admin-notices',
        component: () => import('@/pages/admin/NoticePage.vue'),
      },
      {
        path: 'settings',
        name: 'admin-settings',
        component: () => import('@/pages/admin/SettingPage.vue'),
      },
    ],
  },
]
