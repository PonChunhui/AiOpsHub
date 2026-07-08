import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import Home from '@/views/Home.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/Register.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      name: 'home',
      component: Home,
      meta: { requiresAuth: true }
    },
    {
      path: '/agents',
      name: 'agents',
      component: () => import('@/views/Agents.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/agents-manage',
      name: 'agents-manage',
      component: () => import('@/views/AgentsManage.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/alerts',
      name: 'alerts',
      component: () => import('@/views/Alerts.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/alerts-manage',
      name: 'alerts-manage',
      component: () => import('@/views/AlertsManage.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/knowledge-base',
      name: 'knowledge-base',
      component: () => import('@/views/KnowledgeBase.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/knowledge-base/edit/:id?',
      name: 'knowledge-base-edit',
      component: () => import('@/views/DocumentEditor.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/users-manage',
      name: 'users-manage',
      component: () => import('@/views/UserManage.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/ai-assistant',
      name: 'ai-assistant',
      component: () => import('@/views/AIAssistant.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/mcp-manage',
      name: 'mcp-manage',
      component: () => import('@/views/MCPManage.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/tools-manage',
      name: 'tools-manage',
      component: () => import('@/views/ToolManage.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/host-manage',
      name: 'host-manage',
      component: () => import('@/views/HostManage.vue'),
      meta: { requiresAuth: true }
    }
  ]
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