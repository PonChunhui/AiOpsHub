import type { RouteRecordRaw } from 'vue-router'

const agentRoutes: RouteRecordRaw[] = [
  {
    path: '/agents',
    name: 'agents',
    component: () => import('@/views/agents/Agents.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/agents-manage',
    name: 'agents-manage',
    component: () => import('@/views/agents/AgentsManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default agentRoutes
