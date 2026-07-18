import type { RouteRecordRaw } from 'vue-router'

const userRoutes: RouteRecordRaw[] = [
  {
    path: '/users-manage',
    name: 'users-manage',
    component: () => import('@/views/users/UserManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default userRoutes
