import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { RouteRecordRaw } from 'vue-router'

import authRoutes from './modules/auth'
import homeRoutes from './modules/home'
import agentRoutes from './modules/agents'
import alertRoutes from './modules/alerts'
import knowledgeRoutes from './modules/knowledge'
import userRoutes from './modules/users'
import aiAssistantRoutes from './modules/ai-assistant'
import mcpRoutes from './modules/mcp'
import toolRoutes from './modules/tools'
import hostRoutes from './modules/host'

const routes: RouteRecordRaw[] = [
  ...authRoutes,
  ...homeRoutes,
  ...agentRoutes,
  ...alertRoutes,
  ...knowledgeRoutes,
  ...userRoutes,
  ...aiAssistantRoutes,
  ...mcpRoutes,
  ...toolRoutes,
  ...hostRoutes
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 路由守卫 - 检查登录状态
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  console.log('路由守卫:', to.path, '需要认证:', to.meta.requiresAuth)
  
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      console.log('未登录，跳转登录页')
      next('/login')
    } else {
      console.log('已登录，允许访问')
      next()
    }
  } else {
    next()
  }
})

export default router
