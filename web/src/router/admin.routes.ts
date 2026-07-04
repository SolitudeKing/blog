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
    ],
  },
]

