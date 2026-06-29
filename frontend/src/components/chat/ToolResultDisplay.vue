<template>
  <div class="tool-result-section">
    <div class="tool-result-header">
      <el-icon><DocumentChecked /></el-icon>
      <span>工具执行结果</span>
    </div>
    <div class="tool-result-content">
      <pre>{{ result }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { DocumentChecked } from '@element-plus/icons-vue'

interface Props {
  result: string
}

defineProps<Props>()
</script>

<style scoped>
.tool-result-section {
  margin-bottom: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
}

.tool-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #f0f9ff;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  color: #67c23a;
}

.tool-result-content {
  padding: 16px;
  background: #f5f7fa;
}

.tool-result-content pre {
  margin: 0;
  padding: 12px;
  background: #fff;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>