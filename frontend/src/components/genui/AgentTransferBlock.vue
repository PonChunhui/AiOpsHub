<template>
  <el-card class="agent-transfer-block" shadow="hover">
    <template #header>
      <div class="transfer-header">
        <el-icon><Right /></el-icon>
        <span>Agent 转换</span>
      </div>
    </template>
    <div class="transfer-content">
      <div class="transfer-path">
        <el-tag type="info" size="small">{{ fromAgent }}</el-tag>
        <el-icon class="arrow-icon"><Right /></el-icon>
        <el-tag type="success" size="small">{{ toAgent }}</el-tag>
      </div>
      <div class="transfer-reason">
        <span class="reason-label">原因:</span>
        <span class="reason-text">{{ reason }}</span>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { Right } from '@element-plus/icons-vue'

interface Props {
  fromAgent: string
  toAgent: string
  reason: string
}

const props = defineProps<Props>()
</script>

<style scoped>
.agent-transfer-block {
  margin-bottom: 10px;
  border-left: 4px solid #909399;
}

.transfer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #909399;
}

.transfer-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.transfer-path {
  display: flex;
  align-items: center;
  gap: 10px;
}

.arrow-icon {
  color: #67c23a;
}

.transfer-reason {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.reason-label {
  font-weight: 600;
  color: #606266;
}

.reason-text {
  color: #909399;
}
</style>