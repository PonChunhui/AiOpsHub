<template>
  <div class="ai-assistant">
    <div class="sidebar">
      <div class="sidebar-header">
        <h3>对话历史</h3>
        <el-button type="primary" size="small" @click="createNewSession">
          <el-icon><Plus /></el-icon>
          新对话
        </el-button>
      </div>
      
      <div class="session-list">
        <el-scrollbar height="calc(100vh - 200px)">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="session-item"
            :class="{ active: currentSessionId === session.id }"
            @click="selectSession(session.id)"
          >
            <div class="session-header">
              <div class="session-title">{{ session.title || '新对话' }}</div>
              <el-button
                type="danger"
                size="small"
                text
                @click.stop="deleteSession(session.id)"
                class="delete-btn"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <div class="session-time">{{ formatTime(session.created_at) }}</div>
          </div>
        </el-scrollbar>
      </div>
    </div>

    <div class="main-content">
      <MessageList 
        ref="messageListRef"
        :conversation-rounds="conversationRounds"
        :is-loading="isLoading"
        :user-initial="getUsernameInitial()"
        @show-rag-detail="showRagDetail"
      />
      
      <!-- 输入区域：包含快捷问题和输入框 -->
      <div class="input-area">
        <!-- 快捷问题区域 -->
        <div class="quick-questions">
          <div class="quick-title">常见问题：</div>
          <div class="quick-buttons">
            <el-button
              v-for="question in quickQuestions"
              :key="question"
              size="small"
              @click="sendQuickQuestion(question)"
              :disabled="isLoading"
              class="quick-btn"
            >
              {{ question }}
            </el-button>
          </div>
        </div>
        
        <!-- 输入框容器：包含textarea和发送按钮 -->
        <div class="input-container">
          <!-- 文本输入框：发送后立即可以继续输入新内容 -->
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="3"
            placeholder="请输入您的消息（支持Shift+Enter换行，Enter发送）..."
            :disabled="!currentSessionId"
            @keydown.enter.exact="sendMessage"
            @keydown.enter.shift.exact.prevent="inputMessage += '\n'"
            class="message-input"
            resize="none"
          />
          <!-- 发送按钮：位于输入框内部右下角 -->
          <el-button 
            v-if="!isLoading"
            type="primary" 
            :disabled="!inputMessage.trim() || !currentSessionId"
            @click="sendMessage"
            class="send-button"
          >
            <el-icon><Promotion /></el-icon>
          </el-button>
          <!-- 停止按钮：loading时显示 -->
          <el-button 
            v-else
            type="danger" 
            @click="abortCurrentStream"
            class="send-button stop-button"
          >
            <el-icon><VideoPause /></el-icon>
            <span>停止</span>
          </el-button>
        </div>
      </div>
    </div>
    
    <el-dialog v-model="showRagDialog" title="知识库文档详情" width="50%">
      <div v-if="selectedRagRef" class="rag-detail">
        <h4>{{ selectedRagRef.title }}</h4>
        <el-tag type="info">{{ selectedRagRef.category }}</el-tag>
        <div class="rag-detail-content">{{ selectedRagRef.snippet }}</div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { Plus, Delete, Promotion, VideoPause } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import MessageList from '@/components/chat/MessageList.vue'
import { chatApi } from '@/api'

const messageListRef = ref()
const currentAbortController = ref<AbortController | null>(null)
const sessions = ref<any[]>([])
const currentSessionId = ref<string>('')
const messages = ref<any[]>([])
const inputMessage = ref('')
const isLoading = ref(false)
const showRagDialog = ref(false)
const selectedRagRef = ref<any>(null)

const quickQuestions = ref([
  '如何排查Pod启动失败的问题？',
  '如何查看应用的实时日志？',
  '如何优化应用性能？',
  '如何配置服务自动扩缩容？',
  '如何分析系统告警？',
  '如何执行服务器巡检？'
])

const conversationRounds = computed(() => {
  const rounds: any[] = []
  let i = 0
  
  while (i < messages.value.length) {
    const userMessage = messages.value[i]
    
    if (userMessage.role === 'user') {
      const round = {
        userMessage: userMessage,
        aiMessage: null,
        index: i
      }
      
      if (i + 1 < messages.value.length && messages.value[i + 1].role === 'assistant') {
        const aiMsg = messages.value[i + 1]
        round.aiMessage = aiMsg
        i += 2
      } else {
        i += 1
      }
      
      rounds.push(round)
    } else {
      i += 1
    }
  }
  
  return rounds
})

