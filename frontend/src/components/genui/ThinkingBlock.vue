<template>
  <el-card class="thinking-block" shadow="hover">
    <template #header>
      <div class="thinking-header" @click="toggleExpand">
        <div class="header-left">
          <el-icon class="thinking-icon"><Loading /></el-icon>
          <span>{{ agentName || 'AI' }} 思考过程</span>
          <el-tag size="small" type="info">{{ contentLength }}字</el-tag>
        </div>
        <el-icon class="expand-icon" :class="{ expanded: isExpanded }">
          <ArrowDown />
        </el-icon>
      </div>
    </template>
    <div class="thinking-content" v-show="isExpanded">
      <pre>{{ content }}</pre>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Loading, ArrowDown } from '@element-plus/icons-vue'

interface Props {
  agentName?: string
  content: string
  timestamp?: number
}

const props = defineProps<Props>()

const isExpanded = ref(false)

const contentLength = computed(() => {
  return props.content.length
})

const toggleExpand = () => {
  isExpanded.value = !isExpanded.value
}
</script>

<style scoped>
.thinking-block {
  margin-bottom: 10px;
  border-left: 4px solid #409eff;
  background: #f0f7ff;
}

.thinking-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  user-select: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #409eff;
}

.thinking-icon {
  animation: spin 1s linear infinite;
}

.expand-icon {
  transition: transform 0.3s ease;
  color: #409eff;
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.thinking-content {
  padding: 12px;
  background: #ffffff;
  border-radius: 4px;
  margin-top: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.thinking-content pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-wrap: break-word;
  line-height: 1.6;
  color: #303133;
}
</style>