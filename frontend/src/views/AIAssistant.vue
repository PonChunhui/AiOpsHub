<template>
  <div class="ai-assistant">
    <!-- 左侧边栏 - 会话历史列表 -->
    <div class="sidebar">
      <div class="sidebar-header">
        <h3>对话历史</h3>
        <el-button type="primary" size="small" @click="createNewSession">
          <el-icon><Plus /></el-icon>
          新对话
        </el-button>
      </div>
      
      <!-- 会话列表 -->
      <div class="session-list">
        <el-scrollbar height="calc(100vh - 200px)">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="session-item"
            :class="{ active: currentSessionId === session.id }"
            @click="selectSession(session.id)"
          >
            <div class="session-title">{{ session.title || '新对话' }}</div>
            <div class="session-time">{{ formatTime(session.created_at) }}</div>
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
        </el-scrollbar>
      </div>
    </div>

    <!-- 右侧主区域 - 对话界面 -->
    <div class="main-content">
      <!-- 消息列表 -->
      <div class="messages-container">
        <el-scrollbar ref="messagesScrollbar" height="calc(100vh - 250px)">
<div class="messages">
            <!-- 所有对话轮次合并到一个容器中 -->
            <div class="conversation-wrapper">
              <template v-for="(round, roundIndex) in conversationRounds" :key="roundIndex">
                <!-- 用户消息 -->
                <div class="user-message">
                  <div class="user-content">{{ round.userMessage.content }}</div>
                  <div class="user-avatar">{{ getUsernameInitial() }}</div>
                </div>
                
                <!-- AI回复 -->
                <div v-if="round.aiMessage" class="assistant-message">
                  <div class="ai-avatar">AI</div>
                  <div class="message-content-wrapper">
                    <!-- 如果正在加载且内容为空，显示加载状态 -->
                    <div v-if="isLoading && !round.aiMessage.content" class="ai-content loading-message">
                      <el-icon class="is-loading"><Loading /></el-icon>
                      <span>AI正在思考...</span>
                    </div>
                    <!-- 否则显示AI回复内容 -->
                    <div v-else class="ai-content-wrapper">
                      <!-- MCP 工具调用显示 -->
                      <div v-if="parseToolCalls(round.aiMessage.content).length > 0" class="tool-calls-section">
                        <div class="tool-calls-header">
                          <el-icon><Setting /></el-icon>
                          <span>MCP 工具调用</span>
                        </div>
                        <div class="tool-calls-list">
                          <div 
                            v-for="(toolCall, idx) in parseToolCalls(round.aiMessage.content)" 
                            :key="idx"
                            class="tool-call-item"
                          >
                            <div class="tool-call-header">
                              <el-tag type="primary" size="small">{{ toolCall.tool }}</el-tag>
                              <span class="tool-server">@{{ toolCall.server }}</span>
                            </div>
                            <div v-if="toolCall.arguments && Object.keys(toolCall.arguments).length > 0" class="tool-args">
                              <el-descriptions :column="1" size="small" border>
                                <el-descriptions-item
                                  v-for="(value, key) in toolCall.arguments"
                                  :key="key"
                                  :label="key"
                                >
                                  {{ JSON.stringify(value) }}
                                </el-descriptions-item>
                              </el-descriptions>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <!-- 工具执行结果显示 -->
                      <div v-if="getToolResult(round.aiMessage.content)" class="tool-result-section">
                        <div class="tool-result-header">
                          <el-icon><DocumentChecked /></el-icon>
                          <span>工具执行结果</span>
                        </div>
                        <div class="tool-result-content">
                          <pre>{{ getToolResult(round.aiMessage.content) }}</pre>
                        </div>
                      </div>
                      
                      <!-- 去除工具调用后的正常内容 -->
                      <div class="ai-content markdown-body" v-html="renderMarkdown(cleanContent(round.aiMessage.content))"></div>
                    </div>
                    
                    <!-- RAG引用显示 -->
                    <div v-if="round.aiMessage.rag_references && round.aiMessage.rag_references.length > 0 && !isLoading" class="rag-references">
                      <div class="rag-header">
                        <el-icon><Reading /></el-icon>
                        <span>已引用 {{ round.aiMessage.rag_references.length }} 篇知识库文档</span>
                      </div>
                      <div class="rag-items">
                        <div 
                          v-for="(ref, refIndex) in round.aiMessage.rag_references" 
                          :key="refIndex"
                          class="rag-item"
                          @click="showRagDetail(ref)"
                        >
                          <div class="rag-item-header">
                            <span class="rag-title">{{ ref.title }}</span>
                            <span class="rag-badge">{{ ref.category }}</span>
                          </div>
                          <div class="rag-snippet">{{ ref.snippet }}</div>
                          <div class="rag-score-bar">
                            <div class="rag-score-fill" :style="{ width: (ref.score * 100) + '%' }"></div>
                          </div>
                          <div class="rag-score-text">相关度 {{ (ref.score * 100).toFixed(0) }}%</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </el-scrollbar>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <!-- MCP 工具选择器 -->
        <div v-if="showToolSelector" class="tool-selector-panel">
          <div class="panel-header">
            <span>MCP 工具选择</span>
            <el-button size="small" text @click="showToolSelector = false">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
          <MCPToolSelector @update:selectedTools="handleToolSelection" />
        </div>
        
        <!-- 快捷提问按钮 -->
        <div class="quick-questions">
          <div class="quick-title">快捷提问：</div>
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
        
        <!-- 消息输入框 -->
        <div class="input-wrapper">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :rows="3"
            placeholder="请输入您的消息（支持Shift+Enter换行，Enter发送）..."
            @keydown.enter.exact="sendMessage"
            @keydown.enter.shift.exact.prevent="inputMessage += '\n'"
            :disabled="isLoading"
            class="message-input"
          />
          <div class="input-actions">
            <el-button
              :type="selectedMCPTools.length > 0 ? 'success' : 'default'"
              @click="showToolSelector = !showToolSelector"
              :disabled="isLoading"
              class="tool-btn"
            >
              <el-icon><Setting /></el-icon>
              {{ selectedMCPTools.length > 0 ? `${selectedMCPTools.length} 工具` : 'MCP 工具' }}
            </el-button>
            <el-button
              type="primary"
              @click="sendMessage"
              :disabled="!inputMessage.trim() || isLoading"
              class="send-btn"
            >
              <el-icon><Promotion /></el-icon>
              发送
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import MarkdownIt from 'markdown-it'
import hljs from 'highlight.js'
import katex from 'markdown-it-katex'
import taskLists from 'markdown-it-task-lists'
import markdownItCodeCopy from 'markdown-it-code-copy'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Promotion, Loading, Reading, Setting, DocumentChecked, Close } from '@element-plus/icons-vue'
import { chatApi } from '@/api/index'
import MCPToolSelector from '@/components/MCPToolSelector.vue'

