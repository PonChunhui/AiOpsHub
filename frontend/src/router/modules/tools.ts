import type { RouteRecordRaw } from 'vue-router'

const toolRoutes: RouteRecordRaw[] = [
  {
    path: '/tools-manage',
    name: 'tools-manage',
    component: () => import('@/views/tools/ToolManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default toolRoutes
