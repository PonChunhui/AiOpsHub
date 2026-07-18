<template>
  <div class="ai-assistant-tiny-robot">
    <!-- 主聊天区域 -->
    <div class="main-content">
      <!-- 顶部工具栏 -->
      <div class="toolbar">
        <div class="toolbar-actions">
          <tr-icon-button
            :icon="IconHistory"
            size="28"
            svgSize="20"
            title="历史会话"
            @click="historyDrawerOpen = true"
          />
          <tr-icon-button
            :icon="IconNewSession"
            size="28"
            svgSize="20"
            title="新对话"
            @click="createNewSession"
          />
        </div>
        <h3 class="toolbar-title">{{ currentSession?.title || '新对话' }}</h3>
        <div class="toolbar-actions">
          <tr-icon-button
            :icon="Setting"
            size="28"
            svgSize="20"
            title="配置"
            @click="configDrawerOpen = true"
          />
        </div>
      </div>
      
      <!-- 消息列表 -->
      <div class="messages-container" ref="messagesContainerRef">
        <!-- 欢迎界面 -->
        <div v-if="messages.length === 0 && !isProcessing" class="welcome-container">
          <div class="welcome-title">
            <h2>欢迎使用AI运维助手</h2>
            <p>我可以帮助您进行智能运维、监控分析、故障排查等工作</p>
          </div>
          
          <div class="quick-actions">
            <el-button
              v-for="action in quickActions"
              :key="action.id"
              type="primary"
              size="large"
              @click="handleQuickAction(action.text)"
              class="quick-action-btn"
            >
              {{ action.label }}
            </el-button>
          </div>
        </div>
        
        <!-- TrBubbleList消息显示 -->
        <div v-if="messages.length > 0" class="bubble-list-wrapper">
          <tr-bubble-provider :fallback-content-renderer="CustomMarkdownRenderer">
            <tr-bubble-list
              :messages="messages"
              :role-configs="roleConfigs"
              :content-render-mode="'single'"
              :group-strategy="'divider'"
              :divider-role="'user'"
              :auto-scroll="true"
              ref="bubbleListRef"
            >
              <!-- Agent可视化插槽 -->
              <template #after="{ messages, role }">
                <AgentVisualization
                  v-if="role === 'assistant' && messages[0]?.state?.agentVisualization"
                  :agent-path="(messages[0].state.agentVisualization as any).agentPath"
                  :events="(messages[0].state.agentVisualization as any).events"
                />
              </template>
            </tr-bubble-list>
          </tr-bubble-provider>
        </div>
        
        <!-- 加载状态 -->
        <div v-if="isProcessing && messages.length === 0" class="loading-state">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>AI正在思考...</span>
        </div>
      </div>
      
      <!-- 输入框 -->
      <div class="input-area">
        <tr-sender
          v-model="inputMessage"
          mode="multiple"
          :auto-size="{ minRows: 2, maxRows: 6 }"
          placeholder="请输入您的问题... (Ctrl+Enter发送)"
          :loading="isProcessing"
          :max-length="500"
          show-word-limit
          clearable
          submit-type="ctrlEnter"
          :extensions="extensions"
          @submit="handleSenderSubmit"
          @cancel="handleCancel"
        >
          <template #footer>
            <el-button
              :type="deepThinkingEnabled ? 'primary' : 'default'"
              :icon="Cpu"
              size="small"
              @click="deepThinkingEnabled = !deepThinkingEnabled"
            >
              深度思考
            </el-button>
          </template>
        </tr-sender>
      </div>
    </div>

    <!-- 历史会话抽屉 -->
    <Transition name="drawer">
      <div
        v-if="historyDrawerOpen"
        class="drawer-root drawer-left"
        role="dialog"
        aria-modal="true"
        aria-label="历史对话"
      >
        <div class="drawer-backdrop" @click="historyDrawerOpen = false" />
        <aside class="drawer-panel drawer-panel-left" @click.stop>
          <div class="drawer-header">
            <span class="drawer-title">历史对话</span>
            <div class="drawer-actions">
              <tr-icon-button
                :icon="IconNewSession"
                size="28"
                svgSize="20"
                title="新对话"
                @click="createNewSessionAndCloseDrawer"
              />
              <tr-icon-button
                :icon="IconClose"
                size="28"
                svgSize="20"
                title="关闭"
                @click="historyDrawerOpen = false"
              />
            </div>
          </div>
          <tr-history
            class="drawer-history"
            :data="historyData"
            :selected="currentSessionId"
            :menu-items="historyMenuItems"
            :search-bar="true"
            @item-click="handleHistoryItemClick"
            @item-action="handleHistoryItemAction"
          />
        </aside>
      </div>
    </Transition>

    <!-- 配置面板抽屉 -->
    <Transition name="drawer">
      <div
        v-if="configDrawerOpen"
        class="drawer-root drawer-right"
        role="dialog"
        aria-modal="true"
        aria-label="配置面板"
      >
        <div class="drawer-backdrop" @click="configDrawerOpen = false" />
        <aside class="drawer-panel drawer-panel-right" @click.stop>
          <div class="drawer-header">
            <span class="drawer-title">配置面板</span>
            <div class="drawer-actions">
              <tr-icon-button
                :icon="IconClose"
                size="28"
                svgSize="20"
                title="关闭"
                @click="configDrawerOpen = false"
              />
            </div>
          </div>
          <div class="config-content">
            <!-- 模型配置 -->
            <div class="config-section">
              <div class="config-section-header">
                <el-icon><Monitor /></el-icon>
                <span>模型配置</span>
              </div>
              <div class="config-section-body">
                <div class="config-item">
                  <label class="config-label">模型选择</label>
                  <el-select v-model="selectedModel" placeholder="选择模型" style="width: 100%">
                    <el-option
                      v-for="model in availableModels"
                      :key="model.value"
                      :label="model.label"
                      :value="model.value"
                    />
                  </el-select>
                </div>
                <div class="config-item">
                  <label class="config-label">Temperature</label>
                  <el-slider
                    v-model="temperature"
                    :min="0"
                    :max="2"
                    :step="0.1"
                    show-input
                    :show-input-controls="false"
                  />
                </div>
                <div class="config-item">
                  <label class="config-label">Max Tokens</label>
                  <el-input-number
                    v-model="maxTokens"
                    :min="100"
                    :max="32000"
                    :step="100"
                    style="width: 100%"
                  />
                </div>
              </div>
            </div>

            <!-- 工具配置 -->
            <div class="config-section">
              <div class="config-section-header">
                <el-icon><Tools /></el-icon>
                <span>工具配置</span>
              </div>
              <div class="config-section-body">
                <div v-if="loadingTools" class="config-loading">
                  <el-icon class="is-loading"><Loading /></el-icon>
                  <span>加载工具中...</span>
                </div>
                <div v-else-if="mcpServers.length === 0" class="config-empty">
                  暂无可用的工具
                </div>
                <div v-else class="tool-list">
                  <div
                    v-for="server in mcpServers"
                    :key="server.id"
                    class="tool-item"
                  >
                    <div class="tool-info">
                      <div class="tool-name">{{ server.name }}</div>
                      <div class="tool-desc">{{ server.description || '无描述' }}</div>
                    </div>
                    <el-switch
                      v-model="enabledTools[server.id]"
                      @change="handleToolToggle(server.id, $event)"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- RAG配置 -->
            <div class="config-section">
              <div class="config-section-header">
                <el-icon><Collection /></el-icon>
                <span>RAG配置</span>
              </div>
              <div class="config-section-body">
                <div class="config-item">
                  <div class="config-item-row">
                    <label class="config-label">启用RAG</label>
                    <el-switch v-model="ragEnabled" />
                  </div>
                </div>
                <div v-if="ragEnabled">
                  <div class="config-item">
                    <label class="config-label">检索数量 (Top K)</label>
                    <el-input-number
                      v-model="ragTopK"
                      :min="1"
                      :max="20"
                      style="width: 100%"
                    />
                  </div>
                  <div class="config-item">
                    <label class="config-label">相似度阈值</label>
                    <el-slider
                      v-model="ragThreshold"
                      :min="0"
                      :max="1"
                      :step="0.05"
                      show-input
                      :show-input-controls="false"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { Plus, Delete, Loading, Cpu, Monitor, Tools, Collection, Setting } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { TrBubbleList, TrBubbleProvider, TrSender, TrHistory, TrIconButton } from '@opentiny/tiny-robot'