import 'highlight.js/styles/atom-one-light.css'
import 'katex/dist/katex.min.css'
import 'material-design-icons-iconfont/dist/material-design-icons.css'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
  highlight: function (str: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return `<pre class="hljs"><code>${hljs.highlight(str, { language: lang, ignoreIllegals: true }).value}</code></pre>`
      } catch (__) {}
    }
    return `<pre class="hljs"><code>${md.utils.escapeHtml(str)}</code></pre>`
  }
})

// 使用插件
md.use(katex, {
  throwOnError: false,
  errorColor: '#cc0000'
})

md.use(taskLists, {
  enabled: true,
  label: true,
  labelAfter: true
})

// 使用代码复制插件（使用material-design-icons）
md.use(markdownItCodeCopy, {
  onSuccess: () => {
    ElMessage.success('代码已复制到剪贴板')
  },
  onError: () => {
    ElMessage.error('复制失败')
  }
})

// Mermaid图表处理
function processMermaidCharts(content: string): string {
  const mermaidRegex = /```mermaid\n([\s\S]*?)\n```/g
  return content.replace(mermaidRegex, (match, code) => {
    const id = `mermaid-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    return `<div class="mermaid-chart" id="${id}">${code}</div>`
  })
}

// 渲染Markdown（使用markdown-it）
function renderMarkdown(content: string): string {
  try {
    // 处理Mermaid图表
    content = processMermaidCharts(content)
    
    // 渲染markdown（markdown-it-code-copy插件已自动添加复制按钮）
    let html = md.render(content)
    
    // 后处理：还原被转义的无属性纯标签（如<mount>, <设备>等）
    // 只还原形如 <纯文本> 或 </纯文本> 的标签，排除危险标签
    const dangerousTags = ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button', 'a', 'link', 'style', 'meta', 'base', 'svg', 'math']
    html = html.replace(/&lt;(\/?[\w\u4e00-\u9fa5]+)&gt;/g, (match, tag) => {
      const tagName = tag.replace('/', '').toLowerCase()
      if (dangerousTags.includes(tagName)) {
        return match // 保持转义
      }
      return `<${tag}>`
    })
    
    // 后处理：处理超长行内代码，添加换行机会
    html = processLongInlineCode(html)
    
    // 异步初始化Mermaid图表（在DOM更新后）
    nextTick(() => {
      initMermaidCharts()
    })
    
    return html
  } catch (error) {
    console.error('Markdown渲染失败:', error)
    return content
  }
}

// 处理超长行内代码，在每个字符后添加换行机会
function processLongInlineCode(html: string): string {
  // 匹配行内代码（不在pre标签内的code标签）
  const inlineCodeRegex = /<code(?![^>]*class="language-[^"]*")[^>]*>([^<]+)<\/code>/g
  
  let count = 0
  const result = html.replace(inlineCodeRegex, (match, codeContent) => {
    // 如果代码长度超过20字符，在每个字符后添加<wbr>标签
    if (codeContent.length > 20) {
      count++
      // 将每个字符后都添加<wbr>，确保可以换行
      const processedCode = codeContent.split('').join('<wbr>')
      console.log(`处理行内代码 #${count}: "${codeContent.substring(0, 30)}..." → 添加${processedCode.split('<wbr>').length - 1}个<wbr>标签`)
      return match.replace(codeContent, processedCode)
    }
    return match
  })
  
  console.log(`总共处理了 ${count} 个超长行内代码`)
  return result
}

