import type { RouteRecordRaw } from 'vue-router'

const alertRoutes: RouteRecordRaw[] = [
  {
    path: '/alerts',
    name: 'alerts',
    component: () => import('@/views/alerts/Alerts.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/alerts-manage',
    name: 'alerts-manage',
    component: () => import('@/views/alerts/AlertsManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default alertRoutes
