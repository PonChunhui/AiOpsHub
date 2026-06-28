import axios from 'axios'
import { ElMessage } from 'element-plus'

export interface ApiResponse {
  code: number
  message: string
  data?: any
  results?: any[]
  documents?: any[]
  servers?: any[]
  total?: number
  tools?: any[]
  success?: boolean
}



const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(
  config => {
    console.log('发送请求:', config.url)
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
      console.log('Authorization Header:', config.headers.Authorization)
    }
    return config
  },
  error => {
    console.error('请求拦截器错误:', error)
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  (response: any): ApiResponse => {
    console.log('响应成功:', response.config.url, response.data)
    return response.data as ApiResponse
  },
  error => {
    console.error('响应错误:', error.config?.url, error.message)
    
    if (error.response?.status === 401) {
      console.log('Token失效，跳转登录页')
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('userId')
      
      if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
        window.location.href = '/login'
      }
    }
    
    ElMessage.error('请求失败: ' + (error.response?.data?.message || error.message))
    return Promise.reject(error)
  }
)

export const authApi = {
  login: (username: string, password: string): Promise<ApiResponse> => 
    api.post('/auth/login', { username, password }),
  register: (username: string, email: string, password: string): Promise<ApiResponse> =>
    api.post('/auth/register', { username, email, password }),
  logout: (): Promise<ApiResponse> => api.post('/auth/logout')
}

export const alertApi = {
  list: () => api.get('/alerts'),
  get: (id: string) => api.get(`/alerts/${id}`),
  create: (data: any) => api.post('/alerts', data),
  delete: (id: string) => api.delete(`/alerts/${id}`),
  deleteAnalysis: (id: string): Promise<ApiResponse> => api.delete(`/alerts/analysis/${id}`)
}

export const agentApi = {
  list: () => api.get('/agents'),
  get: (id: string) => api.get(`/agents/${id}`),
  create: (data: any) => api.post('/agents', data),
  update: (id: string, data: any) => api.put(`/agents/${id}`, data),
  delete: (id: string) => api.delete(`/agents/${id}`),
  execute: (id: string, data: any) => api.post(`/agents/${id}/execute`, data)
}

export const userApi = {
  list: () => api.get('/users'),
  get: (id: string) => api.get(`/users/${id}`),
  update: (id: string, data: any) => api.put(`/users/${id}`, data),
  delete: (id: string) => api.delete(`/users/${id}`)
}

export const ragApi = {
  search: (query: string, topK: number = 5): Promise<ApiResponse> =>
    api.post('/rag/search', { query, top_k: topK }),
  getContext: (query: string, maxTokens: number = 500): Promise<ApiResponse> =>
    api.get('/rag/context', { params: { query, max_tokens: maxTokens } }),
  listDocuments: (category?: string, search?: string, page?: number, pageSize?: number): Promise<ApiResponse> =>
    api.get('/rag/documents', { params: { category, search, page, pageSize } }),
  getDocument: (id: string): Promise<ApiResponse> =>
    api.get(`/rag/documents/${id}`),
  addDocument: (data: any): Promise<ApiResponse> =>
    api.post('/rag/documents', data),
  updateDocument: (id: string, data: any): Promise<ApiResponse> =>
    api.put(`/rag/documents/${id}`, data),
  deleteDocument: (id: string): Promise<ApiResponse> =>
    api.delete(`/rag/documents/${id}`)
}

export const mcpApi = {
  listServers: (page?: number, pageSize?: number): Promise<ApiResponse> =>
    api.get('/mcp/servers', { params: { page, pageSize } }),
  getServer: (id: string): Promise<ApiResponse> =>
    api.get(`/mcp/servers/${id}`),
  createServer: (data: any): Promise<ApiResponse> =>
    api.post('/mcp/servers', data),
  updateServer: (id: string, data: any): Promise<ApiResponse> =>
    api.put(`/mcp/servers/${id}`, data),
  deleteServer: (id: string): Promise<ApiResponse> =>
    api.delete(`/mcp/servers/${id}`),
  testServer: (id: string): Promise<ApiResponse> =>
    api.post(`/mcp/servers/${id}/test`),
  getServerTools: (id: string): Promise<ApiResponse> =>
    api.get(`/mcp/servers/${id}/tools`)
}

export const prometheusApi = {
  query: (query: string): Promise<ApiResponse> =>
    api.get('/prometheus/query', { params: { query } }),
  queryRange: (query: string, start: string, end: string, step: string): Promise<ApiResponse> =>
    api.get('/prometheus/query_range', { params: { query, start, end, step } }),
  getServiceMetrics: (service: string): Promise<ApiResponse> =>
    api.get(`/prometheus/service/${service}`),
  getTopServices: (metric: string, limit: number = 5): Promise<ApiResponse> =>
    api.get('/prometheus/top', { params: { metric, limit } }),
  getAlerts: (): Promise<ApiResponse> => api.get('/prometheus/alerts')
}

export const tokenApi = {
  getStats: (): Promise<ApiResponse> => api.get('/tokens/stats'),
  getCost: (): Promise<ApiResponse> => api.get('/tokens/cost'),
  getSessionUsage: (sessionId: string): Promise<ApiResponse> =>
    api.get(`/tokens/session/${sessionId}`),
  estimateCost: (model: string, estimatedTokens: number): Promise<ApiResponse> =>
    api.post('/tokens/estimate', { model, estimated_tokens: estimatedTokens })
}

export const historyApi = {
  get: (id: string): Promise<ApiResponse> => api.get(`/history/${id}`),
  list: (limit: number = 20): Promise<ApiResponse> =>
    api.get('/history/list', { params: { limit } }),
  getStats: (): Promise<ApiResponse> => api.get('/history/stats'),
  getRecent: (hours: number = 24): Promise<ApiResponse> =>
    api.get('/history/recent', { params: { hours } })
}