// 解析工具调用块
function parseToolCalls(content: string): any[] {
  const toolCalls: any[] = []
  const regex = /```tool_call\n([\s\S]*?)\n```/g
  
  let match
  while ((match = regex.exec(content)) !== null) {
    try {
      const jsonStr = match[1].trim()
      const toolCall = JSON.parse(jsonStr)
      toolCalls.push(toolCall)
    } catch (e) {
      console.error('解析工具调用失败:', e)
    }
  }
  
  return toolCalls
}

// 获取工具执行结果
function getToolResult(content: string): string {
  const resultMatch = content.match(/工具.*?执行结果:\n([\s\S]*?)(?:\n\n|$)/)
  if (resultMatch) {
    return resultMatch[1].trim()
  }
  return ''
}

// 清理内容（去除 tool_call 块和工具结果）
function cleanContent(content: string): string {
  // 去除 tool_call 块
  let cleaned = content.replace(/```tool_call\n[\s\S]*?\n```/g, '')
  
  // 去除工具执行结果
  cleaned = cleaned.replace(/工具.*?执行结果:\n[\s\S]*?(?:\n\n|$)/g, '')
  
  // 去除多余的空行
  cleaned = cleaned.replace(/\n{3,}/g, '\n\n')
  
  return cleaned.trim()
}

// 初始化Mermaid图表
async function initMermaidCharts() {
  const mermaidElements = document.querySelectorAll('.mermaid-chart')
  if (mermaidElements.length === 0) return
  
  try {
    const mermaid = (await import('mermaid')).default
    mermaid.initialize({
      startOnLoad: false,
      theme: 'neutral',
      securityLevel: 'loose'
    })
    
    mermaidElements.forEach(async (element) => {
      const id = element.id
      const code = element.textContent || ''
      try {
        const { svg } = await mermaid.render(id, code)
        element.innerHTML = svg
      } catch (err) {
        console.error('Mermaid渲染失败:', err)
        element.innerHTML = `<pre class="error">${code}</pre>`
      }
    })
  } catch (err) {
    console.error('Mermaid初始化失败:', err)
  }
}

// 处理任务列表checkbox点击
if (typeof window !== 'undefined') {
  document.addEventListener('change', (e: Event) => {
    const target = e.target as HTMLElement
    if (target.classList.contains('task-list-item-checkbox')) {
      const checkbox = target as HTMLInputElement
      const listItem = checkbox.closest('.task-list-item')
      if (listItem) {
        if (checkbox.checked) {
          listItem.classList.add('checked')
        } else {
          listItem.classList.remove('checked')
        }
      }
    }
  })
}

// 会话列表
const sessions = ref<any[]>([])
// 当前会话ID
const currentSessionId = ref<string>('')
// 当前会话的消息列表
const messages = ref<any[]>([])
// 输入的消息内容
const inputMessage = ref<string>('')
// 加载状态
const isLoading = ref<boolean>(false)
// 消息滚动条引用
const messagesScrollbar = ref()
// MCP 工具选择器显示
const showToolSelector = ref<boolean>(false)
// 选中的 MCP 工具
const selectedMCPTools = ref<string[]>([])

// 处理工具选择
const handleToolSelection = (tools: string[]) => {
  selectedMCPTools.value = tools
}

