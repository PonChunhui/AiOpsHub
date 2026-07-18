import type { RouteRecordRaw } from 'vue-router'

const hostRoutes: RouteRecordRaw[] = [
  {
    path: '/host-manage',
    name: 'host-manage',
    component: () => import('@/views/host/HostManage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/host-terminal/:id',
    name: 'host-terminal',
    component: () => import('@/views/host/Terminal.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/host-file-manage',
    name: 'host-file-manage',
    component: () => import('@/views/host/FileManage.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/host-file-manage/:id',
    name: 'host-file-manage-host',
    component: () => import('@/views/host/FileManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default hostRoutes
