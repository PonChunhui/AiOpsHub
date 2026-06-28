# 前端开发指南

## 技术栈

- **框架**: Vue 3 + TypeScript
- **UI组件**: Element Plus
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios
- **实时通信**: WebSocket

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API客户端
│   │   ├── index.ts      # REST API
│   │   └── websocket.ts  # WebSocket客户端
│   ├── views/            # 页面组件
│   │   ├── Dashboard.vue          # 仪表板
│   │   ├── WorkflowMonitor.vue    # Workflow监控
│   │   ├── AgentsManage.vue       # Agent管理
│   │   ├── CollaborationMonitor.vue # 协作监控
│   │   ├── AlertManagement.vue    # 告警管理
│   │   ├── KnowledgeBase.vue      # 知识库
│   │   └── Settings.vue           # 系统设置
│   ├── components/       # 通用组件
│   │   ├── WorkflowCard.vue       # Workflow卡片
│   │   ├── AgentStatus.vue        # Agent状态
│   │   ├── MetricsChart.vue       # 指标图表
│   │   └── WebSocketStatus.vue    # WebSocket状态
│   ├── router/           # 路由配置
│   │   └── index.ts
│   ├── stores/           # 状态管理
│   │   ├── user.ts       # 用户状态
│   │   ├── workflow.ts   # Workflow状态
│   │   └── agent.ts      # Agent状态
│   ├── styles/           # 样式文件
│   ├── utils/            # 工具函数
│   ├── App.vue           # 根组件
│   └── main.ts           # 入口文件
├── public/               # 公共资源
├── package.json          # 依赖配置
├── vite.config.ts        # Vite配置
├── tsconfig.json         # TypeScript配置
└── .env.development      # 开发环境变量
```

## 开发环境配置

### 1. 安装依赖

```bash
cd frontend
npm install
```

### 2. 环境变量配置

创建 `.env.development`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_TITLE=AiOpsHub
```

### 3. 启动开发服务器

```bash
npm run dev
```

访问: http://localhost:5173

## API客户端使用

### REST API

```typescript
import api from '@/api'

// 用户认证
const token = await api.login({ username: 'admin', password: 'admin123' })

// 执行协作Workflow
const result = await api.executeCollaboration({
  session_id: 'session-001',
  user_query: '分析服务性能问题',
  context: { service: 'order-service' }
})

// 查询Workflow状态
const status = await api.getWorkflowStatus(workflowId)

// 发送Signal
await api.sendSignal(workflowId, {
  signal_name: 'approval',
  value: { approved: true }
})
```

### WebSocket

```typescript
import { WebSocketClient } from '@/api/websocket'

// 创建连接
const ws = new WebSocketClient('ws://localhost:8080/ws')

// 订阅Workflow更新
ws.subscribe('workflow-001')

// 监听消息
ws.onMessage((message) => {
  if (message.type === 'workflow_update') {
    console.log('Workflow更新:', message.data)
  }
})

// 关闭连接
ws.close()
```

## 页面开发示例

### Workflow监控页面

```vue
<template>
  <div class="workflow-monitor">
    <el-card>
      <template #header>
        <span>Workflow监控</span>
      </template>
      
      <el-table :data="workflows">
        <el-table-column prop="workflow_id" label="ID" />
        <el-table-column prop="status" label="状态" />
        <el-table-column prop="created_at" label="创建时间" />
      </el-table>
    </el-card>
    
    <workflow-card 
      v-for="wf in activeWorkflows"
      :key="wf.id"
      :workflow="wf"
      @update="onWorkflowUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import api from '@/api'
import { WebSocketClient } from '@/api/websocket'
import WorkflowCard from '@/components/WorkflowCard.vue'

const workflows = ref([])
const activeWorkflows = ref([])
let wsClient: WebSocketClient

onMounted(async () => {
  // 加载Workflow列表
  workflows.value = await api.listWorkflows()
  
  // 启动WebSocket
  wsClient = new WebSocketClient(import.meta.env.VITE_WS_URL)
  wsClient.onMessage(handleWSMessage)
})

onUnmounted(() => {
  wsClient?.close()
})

function handleWSMessage(message: any) {
  if (message.type === 'workflow_update') {
    // 更新Workflow状态
    const wf = activeWorkflows.value.find(w => w.id === message.data.workflow_id)
    if (wf) {
      Object.assign(wf, message.data)
    }
  }
}

async function onWorkflowUpdate(workflowId: string) {
  const status = await api.getWorkflowStatus(workflowId)
  console.log('状态更新:', status)
}
</script>
```

