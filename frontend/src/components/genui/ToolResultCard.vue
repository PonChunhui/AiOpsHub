<template>
  <el-card class="tool-result-card" shadow="hover" :class="{ 'is-success': success, 'is-error': !success }">
    <template #header>
      <div class="result-header">
        <el-icon v-if="success" color="#67c23a"><SuccessFilled /></el-icon>
        <el-icon v-else color="#f56c6c"><CircleCloseFilled /></el-icon>
        <span>工具 {{ toolName }} 结果</span>
      </div>
    </template>
    <div class="result-content">
      <div class="result-status">
        <el-tag :type="success ? 'success' : 'danger'" size="small">
          {{ success ? '成功' : '失败' }}
        </el-tag>
      </div>
      <div class="result-body">
        <pre>{{ result }}</pre>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'

interface Props {
  toolName: string
  result: string
  success: boolean
}

const props = defineProps<Props>()
</script>

<style scoped>
.tool-result-card {
  margin-bottom: 10px;
  border-left: 4px solid #67c23a;
}

.tool-result-card.is-error {
  border-left: 4px solid #f56c6c;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.result-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.result-status {
  margin-bottom: 5px;
}

.result-body {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
}

.result-body pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>