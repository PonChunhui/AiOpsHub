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
        class="drawer-root"
        role="dialog"
        aria-modal="true"
        aria-label="历史对话"
      >
        <div class="drawer-backdrop" @click="historyDrawerOpen = false" />
        <aside class="drawer-panel" @click.stop>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h, watch } from 'vue'
import { Plus, Delete, Loading, Cpu } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { TrBubbleList, TrBubbleProvider, TrSender, TrHistory, TrIconButton } from '@opentiny/tiny-robot'
import { IconHistory, IconNewSession, IconClose, IconAi, IconUser } from '@opentiny/tiny-robot-svgs'
import { useMessage, sseStreamToGenerator } from '@opentiny/tiny-robot-kit'
import AgentVisualization from '@/components/tiny-robot/AgentVisualization.vue'
import CustomMarkdownRenderer from '@/components/tiny-robot/CustomMarkdownRenderer.vue'
import { chatApi } from '@/api'

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
        enable_thinking: deepThinkingEnabled.value // 添加深度思考参数
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

// 监听深度思考状态，控制推理内容的展开
watch([messages, deepThinkingEnabled], ([newMessages, newDeepThinking]) => {
  // 更新所有 assistant 消息的 state.open 属性
  if (newMessages && Array.isArray(newMessages)) {
    newMessages.forEach(msg => {
      if (msg.role === 'assistant' && msg.state) {
        msg.state.open = newDeepThinking
      }
    })
  }
}, { immediate: true, deep: true })

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
      const convertedMessages = history.map(msg => ({
        role: msg.role,
        content: msg.content,
        id: msg.id,
        state: msg.role === 'assistant' ? {
          agentVisualization: {
            agentPath: [],
            events: []
          },
          open: false // 默认折叠推理内容
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
      
      if (sessions.value.length > 0) {
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
      if (sessions.value.length > 0 && !currentSessionId.value) {
        await selectSession(sessions.value[0].id)
      }
    }
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

onMounted(() => {
  loadSessions()
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
  left: 0;
  top: 0;
  bottom: 0;
  width: 300px;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 8px;
  background: #fff;
  box-shadow: 4px 0 24px rgba(0, 0, 0, 0.12);
  border-right: 1px solid #e4e7ed;
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
</style>