import { IconHistory, IconNewSession, IconClose, IconAi, IconUser } from '@opentiny/tiny-robot-svgs'
import { useMessage, sseStreamToGenerator } from '@opentiny/tiny-robot-kit'
import AgentVisualization from '@/components/tiny-robot/AgentVisualization.vue'
import CustomMarkdownRenderer from '@/components/tiny-robot/CustomMarkdownRenderer.vue'
import { chatApi, mcpApi } from '@/api'

interface ChatSession {
  id: string
  title: string
  created_at: string
}

const sessions = ref<ChatSession[]>([])
const currentSessionId = ref<string>('')
const inputMessage = ref<string>('')
const messagesContainerRef = ref<HTMLElement | null>(null)
const bubbleListRef = ref<any>(null)
const deepThinkingEnabled = ref<boolean>(false)
const historyDrawerOpen = ref<boolean>(false)
const configDrawerOpen = ref<boolean>(false)

// 配置面板状态
const selectedModel = ref<string>('qwen-turbo')
const temperature = ref<number>(0.7)
const maxTokens = ref<number>(4096)
const ragEnabled = ref<boolean>(true)
const ragTopK = ref<number>(5)
const ragThreshold = ref<number>(0.5)
const mcpServers = ref<any[]>([])
const enabledTools = ref<Record<string, boolean>>({})
const loadingTools = ref<boolean>(false)

