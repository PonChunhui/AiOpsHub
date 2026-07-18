import type { RouteRecordRaw } from 'vue-router'

const mcpRoutes: RouteRecordRaw[] = [
  {
    path: '/mcp-manage',
    name: 'mcp-manage',
    component: () => import('@/views/mcp/MCPManage.vue'),
    meta: { requiresAuth: true }
  }
]

export default mcpRoutes
