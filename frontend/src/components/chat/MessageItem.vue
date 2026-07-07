<template>
  <div class="message-item">
    <div class="user-message">
      <div class="user-content">{{ round.userMessage.content }}</div>
      <div class="user-avatar">{{ userInitial }}</div>
    </div>
    
    <div v-if="round.aiMessage" class="assistant-message">
      <div class="ai-avatar">AI</div>
      <div class="message-content-wrapper">
        <div v-if="isLoading && !round.aiMessage.content && (!round.aiMessage.events || round.aiMessage.events.length === 0)" class="ai-content loading-message">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>AI正在思考...</span>
        </div>
        
        <div v-else class="ai-content-wrapper">
          <GenuiRenderer 
            v-if="round.aiMessage.events && round.aiMessage.events.length > 0"
            :content="convertEventsToSchemaJson(round.aiMessage.events)"
            :components="customComponents"
          />
          
          <ToolCallDisplay 
            v-if="parseToolCalls(round.aiMessage.content).length > 0"
            :tool-calls="parseToolCalls(round.aiMessage.content)"
          />
          
          <ToolResultDisplay 
            v-if="getToolResult(round.aiMessage.content || '')"
            :result="getToolResult(round.aiMessage.content || '') || ''"
          />
          
          <div class="ai-content markdown-body" v-html="renderMarkdown(cleanContent(round.aiMessage.content))"></div>
          
          <AgentPathVisual 
            v-if="round.aiMessage.agentPath && round.aiMessage.agentPath.length > 0 && !isLoading"
            :agent-path="round.aiMessage.agentPath"
          />
        </div>
        
        <RagReferences 
          v-if="round.aiMessage.rag_references && round.aiMessage.rag_references.length > 0 && !isLoading"
          :references="round.aiMessage.rag_references"
          @show-detail="handleRagDetail"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { marked } from 'marked'
import { GenuiRenderer } from '@opentiny/genui-sdk-vue'
import { convertEventsToSchemaJson } from '@/adapters/agentEventToSchemaJson'
import { customComponents } from '@/genui/customComponents'
import ToolCallDisplay from './ToolCallDisplay.vue'
import ToolResultDisplay from './ToolResultDisplay.vue'
import RagReferences from './RagReferences.vue'
import AgentPathVisual from '@/components/genui/AgentPathVisual.vue'

interface Props {
  round: any
  isLoading: boolean
  userInitial: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  showRagDetail: [ref: any]
}>()

const parseToolCalls = (content: string): any[] => {
  if (!content) return []
  
  const calls: any[] = []
  const regex = /```tool_call\n([\s\S]*?)\n```/g
  let match
  
  while ((match = regex.exec(content)) !== null) {
    try {
      const callContent = match[1] || ''
      const call = JSON.parse(callContent)
      calls.push(call)
    } catch (e) {
      console.error('Failed to parse tool call:', e)
    }
  }
  
  return calls
}

const getToolResult = (content: string): string | null => {
  if (!content) return null
  
  const regex = /```tool_result\n([\s\S]*?)\n```/g
  const match = regex.exec(content)
  
  return match && match[1] ? match[1] : null
}

const cleanContent = (content: string): string => {
  if (!content) return ''
  
  const cleaned = content
    .replace(/```tool_call\n[\s\S]*?\n```/g, '')
    .replace(/```tool_result\n[\s\S]*?\n```/g, '')
    .trim()
  
  return cleaned
}

const renderMarkdown = (content: string) => {
  if (!content) return ''
  return marked(content)
}

const handleRagDetail = (ref: any) => {
  emit('showRagDetail', ref)
}
</script>

<style scoped>
.message-item {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.user-message {
  display: flex;
  justify-content: flex-end;
  align-items: flex-start;
  gap: 12px;
}

.user-content {
  max-width: 70%;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
  word-wrap: break-word;
}

.user-avatar {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 16px;
}

.assistant-message {
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  gap: 12px;
}

.ai-avatar {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 14px;
}

/* 消息内容包装器：限制宽度防止溢出 */
.message-content-wrapper {
  max-width: 70%; /* 限制最大宽度 */
  flex: 1;
  min-width: 0; /* 关键：防止flex子元素溢出 */
  overflow: hidden; /* 防止内容溢出容器 */
}

/* 加载状态样式 */
.loading-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 12px;
  color: #909399;
}

/* AI内容包装器：设置背景和圆角 */
.ai-content-wrapper {
  background: #f5f7fa;
  border-radius: 12px;
  padding: 16px;
  /* 关键：防止内容溢出 */
  max-width: 100%;
  min-width: 0; /* 防止flex子元素溢出 */
  overflow: hidden; /* 隐藏溢出内容 */
}