function getUsernameInitial(): string {
  const username = localStorage.getItem('username') || 'User'
  return username.charAt(0).toUpperCase()
}

function formatTime(time: string): string {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) {
    return '刚刚'
  } else if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`
  } else if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`
  } else {
    return date.toLocaleDateString()
  }
}

async function loadSessions() {
  try {
    const response = await chatApi.getSessions()
    
    if (response && response.data) {
      sessions.value = response.data || []
      
      if (sessions.value.length > 0 && !currentSessionId.value) {
        await selectSession(sessions.value[0].id)
      }
    }
  } catch (error) {
    console.error('Failed to load sessions:', error)
  }
}

async function createNewSession() {
  abortCurrentStream()
  
  try {
    const response = await chatApi.createSession('新对话')
    
    if (response && response.data) {
      const newSession = response.data
      sessions.value.unshift(newSession)
      currentSessionId.value = newSession.id
      messages.value = []
    }
  } catch (error) {
    console.error('Failed to create session:', error)
    ElMessage.error('创建会话失败')
  }
}

async function selectSession(sessionId: string) {
  abortCurrentStream()
  
  currentSessionId.value = sessionId
  messages.value = []
  
  try {
    const response = await chatApi.getSessionHistory(sessionId)
    
    if (response && response.data) {
      const history = response.data
      messages.value = history.messages || []
    }
  } catch (error) {
    console.error('Failed to load session history:', error)
  }
}

async function deleteSession(sessionId: string) {
  abortCurrentStream()
  
  try {
    await chatApi.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    
    if (currentSessionId.value === sessionId) {
      if (sessions.value.length > 0) {
        await selectSession(sessions.value[0].id)
      } else {
        currentSessionId.value = ''
        messages.value = []
      }
    }
  } catch (error) {
    console.error('Failed to delete session:', error)
    ElMessage.error('删除会话失败')
  }
}

function abortCurrentStream() {
  if (currentAbortController.value) {
    currentAbortController.value.abort()
    isLoading.value = false
    
    const generatingMsg = messages.value.find(m => 
      m.role === 'assistant' && m.id.startsWith('temp-ai-')
    )
    if (generatingMsg && generatingMsg.content) {
      const msgIndex = messages.value.findIndex(m => m.id === generatingMsg.id)
      messages.value.splice(msgIndex, 1, {
        ...generatingMsg,
        content: generatingMsg.content + '\n\n[已停止]'
      })
    }
    
    ElMessage.warning('已停止生成')
  }
}