export const k8sApi = {
  getPods: (namespace?: string): Promise<ApiResponse> =>
    api.get('/k8s/pods', { params: { namespace } }),
  getPod: (namespace: string, name: string): Promise<ApiResponse> =>
    api.get(`/k8s/pods/${namespace}/${name}`),
  getPodLogs: (namespace: string, name: string, container?: string, tail?: number): Promise<ApiResponse> =>
    api.get(`/k8s/pods/${namespace}/${name}/logs`, { params: { container, tail } }),
  getPodYaml: (namespace: string, name: string): Promise<ApiResponse> =>
    api.get(`/k8s/pods/${namespace}/${name}/yaml`),
  deletePod: (namespace: string, name: string): Promise<ApiResponse> =>
    api.delete(`/k8s/pods/${namespace}/${name}`),
  getDeployments: (namespace?: string): Promise<ApiResponse> =>
    api.get('/k8s/deployments', { params: { namespace } }),
  getDeployment: (namespace: string, name: string): Promise<ApiResponse> =>
    api.get(`/k8s/deployments/${namespace}/${name}`),
  getDeploymentYaml: (namespace: string, name: string): Promise<ApiResponse> =>
    api.get(`/k8s/deployments/${namespace}/${name}/yaml`),
  scaleDeployment: (namespace: string, name: string, replicas: number): Promise<ApiResponse> =>
    api.post(`/k8s/deployments/${namespace}/${name}/scale`, { replicas }),
  restartDeployment: (namespace: string, name: string): Promise<ApiResponse> =>
    api.post(`/k8s/deployments/${namespace}/${name}/restart`),
  getServices: (namespace?: string): Promise<ApiResponse> =>
    api.get('/k8s/services', { params: { namespace } }),
  getEvents: (namespace?: string, type?: string): Promise<ApiResponse> =>
    api.get('/k8s/events', { params: { namespace, type } }),
  getResourceUsage: (namespace?: string): Promise<ApiResponse> =>
    api.get('/k8s/usage', { params: { namespace } })
}

export const logApi = {
  queryLoki: (query: string, start: string, end: string, limit: number = 100): Promise<ApiResponse> =>
    api.post('/logs/loki/query', { query, start, end, limit }),
  query: (params: any): Promise<ApiResponse> => api.post('/logs/query', params),
  getStats: (): Promise<ApiResponse> => api.get('/logs/stats'),
  getServiceLogs: (service: string, limit: number = 50): Promise<ApiResponse> =>
    api.get(`/logs/service/${service}`, { params: { limit } }),
  getErrorLogs: (limit: number = 20): Promise<ApiResponse> =>
    api.get('/logs/errors', { params: { limit } }),
  search: (keywords: string[], limit: number = 50): Promise<ApiResponse> =>
    api.post('/logs/search', { keywords, limit }),
  getRecent: (minutes: number = 30): Promise<ApiResponse> =>
    api.get('/logs/recent', { params: { minutes } }),
  export: (format: string = 'json'): Promise<ApiResponse> =>
    api.get('/logs/export', { params: { format } })
}

export const remediationApi = {
  getRules: (): Promise<ApiResponse> => api.get('/remediation/rules'),
  createRule: (data: any): Promise<ApiResponse> => api.post('/remediation/rules', data),
  updateRule: (id: string, data: any): Promise<ApiResponse> => api.put(`/remediation/rules/${id}`, data),
  deleteRule: (id: string): Promise<ApiResponse> => api.delete(`/remediation/rules/${id}`),
  getHistory: (params: any): Promise<ApiResponse> => api.get('/remediation/history', { params }),
  getExecutionLogs: (executionId: string): Promise<ApiResponse> =>
    api.get(`/remediation/executions/${executionId}/logs`),
  executeManual: (data: any): Promise<ApiResponse> => api.post('/remediation/manual/execute', data),
  dryRun: (data: any): Promise<ApiResponse> => api.post('/remediation/manual/dry-run', data),
  createPlan: (alertId: string, alertName: string): Promise<ApiResponse> =>
    api.post('/remediation/plans', { alert_id: alertId, alert_name: alertName }),
  executePlan: (planId: string): Promise<ApiResponse> =>
    api.post(`/remediation/plans/${planId}/execute`),
  getPlan: (planId: string): Promise<ApiResponse> =>
    api.get(`/remediation/plans/${planId}`),
  listPlans: (limit: number = 20): Promise<ApiResponse> =>
    api.get('/remediation/plans', { params: { limit } }),
  cancelPlan: (planId: string): Promise<ApiResponse> =>
    api.post(`/remediation/plans/${planId}/cancel`),
  approveAction: (actionId: string): Promise<ApiResponse> =>
    api.post(`/remediation/actions/${actionId}/approve`),
  getStats: (): Promise<ApiResponse> => api.get('/remediation/stats')
}

// AI对话助手API
export const chatApi = {
  // 创建新会话
  createSession: (title: string, model: string = 'gpt-3.5-turbo'): Promise<ApiResponse> =>
    api.post('/chat/sessions', { title, model }),
  // 获取用户的所有会话列表
  getSessions: (): Promise<ApiResponse> => api.get('/chat/sessions'),
  // 获取会话历史记录
  getSessionHistory: (sessionId: string): Promise<ApiResponse> =>
    api.get(`/chat/sessions/${sessionId}/history`),
  // 删除会话
  deleteSession: (sessionId: string): Promise<ApiResponse> =>
    api.delete(`/chat/sessions/${sessionId}`),
  // 发送消息并获取AI回复
  sendMessage: (sessionId: string, content: string): Promise<ApiResponse> =>
    api.post('/chat/messages', { session_id: sessionId, content })
}

export default api