// 将messages数组转换为对话轮次数组（每轮包含用户消息和AI回复）
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
      
      // 检查是否有对应的AI回复
      if (i + 1 < messages.value.length && messages.value[i + 1].role === 'assistant') {
        const aiMsg = messages.value[i + 1]
        round.aiMessage = aiMsg
        // 调试：检查AI消息的rag_references
        if (aiMsg.rag_references) {
          console.log(`轮次 ${rounds.length} 的RAG引用:`, aiMsg.rag_references)
        }
        i += 2
      } else {
        i += 1
      }
      
      rounds.push(round)
    } else {
      // 如果第一条消息不是用户消息（比如只有AI消息），跳过
      i += 1
    }
  }
  
  console.log('conversationRounds:', rounds)
  return rounds
})

// 快捷提问列表
const quickQuestions = ref([
  '如何部署应用到Kubernetes？',
  '如何排查Pod启动失败的问题？',
  '如何查看应用的实时日志？',
  '如何优化应用性能？',
  '如何配置服务自动扩缩容？'
])

// 发送快捷提问
async function sendQuickQuestion(question: string) {
  if (isLoading.value) return
  
  if (!currentSessionId.value) {
    ElMessage.warning('请先创建或选择一个对话')
    return
  }
  
  inputMessage.value = question
  await sendMessage()
}

// 获取用户的所有会话列表
async function loadSessions() {
  try {
    const response = await chatApi.getSessions()
    
    if (response.message === '获取成功') {
      sessions.value = response.data
      // 如果有会话，默认选择第一个
      if (sessions.value.length > 0 && !currentSessionId.value) {
        await selectSession(sessions.value[0].id)
      }
    }
  } catch (error: any) {
    console.error('获取会话列表失败:', error)
    ElMessage.error(error.response?.data?.error || '获取会话列表失败')
  }
}

// 创建新会话
async function createNewSession() {
  try {
    const { value: title } = await ElMessageBox.prompt('请输入对话标题', '创建新对话', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：AI运维助手咨询'
    })
    
    if (title) {
      const response = await chatApi.createSession(title)
      
      if (response.message === '会话创建成功') {
        await loadSessions()
        await selectSession(response.data.id)
        ElMessage.success('会话创建成功')
      }
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('创建会话失败:', error)
      ElMessage.error(error.response?.data?.error || '创建会话失败')
    }
  }
}

// 选择会话并加载历史消息
async function selectSession(sessionId: string) {
  currentSessionId.value = sessionId
  try {
    const response = await chatApi.getSessionHistory(sessionId)
    
    if (response.message === '获取成功') {
      console.log('历史消息原始数据:', response.data.messages)
      
      // 为每条消息添加rag_references字段（如果不存在）并解析JSON字符串
      messages.value = (response.data.messages || []).map((msg: any) => {
        let ragReferences = null
        
        // 检查可能的字段名（snake_case或camelCase）
        const ragRefsStr = msg.rag_references || msg.RAGReferences
        
        if (ragRefsStr) {
          console.log(`消息 ${msg.id} 的原始RAG引用字符串:`, ragRefsStr)
          try {
            ragReferences = JSON.parse(ragRefsStr)
            console.log(`消息 ${msg.id} 解析后的RAG引用:`, ragReferences)
          } catch (e) {
            console.error('解析RAG引用失败:', e)
          }
        }
        
        return {
          ...msg,
          rag_references: ragReferences
        }
      })
      console.log('加载历史消息（处理后）:', messages.value)
      // 滚动到底部
      await nextTick()
      scrollToBottom()
    }
  } catch (error: any) {
    console.error('获取会话历史失败:', error)
    ElMessage.error(error.response?.data?.error || '获取会话历史失败')
  }
}

