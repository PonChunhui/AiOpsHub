<template>
  <div class="tool-calls-section">
    <div class="tool-calls-header">
      <el-icon><Setting /></el-icon>
      <span>MCP 工具调用</span>
    </div>
    <div class="tool-calls-list">
      <div 
        v-for="(toolCall, idx) in toolCalls"
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
</template>

<script setup lang="ts">
import { Setting } from '@element-plus/icons-vue'

interface Props {
  toolCalls: any[]
}

defineProps<Props>()
</script>

<style scoped>
.tool-calls-section {
  margin-bottom: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
}

.tool-calls-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #ecf5ff;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  color: #409eff;
}

.tool-calls-list {
  padding: 16px;
}

.tool-call-item {
  margin-bottom: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.tool-call-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.tool-server {
  color: #909399;
  font-size: 12px;
}

.tool-args {
  margin-top: 8px;
}
</style>