<template>
  <el-collapse v-if="agentPath && agentPath.length > 0">
    <el-collapse-item title="Agent执行路径" name="path">
      <el-timeline>
        <el-timeline-item
          v-for="(step, index) in agentPath"
          :key="index"
          :timestamp="formatTimestamp(step.timestamp)"
          placement="top"
          :type="getStepType(step.action)"
        >
          <el-card>
            <div class="step-content">
              <el-tag :type="getStepType(step.action)" size="small">
                {{ step.agent_name }}
              </el-tag>
              <span class="step-action">{{ getActionText(step.action) }}</span>
            </div>
            <div v-if="step.agent_id" class="step-id">
              ID: {{ step.agent_id }}
            </div>
          </el-card>
        </el-timeline-item>
      </el-timeline>
    </el-collapse-item>
  </el-collapse>
</template>

<script setup lang="ts">
interface AgentRunStep {
  agent_id?: string
  agent_name: string
  action: string
  timestamp?: number
}

interface Props {
  agentPath: AgentRunStep[]
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
    case 'transfer': return '转换到其他Agent'
    case 'complete': return '完成'
    case 'error': return '发生错误'
    default: return action
  }
}

const formatTimestamp = (timestamp?: number): string => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString()
}
</script>

<style scoped>
.agent-path-visual {
  margin-top: 15px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 8px;
}

.step-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.step-action {
  color: #606266;
  font-size: 14px;
}

.step-id {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
}
</style>