// 删除会话
async function deleteSession(sessionId: string) {
  try {
    await ElMessageBox.confirm('确定要删除这个对话吗？', '删除对话', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await chatApi.deleteSession(sessionId)
    
    ElMessage.success('会话删除成功')
    await loadSessions()
    
    // 如果删除的是当前会话，清空消息
    if (currentSessionId.value === sessionId) {
      messages.value = []
      currentSessionId.value = ''
      // 如果还有其他会话，选择第一个
      if (sessions.value.length > 0) {
        await selectSession(sessions.value[0].id)
      }
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除会话失败:', error)
      ElMessage.error(error.response?.data?.error || '删除会话失败')
    }
  }
}

// 发送消息（流式）
async function sendMessage() {
  if (!inputMessage.value.trim() || isLoading.value) return
  
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
    rag_references: null,
    created_at: new Date().toISOString()
  }
  messages.value.push(aiMessage)
  
  await nextTick()
  scrollToBottom()
  
  try {
    const token = localStorage.getItem('token')
    const response = await fetch('/api/v1/chat/messages/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        session_id: currentSessionId.value,
        content: userContent
      })
    })
    
    if (!response.ok) {
      throw new Error('发送消息失败')
    }
    
    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('无法读取响应流')
    }
    
    const decoder = new TextDecoder()
    let buffer = ''
    let currentEvent = ''
    
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      
      buffer += decoder.decode(value, { stream: true })
      
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''
      
      for (const line of lines) {
        if (line.startsWith('event: ')) {
          currentEvent = line.substring(7).trim()
          continue
        }
        
        if (line.startsWith('data: ')) {
          const data = line.substring(6).trim()
          if (!data) continue
          
          try {
            const parsed = JSON.parse(data)
            
            switch (currentEvent) {
              case 'user_message':
                if (parsed.id) {
                  const msgIndex = messages.value.findIndex(m => m.id === userMessage.id)
                  if (msgIndex !== -1) {
                    messages.value[msgIndex].id = parsed.id
                  }
                }
                break
                
              case 'rag_references':
                if (Array.isArray(parsed)) {
                  console.log('收到RAG引用:', parsed)
                  const msgIndex = messages.value.findIndex(m => m.id === aiMessageId)
                  if (msgIndex !== -1) {
                    messages.value[msgIndex].rag_references = parsed
                    await nextTick()
                    scrollToBottom()
                  }
                }
                break
                
              case 'chunk':
                if (parsed.content) {
                  const msgIndex = messages.value.findIndex(m => m.id === aiMessageId)
                  if (msgIndex !== -1) {
                    messages.value[msgIndex].content += parsed.content
                    await nextTick()
                    scrollToBottom()
                  }
                }
                break
                
              case 'ai_message':
                if (parsed.id) {
                  const msgIndex = messages.value.findIndex(m => m.id === aiMessageId)
                  if (msgIndex !== -1) {
                    messages.value[msgIndex].id = parsed.id
                  }
                }
                break
                
              case 'done':
                console.log('流式输出完成')
                break
                
              case 'error':
                ElMessage.error(parsed.message || '发送消息失败')
                break
            }
            
          } catch (e) {
            console.error('解析SSE数据失败:', e, 'event:', currentEvent, 'data:', data)
          }
          
          currentEvent = ''
        }
      }
    }
    
  } catch (error: any) {
    console.error('发送消息失败:', error)
    ElMessage.error(error.message || '发送消息失败')
    messages.value = messages.value.filter(m => m.id !== userMessage.id && m.id !== aiMessageId)
    inputMessage.value = userContent
  } finally {
    isLoading.value = false
  }
}

// 获取用户名首字母
function getUsernameInitial(): string {
  const username = localStorage.getItem('username') || 'U'
  return username.charAt(0).toUpperCase()
}

// 显示RAG详情
function showRagDetail(ref: any) {
  ElMessageBox.alert(
    `<div style="line-height: 1.8;">
      <p><strong>文档标题：</strong>${ref.title}</p>
      <p><strong>文档分类：</strong>${ref.category}</p>
      <p><strong>相关度：</strong>${(ref.score * 100).toFixed(0)}%</p>
      <p><strong>内容摘要：</strong></p>
      <p style="background: #f5f7fa; padding: 12px; border-radius: 4px; margin-top: 8px;">${ref.snippet}</p>
    </div>`,
    '知识库引用详情',
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关闭'
    }
  )
}

// 格式化时间
function formatTime(time: string): string {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // 一分钟内
  if (diff < 60000) {
    return '刚刚'
  }
  // 一小时内
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`
  }
  // 一天内
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`
  }
  // 一周内
  if (diff < 604800000) {
    return `${Math.floor(diff / 86400000)}天前`
  }
  // 其他
  return date.toLocaleDateString('zh-CN')
}

// 滚动到底部
function scrollToBottom() {
  if (messagesScrollbar.value) {
    messagesScrollbar.value.setScrollTop(messagesScrollbar.value.wrapRef.scrollHeight)
  }
}

// 页面加载时获取会话列表
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

/* 左侧边栏 */
.sidebar {
  width: 260px;
  height: calc(100vh - 60px);
  background: #fff;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
  overflow: hidden;
}

.sidebar .sidebar-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 60px;
}

.sidebar .sidebar-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
  font-weight: 600;
}

.sidebar .session-list {
  flex: 1;
  overflow-y: auto;
}

.sidebar .session-list::-webkit-scrollbar {
  width: 6px;
}

.sidebar .session-list::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 3px;
}

.sidebar .session-list::-webkit-scrollbar-track {
  background: #f5f7fa;
}

.sidebar .session-list .session-item {
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
  display: flex;
  flex-direction: column;
  position: relative;
  border-bottom: 1px solid #f2f3f5;
}

