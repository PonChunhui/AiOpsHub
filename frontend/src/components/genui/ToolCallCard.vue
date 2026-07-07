<template>
  <el-card class="tool-call-card" shadow="hover">
    <template #header>
      <div class="tool-header">
        <el-icon><Connection /></el-icon>
        <span>调用工具: {{ toolName }}</span>
      </div>
    </template>
    <div class="tool-content">
      <div class="tool-id">
        <el-tag type="info" size="small">ID: {{ toolId }}</el-tag>
      </div>
      <div class="tool-args">
        <div class="args-label">参数:</div>
        <pre class="args-content">{{ JSON.stringify(args, null, 2) }}</pre>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { Connection } from '@element-plus/icons-vue'

interface Props {
  toolName: string
  toolId: string
  args: Record<string, any>
}

const props = defineProps<Props>()
</script>

<style scoped>
.tool-call-card {
  margin-bottom: 10px;
  border-left: 4px solid #e6a23c;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #e6a23c;
}

.tool-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-id {
  margin-bottom: 5px;
}

.tool-args {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.args-label {
  font-weight: 600;
  color: #606266;
  margin-bottom: 5px;
}

.args-content {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>