// 可用模型列表
const availableModels = [
  { label: 'Qwen-Turbo', value: 'qwen-turbo' },
  { label: 'Qwen-Plus', value: 'qwen-plus' },
  { label: 'Qwen-Max', value: 'qwen-max' },
  { label: 'Qwen3.7-Max', value: 'qwen3.7-max' },
  { label: 'DeepSeek-R1', value: 'deepseek-r1' },
  { label: 'GLM-5.2', value: 'glm-5.2' }
]

const historyData = computed(() => {
  return sessions.value.map(session => ({
    id: session.id,
    title: session.title
  }))
})

const historyMenuItems = [
  { id: 'delete', text: '删除', icon: Delete }
]

const currentSession = computed(() => {
  return sessions.value.find(s => s.id === currentSessionId.value)
})

const quickActions = [
  { id: 1, label: '查询系统状态', text: '帮我查询当前系统的CPU和内存使用情况' },
  { id: 2, label: '故障排查', text: '帮我排查最近1小时的系统异常日志' },
  { id: 3, label: '性能优化建议', text: '分析系统性能并提供优化建议' },
  { id: 4, label: '运维知识问答', text: '什么是Kubernetes的Pod？' }
]

const suggestions = [
  { content: '查询系统状态' },
  { content: '分析故障日志' },
  { content: '性能优化建议' },
  { content: '运维知识问答' },
  { content: '监控告警分析' },
  { content: '部署应用' }
]

const extensions = [
  TrSender.suggestion(suggestions, {
    filterFn: (items, query) => items.filter(item =>
      item.content.toLowerCase().includes(query.toLowerCase())
    )
  })
]

const aiAvatar = h(IconAi, { style: { fontSize: '32px' } })
const userAvatar = h(IconUser, { style: { fontSize: '32px' } })

const roleConfigs = {
  user: {
    placement: 'end' as const,
    shape: 'corner' as const,
    avatar: userAvatar
  },
  assistant: {
    placement: 'start' as const,
    shape: 'corner' as const,
    avatar: aiAvatar
  },
  tool: {
    placement: 'start' as const,
    shape: 'rounded' as const,
    avatar: undefined
  }
}