/* AI内容基础样式 */
.ai-content {
  line-height: 1.6;
  color: #303133;
  /* 防止长文本溢出 */
  word-wrap: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
  overflow-x: hidden; /* 防止横向溢出 */
}

/* ===== Markdown内容样式 ===== */
/* markdown-body 容器：确保内容不溢出 */
.markdown-body {
  max-width: 100%;
  word-wrap: break-word;
  overflow-wrap: break-word;
  overflow-x: hidden; /* 防止横向溢出，超出部分隐藏 */
}

/* markdown标题样式 */
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin-top: 16px;
  margin-bottom: 12px;
  font-weight: 600;
  line-height: 1.4;
  color: #303133;
}

.markdown-body :deep(h1) { font-size: 20px; }
.markdown-body :deep(h2) { font-size: 18px; }
.markdown-body :deep(h3) { font-size: 16px; }
.markdown-body :deep(h4) { font-size: 15px; }
.markdown-body :deep(h5) { font-size: 14px; }
.markdown-body :deep(h6) { font-size: 13px; }

/* markdown段落样式 */
.markdown-body :deep(p) {
  margin-bottom: 12px;
  line-height: 1.6;
}

/* markdown列表样式 */
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin-bottom: 12px;
  padding-left: 24px;
  line-height: 1.6;
}

.markdown-body :deep(li) {
  margin-bottom: 4px;
}

/* markdown链接样式 */
.markdown-body :deep(a) {
  color: #3b82f6;
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: all 0.3s ease;
}

.markdown-body :deep(a:hover) {
  color: #1e40af;
  border-bottom-color: #1e40af;
}

/* markdown引用样式 */
.markdown-body :deep(blockquote) {
  margin: 12px 0;
  padding: 8px 16px;
  border-left: 4px solid #3b82f6;
  background: #f0f7ff;
  border-radius: 4px;
}

.markdown-body :deep(blockquote p) {
  margin-bottom: 0;
  color: #606266;
}

/* ===== 代码块样式（防止溢出） ===== */
.markdown-body :deep(pre) {
  margin: 12px 0;
  padding: 12px;
  background: #282c34;
  border-radius: 6px;
  overflow-x: auto; /* 横向溢出时显示滚动条 */
  max-width: 100%; /* 限制最大宽度 */
}

/* 代码块内联样式 */
.markdown-body :deep(pre code) {
  font-family: 'Courier New', 'Monaco', 'Consolas', monospace;
  font-size: 13px;
  color: #abb2bf;
  background: transparent;
  padding: 0;
}

/* 行内代码样式 */
.markdown-body :deep(code:not(pre code)) {
  font-family: 'Courier New', 'Monaco', 'Consolas', monospace;
  font-size: 13px;
  padding: 2px 6px;
  background: #f0f0f0;
  border-radius: 3px;
  color: #e83e8c;
}

/* ===== 表格样式（防止溢出） ===== */
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  max-width: 100%; /* 限制表格最大宽度 */
  margin: 12px 0;
  display: block; /* 使表格成为块级元素 */
  overflow-x: auto; /* 横向溢出时显示滚动条 */
}

/* 表格单元格样式 */
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #e5e7eb;
  padding: 8px 12px;
  text-align: left;
}

/* 表头样式 */
.markdown-body :deep(th) {
  background: #f9fafb;
  font-weight: 600;
  color: #303133;
}

/* 表格行样式（斑马条纹） */
.markdown-body :deep(tr:nth-child(even)) {
  background: #f9fafb;
}

/* ===== 图片样式（防止溢出） ===== */
.markdown-body :deep(img) {
  max-width: 100%; /* 图片最大宽度不超过容器 */
  height: auto; /* 高度自适应 */
  display: block;
  margin: 12px auto;
  border-radius: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* ===== 其他markdown元素样式 ===== */
/* 分隔线样式 */
.markdown-body :deep(hr) {
  margin: 16px 0;
  border: none;
  height: 2px;
  background: linear-gradient(to right, transparent, #e5e7eb, transparent);
}

/* 删除线样式 */
.markdown-body :deep(del) {
  text-decoration: line-through;
  color: #909399;
}

/* 强调样式 */
.markdown-body :deep(strong) {
  font-weight: 600;
  color: #303133;
}

/* 斜体样式 */
.markdown-body :deep(em) {
  font-style: italic;
  color: #606266;
}
</style>