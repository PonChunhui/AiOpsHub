<template>
  <div class="agent-visualization">
    <el-collapse v-if="agentPath && agentPath.length > 0">
      <el-collapse-item title="Agent执行路径" name="path">
        <el-timeline>
          <el-timeline-item
            v-for="(step, index) in agentPath"
            :key="index"
            :timestamp="formatTimestamp(step.timestamp)"
            :type="getStepType(step.action)"
          >
            <div class="step-content">
              <el-tag :type="getStepType(step.action)" size="small">
                {{ step.agent_name || 'Agent' }}
              </el-tag>
              <span class="step-action">{{ getActionText(step.action) }}</span>
            </div>
          </el-timeline-item>
        </el-timeline>
      </el-collapse-item>
    </el-collapse>
    
    <div v-if="events && events.length > 0" class="events-summary">
      <el-tag 
        v-for="(event, index) in getUniqueEventTypes(events)" 
        :key="index"
        :type="getEventTagType(event)"
        size="small"
        class="event-tag"
      >
        {{ getEventLabel(event) }}: {{ getEventCount(events, event) }}
      </el-tag>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface AgentPathStep {
  agent_id?: string
  agent_name?: string
  action: string
  timestamp?: number
}

interface AgentEvent {
  type: string
  agent_name?: string
  data?: any
  timestamp?: number
}

interface Props {
  agentPath?: AgentPathStep[]
  events?: AgentEvent[]
}

const props = defineProps<Props>()

const getStepType = (action: string): 'success' | 'warning' | 'info' | 'danger' | '' => {
  switch (action) {
    case 'start': return 'success'
    case 'tool_call': return 'warning'
    case 'transfer': return 'info'
    case 'complete': return 'success'
    case 'error': return 'danger'
    default: return ''
  }
}

const getActionText = (action: string): string => {
  switch (action) {
    case 'start': return '开始执行'
    case 'tool_call': return '调用工具'
    case 'transfer': return '转换Agent'
    case 'complete': return '完成'
    case 'error': return '出错'
    default: return action
  }
}

const formatTimestamp = (timestamp?: number): string => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString()
}

const getUniqueEventTypes = (events: AgentEvent[]): string[] => {
  return Array.from(new Set(events.map(e => e.type)))
}

const getEventCount = (events: AgentEvent[], type: string): number => {
  return events.filter(e => e.type === type).length
}

const getEventTagType = (type: string): 'success' | 'warning' | 'info' | 'danger' | '' => {
  switch (type) {
    case 'thinking': return 'info'
    case 'tool_call': return 'warning'
    case 'tool_result': return 'success'
    case 'agent_transfer': return 'info'
    case 'error': return 'danger'
    default: return ''
  }
}

const getEventLabel = (type: string): string => {
  switch (type) {
    case 'thinking': return '思考'
    case 'tool_call': return '工具调用'
    case 'tool_result': return '工具结果'
    case 'agent_transfer': return 'Agent转换'
    case 'content_chunk': return '内容块'
    case 'error': return '错误'
    default: return type
  }
}
</script>

<style scoped>
.agent-visualization {
  margin-top: 12px;
  padding: 8px;
  background: #f9fafb;
  border-radius: 6px;
}

.step-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.step-action {
  color: #606266;
  font-size: 13px;
}

.events-summary {
  margin-top: 8px;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.event-tag {
  margin: 0;
}
</style>