<template>
  <el-card class="error-block" shadow="hover">
    <template #header>
      <div class="error-header">
        <el-icon color="#f56c6c"><WarningFilled /></el-icon>
        <span>错误</span>
      </div>
    </template>
    <div class="error-content">
      <div class="error-code">
        <el-tag type="danger" size="small">错误码: {{ code }}</el-tag>
      </div>
      <div class="error-message">
        <pre>{{ message }}</pre>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { WarningFilled } from '@element-plus/icons-vue'

interface Props {
  message: string
  code: number
}

const props = defineProps<Props>()
</script>

<style scoped>
.error-block {
  margin-bottom: 10px;
  border-left: 4px solid #f56c6c;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #f56c6c;
}

.error-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.error-code {
  margin-bottom: 5px;
}

.error-message {
  padding: 10px;
  background: #fef0f0;
  border-radius: 4px;
}

.error-message pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #f56c6c;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>