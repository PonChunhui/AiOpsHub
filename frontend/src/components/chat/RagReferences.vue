<template>
  <div class="rag-references">
    <div class="rag-header">
      <el-icon><Reading /></el-icon>
      <span>已引用 {{ references.length }} 篇知识库文档</span>
    </div>
    <div class="rag-items">
      <div 
        v-for="(ref, refIndex) in references"
        :key="refIndex"
        class="rag-item"
        @click="handleShowDetail(ref)"
      >
        <div class="rag-item-header">
          <span class="rag-title">{{ ref.title }}</span>
          <span class="rag-badge" v-if="ref.doc_type">{{ formatDocType(ref.doc_type) }}</span>
          <span class="rag-badge rag-component-badge" v-if="ref.component">{{ ref.component }}</span>
          <span class="rag-relevance-badge" :class="getRelevanceClass(ref.relevance_level)">
            {{ getRelevanceLabel(ref.relevance_level) }}
          </span>
        </div>
        <div class="rag-snippet">{{ ref.snippet }}</div>
        <div class="rag-score-bar">
          <div class="rag-score-fill" :style="{ width: (ref.score * 100) + '%', background: getRelevanceColor(ref.relevance_level) }"></div>
        </div>
        <div class="rag-score-text">相关度 {{ (ref.score * 100).toFixed(0) }}%</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Reading } from '@element-plus/icons-vue'

interface Props {
  references: any[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  showDetail: [ref: any]
}>()

const handleShowDetail = (ref: any) => {
  emit('showDetail', ref)
}

const formatDocType = (docType: string) => {
  const labels: Record<string, string> = {
    sop: 'SOP',
    faq: 'FAQ',
    alert: '告警',
  }
  return labels[docType] || docType
}

const getRelevanceClass = (level: string) => {
  switch (level) {
    case 'high': return 'relevance-high'
    case 'medium': return 'relevance-medium'
    case 'low': return 'relevance-low'
    default: return ''
  }
}

const getRelevanceLabel = (level: string) => {
  switch (level) {
    case 'high': return '高度相关'
    case 'medium': return '中等相关'
    case 'low': return '可能相关'
    default: return ''
  }
}

const getRelevanceColor = (level: string) => {
  switch (level) {
    case 'high': return 'linear-gradient(90deg, #67c23a 0%, #85ce61 100%)'
    case 'medium': return 'linear-gradient(90deg, #409eff 0%, #66b1ff 100%)'
    case 'low': return 'linear-gradient(90deg, #909399 0%, #b4b4b4 100%)'
    default: return 'linear-gradient(90deg, #67c23a 0%, #409eff 100%)'
  }
}
</script>

<style scoped>
.rag-references {
  margin-top: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
}

.rag-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef0f0;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  color: #f56c6c;
}

.rag-items {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rag-item {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s;
}

.rag-item:hover {
  background: #ecf5ff;
  transform: translateY(-2px);
}

.rag-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.rag-title {
  font-weight: 600;
  color: #303133;
}

.rag-badge {
  padding: 2px 8px;
  background: #409eff;
  color: white;
  border-radius: 4px;
  font-size: 12px;
}

.rag-component-badge {
  background: #67c23a;
  margin-left: 8px;
}

.rag-snippet {
  color: #606266;
  font-size: 14px;
  margin-bottom: 8px;
  line-height: 1.5;
}

.rag-score-bar {
  height: 4px;
  background: #e4e7ed;
  border-radius: 2px;
  margin-bottom: 4px;
}

.rag-score-fill {
  height: 100%;
  background: linear-gradient(90deg, #67c23a 0%, #409eff 100%);
  border-radius: 2px;
  transition: width 0.3s;
}

.rag-score-text {
  color: #909399;
  font-size: 12px;
}

.rag-relevance-badge {
  padding: 2px 8px;
  font-size: 12px;
  border-radius: 4px;
  margin-left: 8px;
}

.relevance-high {
  background: #67c23a;
  color: white;
}

.relevance-medium {
  background: #409eff;
  color: white;
}

.relevance-low {
  background: #909399;
  color: white;
}
</style>