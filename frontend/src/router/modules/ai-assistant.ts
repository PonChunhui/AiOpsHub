import type { RouteRecordRaw } from 'vue-router'

const aiAssistantRoutes: RouteRecordRaw[] = [
  {
    path: '/ai-assistant',
    name: 'ai-assistant',
    component: () => import('@/views/ai-assistant/AIAssistant.vue'),
    meta: { requiresAuth: true }
  }
]

export default aiAssistantRoutes