async function sendMessage() {
  abortCurrentStream()
  
  if (!inputMessage.value.trim()) return
  
  if (!currentSessionId.value) {
    ElMessage.warning('请先创建或选择一个对话')
    return
  }
  
  const userContent = inputMessage.value.trim()
  inputMessage.value = ''
  isLoading.value = true
  
  const userMessage = {
    id: 'temp-user-' + Date.now(),
    role: 'user',
    content: userContent,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)
  
  const aiMessageId = 'temp-ai-' + Date.now()
  const aiMessage = {
    id: aiMessageId,
    role: 'assistant',
    content: '',
    rag_references: [],
    created_at: new Date().toISOString()
  }
  messages.value.push(aiMessage)
  
  await nextTick()
  messageListRef.value?.scrollToBottom()
  
  currentAbortController.value = new AbortController()
  const signal = currentAbortController.value.signal
  
  try {
    const token = localStorage.getItem('token')
    
    const response = await fetch('/api/v1/chat/messages/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({
        session_id: currentSessionId.value,
        content: userContent
      }),
      signal: signal
    })
    
    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`请求失败: ${response.status} ${errorText}`)
    }
    
    if (!response.body) {
      throw new Error('响应体为空')
    }
    
    // ============ SSE流式响应解析开始 ============
    // 创建UTF-8解码器，用于将二进制数据转换为文本
    const decoder = new TextDecoder('utf-8')
    // buffer用于暂存未完成的行，处理跨chunk的数据分割
    let buffer = ''
    // currentEvent记录当前正在处理的SSE事件类型
    let currentEvent = ''
    
    // 获取响应流的reader，用于逐块读取数据
    const reader = response.body.getReader()
    
    try {
      // 循环读取流数据直到结束
      while (true) {
        const { done, value } = await reader.read()
        
        // 流传输结束
        if (done) {
          console.log('[SSE] Stream completed')
          // 检查buffer是否还有未处理的数据（异常情况）
          if (buffer.trim()) {
            console.warn('[SSE] Remaining buffer on stream end:', buffer)
          }
          break
        }
        
        // 解码当前chunk并追加到buffer
        const chunk = decoder.decode(value, { stream: true })
        buffer += chunk
        
        // 逐行解析buffer中的SSE数据
        // 使用indexOf逐行提取，避免split导致的跨chunk分割问题
        let newlineIndex: number
        while ((newlineIndex = buffer.indexOf('\n')) !== -1) {
          // 提取一行完整数据
          const line = buffer.substring(0, newlineIndex)
          // 从buffer中移除已处理的行
          buffer = buffer.substring(newlineIndex + 1)
          
          // 空行：SSE事件结束标记，重置当前事件类型
          if (!line.trim()) {
            currentEvent = ''
            continue
          }
          
          // event行：记录事件类型（如user_message、chunk、done等）
          if (line.startsWith('event:')) {
            currentEvent = line.substring(6).trim()
            console.log(`[SSE] Received event type: ${currentEvent}`)
            continue
          }
          
          // data行：解析JSON数据内容
          if (line.startsWith('data:')) {
            const data = line.substring(5).trim()
            if (!data) continue
          
            try {
              // 解析JSON数据
              const parsed = JSON.parse(data)
              console.log(`[SSE] Event: ${currentEvent}, Data:`, parsed)
              
              // 找到AI消息在列表中的索引位置
              const msgIndex = messages.value.findIndex(m => m.id === aiMessageId)
              
              // 处理user_message事件：更新用户消息的真实ID
              if (currentEvent === 'user_message' && parsed.id) {
                const userMsgIndex = messages.value.findIndex(m => m.id === userMessage.id)
                if (userMsgIndex !== -1) {
                  messages.value.splice(userMsgIndex, 1, {
                    ...messages.value[userMsgIndex],
                    id: parsed.id
                  })
                }
              }
              
              // 处理rag_references事件：添加RAG检索的知识库引用
              if (currentEvent === 'rag_references') {
                if (msgIndex !== -1 && Array.isArray(parsed)) {
                  messages.value.splice(msgIndex, 1, {
                    ...messages.value[msgIndex],
                    rag_references: parsed
                  })
                  // 确保UI更新后滚动到底部
                  await nextTick()
                  messageListRef.value?.scrollToBottom()
                }
              }
              
              // 处理chunk事件：追加AI回复内容（流式更新）
              if (currentEvent === 'chunk' && parsed.content) {
                if (msgIndex !== -1) {
                  // 将新内容追加到现有内容
                  const newContent = (messages.value[msgIndex].content || '') + parsed.content
                  messages.value.splice(msgIndex, 1, {
                    ...messages.value[msgIndex],
                    content: newContent
                  })
                  console.log(`[SSE] Updated content length: ${newContent.length}`)
                  // 每次更新后滚动到底部，确保用户看到最新内容
                  await nextTick()
                  messageListRef.value?.scrollToBottom()
                }
              }
              
              // 处理ai_message事件：更新AI消息的真实ID
              if (currentEvent === 'ai_message' && parsed.id) {
                if (msgIndex !== -1) {
                  messages.value.splice(msgIndex, 1, {
                    ...messages.value[msgIndex],
                    id: parsed.id
                  })
                }
              }
              
              // 处理done事件：流式输出完成，关闭连接
              if (currentEvent === 'done') {
                console.log('[SSE] Done event received, closing stream')
                currentAbortController.value.abort()
                break
              }
              
              // 处理error事件：服务器返回错误
              if (currentEvent === 'error') {
                throw new Error(parsed.message || '服务器错误')
              }
              
            } catch (parseError: any) {
              // JSON解析错误处理
              if (parseError.message !== '服务器错误') {
                console.error('[SSE] Parse error:', parseError, 'Data:', data)
              } else {
                // 服务器错误直接抛出
                throw parseError
              }
            }
          }
        }
        
        // 如果请求被中止（如done事件），退出循环
        if (signal.aborted) {
          break
        }
      }
    } finally {
      // 确保reader被正确关闭
      if (!signal.aborted) {
        reader.cancel()
      }
    }
    // ============ SSE流式响应解析结束 ============
    
  } catch (error: any) {
    if (error.name === 'AbortError') {
      console.log('[SSE] Request aborted successfully')
    } else {
      console.error('[SSE] Error:', error)
      
      let errorMessage = '发送消息失败'
      
      if (error.message?.includes('Failed to fetch') || error.message?.includes('NetworkError')) {
        errorMessage = '网络连接失败，请检查网络或后端服务是否可用'
      } else if (error.message?.includes('502') || error.response?.status === 502) {
        errorMessage = '服务器错误(502)：后端服务不可用，请稍后重试'
      } else if (error.message?.includes('503') || error.response?.status === 503) {
        errorMessage = '服务器错误(503)：服务暂时不可用，请稍后重试'
      } else if (error.message?.includes('504') || error.response?.status === 504) {
        errorMessage = '服务器错误(504)：请求超时，请稍后重试'
      } else if (error.message?.includes('请求失败')) {
        errorMessage = error.message
      }
      
      ElMessage.error(errorMessage)
      messages.value = messages.value.filter(m => m.id !== userMessage.id && m.id !== aiMessageId)
      inputMessage.value = userContent
    }
  } finally {
    isLoading.value = false
    currentAbortController.value = null
  }
}