## 状态管理

### Workflow Store

```typescript
import { defineStore } from 'pinia'
import api from '@/api'

export const useWorkflowStore = defineStore('workflow', {
  state: () => ({
    workflows: [] as Workflow[],
    activeWorkflow: null as Workflow | null,
    loading: false,
    error: null as string | null
  }),
  
  actions: {
    async fetchWorkflows() {
      this.loading = true
      try {
        this.workflows = await api.listWorkflows()
      } catch (error) {
        this.error = error.message
      } finally {
        this.loading = false
      }
    },
    
    async executeCollaboration(input: CollaborationInput) {
      const result = await api.executeCollaboration(input)
      this.activeWorkflow = result
      return result
    },
    
    async sendApproval(workflowId: string, approved: boolean) {
      await api.sendSignal(workflowId, {
        signal_name: 'approval',
        value: { approved }
      })
    }
  }
})
```

## 路由配置

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('@/views/Dashboard.vue')
  },
  {
    path: '/workflows',
    component: () => import('@/views/WorkflowMonitor.vue')
  },
  {
    path: '/agents',
    component: () => import('@/views/AgentsManage.vue')
  },
  {
    path: '/collaboration',
    component: () => import('@/views/CollaborationMonitor.vue')
  },
  {
    path: '/alerts',
    component: () => import('@/views/AlertManagement.vue')
  },
  {
    path: '/knowledge',
    component: () => import('@/views/KnowledgeBase.vue')
  },
  {
    path: '/settings',
    component: () => import('@/views/Settings.vue')
  }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})
```

## 构建和部署

### 开发构建

```bash
npm run build
```

输出到 `dist/` 目录

### 生产部署

1. 构建前端:
```bash
npm run build
```

2. 使用nginx部署:
```nginx
server {
  listen 80;
  server_name aiops.example.com;
  
  location / {
    root /var/www/aiops/dist;
    try_files $uri $uri/ /index.html;
  }
  
  location /api {
    proxy_pass http://localhost:8080;
  }
  
  location /ws {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
  }
}
```

## 测试

### 单元测试

```bash
npm run test
```

### E2E测试

```bash
npm run test:e2e
```

## 常见问题

### 1. WebSocket连接失败

检查后端WebSocket服务是否启动:
```bash
curl http://localhost:8080/health
```

### 2. API请求401错误

检查Token是否有效:
```typescript
localStorage.getItem('token')
```

### 3. 跨域问题

配置vite代理:
```typescript
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
```

## 开发建议

1. **组件复用**: 创建通用组件库，减少重复代码
2. **类型安全**: 使用TypeScript严格模式
3. **状态管理**: 合理使用Pinia，避免过度使用全局状态
4. **性能优化**: 
   - 使用虚拟滚动处理大数据列表
   - 懒加载路由组件
   - 合理使用缓存
5. **错误处理**: 统一错误处理机制，友好的用户提示
6. **实时更新**: WebSocket连接断线重连机制

## 相关文档

- [Vue 3文档](https://vuejs.org/)
- [Element Plus文档](https://element-plus.org/)
- [Pinia文档](https://pinia.vuejs.org/)
- [Vite文档](https://vitejs.dev/)