// 使用tiny-robot-kit的useMessage管理消息和流式响应
const {
  messages,
  requestState,
  processingState,
  isProcessing,
  sendMessage,
  abortRequest
} = useMessage({
  responseProvider: async (requestBody, abortSignal) => {
    if (!currentSessionId.value) {
      throw new Error('未选择会话')
    }

    // 获取启用的工具ID列表
    const enabledToolIds = Object.keys(enabledTools.value).filter(
      id => enabledTools.value[id]
    )

    const response = await fetch('/api/v1/chat/messages/stream/events', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({
        session_id: currentSessionId.value,
        content: requestBody.messages[requestBody.messages.length - 1]?.content || '',
        enable_thinking: deepThinkingEnabled.value,
        model: selectedModel.value,
        temperature: temperature.value,
        max_tokens: maxTokens.value,
        enable_rag: ragEnabled.value,
        rag_top_k: ragTopK.value,
        rag_threshold: ragThreshold.value,
        enabled_tools: enabledToolIds
      }),
      signal: abortSignal
    })

    if (!response.ok) {
      throw new Error(`请求失败: ${response.status}`)
    }

    // 后端已发送标准OpenAI格式，sseStreamToGenerator自动解析
    // 无需transform参数，直接使用即可
    return sseStreamToGenerator(response)
  }
})

const handleSenderSubmit = async (text: string) => {
  if (!text.trim() || !currentSessionId.value) return
  
  inputMessage.value = ''
  
  try {
    await sendMessage(text)
  } catch (error: any) {
    ElMessage.error(`发送失败: ${error.message}`)
  }
}

const toggleDeepThinking = () => {
  deepThinkingEnabled.value = !deepThinkingEnabled.value
}

const handleSendMessage = async () => {
  if (!inputMessage.value.trim() || !currentSessionId.value) return
  
  const userContent = inputMessage.value.trim()
  inputMessage.value = ''
  
  try {
    await sendMessage(userContent)
  } catch (error: any) {
    ElMessage.error(`发送失败: ${error.message}`)
  }
}

const handleCancel = async () => {
  await abortRequest()
}

const handleQuickAction = async (text: string) => {
  inputMessage.value = text
  await handleSendMessage()
}

const createNewSession = async () => {
  try {
    const response = await chatApi.createSession('新对话')
    if (response && response.data) {
      sessions.value.unshift(response.data)
      currentSessionId.value = response.data.id
      messages.value = []
    }
  } catch (error) {
    ElMessage.error('创建会话失败')
  }
}

const createNewSessionAndCloseDrawer = async () => {
  await createNewSession()
  historyDrawerOpen.value = false
}

const handleHistoryItemClick = async (item: any) => {
  await selectSession(item.id)
  historyDrawerOpen.value = false
}

const handleHistoryItemAction = async (action: any, item: any) => {
  if (action.id === 'delete') {
    await deleteSession(item.id)
  }
}

const selectSession = async (sessionId: string) => {
  currentSessionId.value = sessionId
  
  try {
    const response = await chatApi.getSessionHistory(sessionId)
    console.log('历史消息API返回:', response)
    
    if (response && response.data) {
      const history = response.data.messages || []
      console.log('历史消息数组:', history)
      
      // 转换历史消息为tiny-robot格式
      const convertedMessages = history.map((msg: any) => ({
        role: msg.role,
        content: msg.content,
        id: msg.id,
        state: msg.role === 'assistant' ? {
          agentVisualization: {
            agentPath: [],
            events: []
          }
        } : {}
      }))
      
      messages.value = convertedMessages
    }
  } catch (error) {
    console.error('加载历史失败:', error)
  }
}

const deleteSession = async (sessionId: string) => {
  try {
    await chatApi.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    
    if (currentSessionId.value === sessionId) {
      messages.value = []
      
      if (sessions.value.length > 0 && sessions.value[0]) {
        await selectSession(sessions.value[0].id)
      } else {
        currentSessionId.value = ''
      }
    }
  } catch (error) {
    ElMessage.error('删除会话失败')
  }
}