.sidebar .session-list .session-item:hover {
  background: #f5f7fa;
}

.sidebar .session-list .session-item.active {
  background: #ecf5ff;
}

.sidebar .session-list .session-item .session-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar .session-list .session-item .session-time {
  font-size: 12px;
  color: #909399;
}

.sidebar .session-list .session-item .delete-btn {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0;
  transition: opacity 0.2s;
}

.sidebar .session-list .session-item:hover .delete-btn {
  opacity: 1;
}

/* 右侧主区域 */
.main-content {
  flex: 1;
  height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
  overflow: hidden;
}

/* 消息容器 */
.main-content .messages-container {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.main-content .messages-container .messages {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px 20px 20px;
  width: 100%;
  box-sizing: border-box;
}

/* 对话外层容器 - 统一背景色和边框 */
.main-content .messages-container .messages .conversation-wrapper {
  width: 100%;
  max-width: 1200px;
  box-sizing: border-box;
  overflow: hidden;
  overflow-wrap: anywhere !important;
  word-wrap: break-word !important;
  word-break: break-all !important;
  background: #f8f9fa;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 用户消息（右侧） */
.main-content .messages-container .messages .conversation-wrapper .user-message {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  justify-content: flex-end;
  margin-bottom: 16px;
}

.main-content .messages-container .messages .conversation-wrapper .user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.3);
}

.main-content .messages-container .messages .conversation-wrapper .user-content {
  background: #409eff;
  padding: 12px 16px;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  line-height: 1.6;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: calc(100% - 52px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

/* AI消息（左侧） */
.main-content .messages-container .messages .conversation-wrapper .assistant-message {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
}

.main-content .messages-container .messages .conversation-wrapper .ai-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #67c23a;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 2px 4px rgba(103, 194, 58, 0.3);
}

.main-content .messages-container .messages .conversation-wrapper .message-content-wrapper {
  background: #fff;
  padding: 16px;
  border-radius: 8px;
  max-width: calc(100% - 52px);
  width: calc(100% - 52px);
  box-sizing: border-box;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  overflow-wrap: anywhere !important;
  word-wrap: break-word !important;
  word-break: break-all !important;
}

/* AI内容 */
.main-content .messages-container .messages .conversation-wrapper .ai-content {
  color: #303133;
  font-size: 15px;
  line-height: 1.8;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
}

.main-content .messages-container .messages .conversation-wrapper .ai-content .markdown-body {
  font-size: 15px;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
}

.main-content .messages-container .messages .message-item {
  margin-bottom: 0;
  width: 100%;
  max-width: 100%;
}

/* 用户消息（右侧） */
.main-content .messages-container .messages .message-item .user-message {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  justify-content: flex-end;
}

.main-content .messages-container .messages .message-item .user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.3);
}

/* AI消息（左侧） */
.main-content .messages-container .messages .message-item .assistant-message {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
}

.main-content .messages-container .messages .message-item .ai-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #67c23a;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
  box-shadow: 0 2px 4px rgba(103, 194, 58, 0.3);
}

.main-content .messages-container .messages .message-item .message-content-wrapper {
  background: #fff;
  padding: 16px;
  border-radius: 8px;
  max-width: calc(100% - 52px);
  width: calc(100% - 52px);
  box-sizing: border-box;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  overflow-wrap: anywhere !important;
  word-wrap: break-word !important;
  word-break: break-all !important;
}

.main-content .messages-container .messages .message-item .ai-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #67c23a;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 600;
flex-shrink: 0;
}

/* AI内容 */
.main-content .messages-container .messages .message-item .ai-content {
  color: #303133;
  font-size: 15px;
  line-height: 1.8;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
}

.main-content .messages-container .messages .message-item .ai-content .markdown-body {
  font-size: 15px;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
}

