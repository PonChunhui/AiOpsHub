import type { RouteRecordRaw } from 'vue-router'

const knowledgeRoutes: RouteRecordRaw[] = [
  {
    path: '/knowledge-base',
    name: 'knowledge-base',
    component: () => import('@/views/knowledge/KnowledgeBase.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/knowledge-base/edit/:id?',
    name: 'knowledge-base-edit',
    component: () => import('@/views/knowledge/DocumentEditor.vue'),
    meta: { requiresAuth: true }
  }
]

export default knowledgeRoutes