const loadSessions = async () => {
  try {
    const response = await chatApi.getSessions()
    if (response && response.data) {
      sessions.value = response.data || []
      if (sessions.value.length > 0 && sessions.value[0] && !currentSessionId.value) {
        await selectSession(sessions.value[0].id)
      }
    }
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

const loadMCPServers = async () => {
  try {
    loadingTools.value = true
    const response = await mcpApi.listServers()
    if (response && response.servers) {
      mcpServers.value = response.servers.map((server: any) => ({
        id: server.id,
        name: server.name,
        description: server.description
      }))
      // 默认启用所有工具
      mcpServers.value.forEach((server: any) => {
        if (enabledTools.value[server.id] === undefined) {
          enabledTools.value[server.id] = true
        }
      })
    }
  } catch (error) {
    console.error('加载MCP服务器失败:', error)
  } finally {
    loadingTools.value = false
  }
}

const handleToolToggle = (serverId: string, enabled: boolean) => {
  enabledTools.value[serverId] = enabled
}

onMounted(() => {
  loadSessions()
  loadMCPServers()
})
</script>

<style scoped>
.ai-assistant-tiny-robot {
  display: flex;
  height: calc(100vh - 60px);
  background: #fff;
  position: relative;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-title {
  flex: 1;
  text-align: center;
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px 24px;
  position: relative;
}

.bubble-list-wrapper {
  height: 100%;
  overflow-y: auto;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: #909399;
}

.input-area {
  padding: 16px 24px;
  border-top: 1px solid #e4e7ed;
  background: #fafafa;
  margin-bottom: 16px;
}

.welcome-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px 20px;
}

.welcome-title {
  text-align: center;
  margin-bottom: 40px;
}

.welcome-title h2 {
  font-size: 28px;
  color: #303133;
  margin-bottom: 16px;
}

.welcome-title p {
  font-size: 16px;
  color: #606266;
}

.quick-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  max-width: 600px;
}

.quick-action-btn {
  width: 100%;
  height: 60px;
  font-size: 16px;
}

/* Drawer transition */
.drawer-enter-active,
.drawer-leave-active {
  transition: none;
}

.drawer-enter-active .drawer-backdrop,
.drawer-leave-active .drawer-backdrop {
  transition: opacity 0.28s ease;
}

.drawer-enter-from .drawer-backdrop,
.drawer-leave-to .drawer-backdrop {
  opacity: 0;
}

.drawer-root {
  position: absolute;
  inset: 0;
  z-index: 10;
  pointer-events: none;
}

.drawer-root > * {
  pointer-events: auto;
}

.drawer-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
}

.drawer-panel {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 320px;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 8px;
  background: #fff;
  box-shadow: 0 0 24px rgba(0, 0, 0, 0.12);
}

.drawer-panel-left {
  left: 0;
  border-right: 1px solid #e4e7ed;
}

.drawer-panel-right {
  right: 0;
  border-left: 1px solid #e4e7ed;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 4px 8px;
}

.drawer-title {
  font-weight: 600;
  font-size: 15px;
}

.drawer-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.drawer-history {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.config-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 4px;
}

.config-section {
  margin-bottom: 16px;
}

.config-section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  border-bottom: 1px solid #e4e7ed;
}

.config-section-body {
  padding: 12px 0;
}

.config-item {
  margin-bottom: 16px;
}

.config-item:last-child {
  margin-bottom: 0;
}

.config-label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  color: #606266;
}

.config-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.config-item-row .config-label {
  margin-bottom: 0;
}

.config-loading,
.config-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px;
  color: #909399;
  font-size: 13px;
}

.tool-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tool-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
  transition: background-color 0.2s;
}

.tool-item:hover {
  background: #eef1f6;
}

.tool-info {
  flex: 1;
  min-width: 0;
}

.tool-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.tool-desc {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>