/* 用户内容 */
.main-content .messages-container .messages .message-item .user-content {
  background: #409eff;
  padding: 12px 16px;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  line-height: 1.6;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: calc(100% - 52px);
  width: fit-content;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.main-content .messages-container .messages .message-item .ai-content .markdown-body {
  font-size: 15px;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
}

/* 用户内容 - 暗色主题 */
.main-content .messages-container .messages .message-item .user-content {
  background: #5436da;
  padding: 12px 16px;
  border-radius: 8px;
  color: #fff;
  font-size: 14px;
  line-height: 1.6;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: calc(100% - 52px);
  width: fit-content;
  box-shadow: 0 2px 8px rgba(84, 54, 218, 0.2);
}

/* RAG引用区域 - 使用markdown引用块样式 */
.rag-references {
  margin: 16px 0;
  padding: 12px 16px;
  border-left: 4px solid #409eff;
  background: #f5f7fa;
  border-radius: 4px;
}

.rag-references .rag-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #409eff;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}

.rag-references .rag-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rag-references .rag-items .rag-item {
  padding: 8px 12px;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.rag-references .rag-items .rag-item:hover {
  background: rgba(64, 158, 255, 0.05);
}

.rag-references .rag-items .rag-item .rag-item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rag-references .rag-items .rag-item .rag-title {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  flex: 1;
}

.rag-references .rag-items .rag-item .rag-badge {
  font-size: 12px;
  color: #606266;
  background: transparent;
  padding: 0;
  border-radius: 0;
  border: none;
}

.rag-references .rag-items .rag-item .rag-snippet {
  font-size: 13px;
  color: #606266;
  margin-bottom: 4px;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.rag-references .rag-items .rag-item .rag-score-bar {
  height: 2px;
  background: #dcdfe6;
  border-radius: 1px;
  overflow: hidden;
  margin-bottom: 2px;
}

.rag-references .rag-items .rag-item .rag-score-fill {
  height: 100%;
  background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
  transition: width 0.3s ease;
}

.rag-references .rag-items .rag-item .rag-score-text {
  font-size: 11px;
  color: #909399;
  font-weight: normal;
}

/* Markdown样式 */
.markdown-body {
  color: #303133;
  font-size: 14px;
  line-height: 1.8;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: 100%;
}

.markdown-body h1,
.markdown-body h2,
.markdown-body h3,
.markdown-body h4 {
  margin: 24px 0 12px 0;
  color: #303133;
  font-weight: 600;
  line-height: 1.3;
}

.markdown-body h1 {
  font-size: 28px;
  border-bottom: 2px solid #e4e7ed;
  padding-bottom: 8px;
}

.markdown-body h2 {
  font-size: 22px;
  border-bottom: 1px solid #e4e7ed;
  padding-bottom: 6px;
}

.markdown-body h3 {
  font-size: 18px;
}

.markdown-body h4 {
  font-size: 16px;
}

.markdown-body p {
  margin: 16px 0;
}

.markdown-body ul,
.markdown-body ol {
  margin: 16px 0;
  padding-left: 24px;
}

.markdown-body li {
  margin: 8px 0;
  line-height: 1.6;
}

.markdown-body a {
  color: #409eff;
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.markdown-body a:hover {
  border-bottom-color: #409eff;
}

.markdown-body blockquote {
  border-left: 4px solid #409eff;
  padding: 12px 16px;
  margin: 16px 0;
  color: #606266;
  background: #f5f7fa;
  border-radius: 4px;
}

.markdown-body blockquote p {
  margin: 8px 0;
}

/* 代码块样式 */
.markdown-body pre.hljs {
  position: relative;
  margin: 16px 0;
  padding: 16px;
  background: #f8f8f8;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 100%;
}

.markdown-body pre.hljs code {
  font-family: 'SF Mono', 'Consolas', 'Monaco', 'Menlo', monospace;
  font-size: 14px;
  line-height: 1.6;
  color: #383a42;
  white-space: pre;
  display: block;
  overflow-x: auto;
}

/* 行内代码 - 强制换行 */
.markdown-body code:not(pre code) {
  background: linear-gradient(135deg, #fff5f0 0%, #ffe8d6 100%) !important;
  color: #e96900 !important;
  padding: 3px 6px !important;
  border-radius: 3px !important;
  font-size: 14px !important;
  font-family: 'SF Mono', 'Consolas', 'Monaco', monospace !important;
  border: 1px solid #ffd9b3 !important;
  font-weight: 500 !important;
  display: inline !important;
  max-width: 100% !important;
  white-space: pre-wrap !important;
  word-wrap: break-word !important;
  word-break: break-all !important;
  overflow-wrap: anywhere !important;
  line-height: 1.4 !important;
}

/* <wbr>标签强制生效 */
.markdown-body code:not(pre code) wbr {
  display: inline !important;
}

/* 父容器强制换行 */
.markdown-body p,
.markdown-body li {
  overflow-wrap: anywhere !important;
  word-wrap: break-word !important;
  word-break: break-all !important;
  max-width: 100% !important;
  overflow: visible !important;
}

/* AI内容 */
.main-content .messages-container .messages .message-item .ai-content {
  color: #303133;
  font-size: 15px;
  line-height: 1.8;
  word-wrap: break-word;
  overflow-wrap: break-word;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
}

.main-content .messages-container .messages .message-item .ai-content .markdown-body {
  font-size: 15px;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  word-break: break-word;
  max-width: 100%;
  width: 100%;
}

/* 表格优化 - 支持横向滚动 */
.markdown-body table {
  border-collapse: collapse;
  margin: 16px 0;
  width: 100%;
  display: block;
  overflow-x: auto;
}

.markdown-body th,
.markdown-body td {
  border: 1px solid #dcdfe6;
  padding: 12px 16px;
}

.markdown-body th {
  background: #f5f7fa;
  color: #303133;
  font-weight: 600;
  text-align: left;
}

.markdown-body td {
  color: #606266;
}

.markdown-body tr:nth-child(even) {
  background: #fafafa;
}

.markdown-body hr {
  border: none;
  height: 2px;
  background: #e4e7ed;
  margin: 24px 0;
}

.markdown-body img {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 16px 0;
}

/* 表格优化 - 支持横向滚动 */
.markdown-body table {
  display: block;
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  border-collapse: collapse;
  margin: 16px 0;
}

.markdown-body th,
.markdown-body td {
  border: 1px solid #dcdfe6;
  padding: 12px 16px;
  text-align: left;
  word-break: keep-all;
}

.markdown-body th {
  background: #f5f7fa;
  color: #303133;
  font-weight: 600;
}

.markdown-body td {
  color: #606266;
}

.markdown-body tr:nth-child(even) {
  background: #fafafa;
}

/* 图片样式 */
.markdown-body img {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 16px 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 加载状态 */
.loading-message {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  color: #909399;
  gap: 8px;
}

.loading-message .el-icon {
  font-size: 16px;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* 输入区域 */
/* 输入区域整体样式 */
.main-content .input-container {
  padding: 20px;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 快捷提问区域 */
.main-content .input-container .quick-questions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.main-content .input-container .quick-questions .quick-title {
  font-size: 13px;
  color: #606266;
  font-weight: 600;
  white-space: nowrap;
}

.main-content .input-container .quick-questions .quick-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.main-content .input-container .quick-questions .quick-btn {
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
  color: #409eff;
  font-size: 13px;
  border-radius: 20px;
  padding: 6px 14px;
  transition: all 0.2s;
}

.main-content .input-container .quick-questions .quick-btn:hover:not(:disabled) {
  background: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}

.main-content .input-container .quick-questions .quick-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 输入框wrapper */
.main-content .input-container .input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.main-content .input-container .input-wrapper .message-input {
  flex: 1;
}

.main-content .input-container .input-wrapper .message-input .el-textarea__inner {
  border-radius: 12px;
  padding: 14px 18px;
  font-size: 15px;
  line-height: 1.6;
  resize: none;
  border: 2px solid #dcdfe6;
  transition: border-color 0.2s;
}

.main-content .input-container .input-wrapper .message-input .el-textarea__inner:focus {
  border-color: #409eff;
}

.main-content .input-container .input-wrapper .message-input .el-textarea__inner::placeholder {
  color: #909399;
}

/* 发送按钮 */
.main-content .input-container .input-wrapper .send-btn {
  height: 48px;
  width: 100px;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.main-content .input-container .input-wrapper .send-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

/* MCP 工具调用显示样式 */
.tool-calls-section {
  margin-bottom: 15px;
  padding: 15px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-radius: 8px;
  border: 1px solid #bae6fd;
}

.tool-calls-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #0369a1;
  margin-bottom: 12px;
}

.tool-calls-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-call-item {
  padding: 12px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e0f2fe;
}

.tool-call-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.tool-server {
  color: #64748b;
  font-size: 13px;
}

.tool-args {
  margin-top: 8px;
}

/* 工具执行结果显示样式 */
.tool-result-section {
  margin-bottom: 15px;
  padding: 15px;
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  border-radius: 8px;
  border: 1px solid #fcd34d;
}

.tool-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #92400e;
  margin-bottom: 12px;
}

.tool-result-content {
  background: white;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #fde68a;
}

.tool-result-content pre {
  margin: 0;
  padding: 0;
  font-size: 13px;
  color: #57534e;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* AI 内容包装器 */
.ai-content-wrapper {
  width: 100%;
}

/* 输入操作区域 */
.input-actions {
  display: flex;
  gap: 10px;
  align-items: flex-end;
}

.tool-btn {
  height: 48px;
  border-radius: 12px;
  font-size: 14px;
  transition: all 0.2s;
}

/* MCP 工具选择器面板 */
.tool-selector-panel {
  margin-bottom: 15px;
  padding: 15px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  font-weight: 600;
  color: #303133;
}

.main-content .input-container .input-wrapper .send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>