function showRagDetail(ref: any) {
  selectedRagRef.value = ref
  showRagDialog.value = true
}

async function sendQuickQuestion(question: string) {
  if (isLoading.value || !currentSessionId.value) return
  
  inputMessage.value = question
  await sendMessage()
}

onMounted(() => {
  loadSessions()
})
</script>

<style scoped>
.ai-assistant {
  display: flex;
  height: calc(100vh - 60px);
  background: #f5f7fa;
  overflow: hidden;
}

.sidebar {
  width: 260px;
  height: 100%;
  background: #fff;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
  overflow: hidden;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #e4e7ed;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.session-list {
  flex: 1;
  overflow: hidden;
}

.session-item {
  padding: 12px 20px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.3s;
}

.session-item:hover {
  background: #f5f7fa;
}

.session-item.active {
  background: #ecf5ff;
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.session-title {
  font-size: 14px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.session-time {
  font-size: 12px;
  color: #909399;
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.3s;
  margin-left: 8px;
  flex-shrink: 0;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
  border-radius: 8px;
  margin: 0 20px 20px 0;
  padding: 0 0 20px 20px;
  box-sizing: border-box;
}

.main-content :deep(.el-scrollbar__view) {
  padding-bottom: 40px !important;
}

/* 输入区域：包含快捷问题和输入框 */
.input-area {
  padding: 20px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
}

/* 快捷问题区域样式 */
.quick-questions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

/* 快捷问题标题 */
.quick-title {
  font-size: 14px;
  color: #606266;
  white-space: nowrap;
}

/* 快捷问题按钮组 */
.quick-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  flex: 1;
}

/* 快捷问题按钮样式 */
.quick-btn {
  border-radius: 16px;
}

/* 输入框容器：包含textarea和发送按钮 */
.input-container {
  position: relative; /* 相对定位，作为按钮的定位基准 */
  width: 100%;
}

/* 输入框样式：为按钮预留右侧空间 */
.message-input {
  width: 100%;
}

/* 为textarea添加右侧padding，为按钮预留空间 */
.message-input :deep(.el-textarea__inner) {
  padding-right: 100px; /* 为按钮预留空间 */
  resize: none; /* 禁止手动调整大小 */
}

/* 发送按钮样式：位于输入框内部右下角 */
.send-button {
  position: absolute; /* 绝对定位 */
  right: 12px; /* 右侧距离 */
  bottom: 12px; /* 底部距离 */
  height: 32px; /* 按钮高度 */
  min-width: 80px; /* 最小宽度 */
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border-radius: 8px;
  font-weight: 500;
  z-index: 10; /* 确保按钮在textarea上方 */
}

/* RAG详情对话框样式 */
.rag-detail {
  padding: 20px;
}

.rag-detail h4 {
  margin: 0 0 12px 0;
}

.rag-detail-content {
  margin-top: 16px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.6;
}
</style>