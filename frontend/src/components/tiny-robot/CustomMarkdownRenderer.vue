<template>
  <div class="markdown-content" v-html="renderedContent"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { renderMarkdownWithCopy } from '@/utils/markdownWithCopy'

interface Props {
  message: any
  contentIndex: number
}

const props = defineProps<Props>()

const renderedContent = computed(() => {
  const content = props.message.content
  
  if (typeof content === 'string') {
    return renderMarkdownWithCopy(content)
  }
  
  if (Array.isArray(content) && content[props.contentIndex]) {
    const item = content[props.contentIndex]
    if (item.type === 'text' || item.type === 'code') {
      const text = item.text || ''
      return renderMarkdownWithCopy(text)
    }
  }
  
  return ''
})
</script>

<style scoped>
.markdown-content {
  line-height: 1.6;
  color: #303133;
  font-size: 14px;
  max-width: 100%;
  word-break: break-word;
  overflow-wrap: break-word;
}

/* ========== 代码块样式优化（协调版本） ========== */
.markdown-content :deep(.code-block-wrapper) {
  position: relative;
  margin: 16px 0;
  border-radius: 10px;
  background: #f6f8fa;
  border: 1px solid #e1e4e8;
  overflow: hidden;
  transition: all 0.2s ease;
  max-width: 100%;
}

.markdown-content :deep(.code-block-wrapper:hover) {
  border-color: #c9d1d9;
  box-shadow: 0 3px 6px rgba(0, 0, 0, 0.08);
}

/* 代码块头部（协调背景色） */
.markdown-content :deep(.code-header) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: #f0f3f6;
  border-bottom: 1px solid #e1e4e8;
  color: #24292e;
}

/* 语言标签（协调样式） */
.markdown-content :deep(.code-lang) {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 4px 10px;
  background: #e1e4e8;
  border-radius: 6px;
  color: #586069;
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
}

/* 复制按钮（协调样式） */
.markdown-content :deep(.copy-btn) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid #d1d5da;
  border-radius: 6px;
  background: #fafbfc;
  color: #586069;
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 13px;
  font-weight: 500;
  outline: none;
  position: relative;
  overflow: hidden;
}

.markdown-content :deep(.copy-btn::before) {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  border-radius: 50%;
  background: rgba(3, 102, 214, 0.15);
  transform: translate(-50%, -50%);
  transition: width 0.3s, height 0.3s;
}

.markdown-content :deep(.copy-btn:hover::before) {
  width: 150px;
  height: 150px;
}

.markdown-content :deep(.copy-btn:hover) {
  background: #f3f4f6;
  color: #0366d6;
  border-color: #0366d6;
}

.markdown-content :deep(.copy-btn:active) {
  transform: scale(0.98);
}

.markdown-content :deep(.copy-btn svg) {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  transition: transform 0.15s;
}

.markdown-content :deep(.copy-btn:hover svg) {
  transform: scale(1.1);
}

.markdown-content :deep(.copy-btn span) {
  font-size: 12px;
  position: relative;
  z-index: 1;
}

/* 代码区域（协调背景） */
.markdown-content :deep(pre) {
  margin: 0;
  padding: 16px 20px;
  overflow-x: auto;
  background: #ffffff;
  scrollbar-width: thin;
  scrollbar-color: #d1d5da transparent;
  white-space: pre-wrap;
  word-break: break-word;
  max-width: 100%;
}

.markdown-content :deep(pre::-webkit-scrollbar) {
  height: 6px;
}

.markdown-content :deep(pre::-webkit-scrollbar-track) {
  background: transparent;
}

.markdown-content :deep(pre::-webkit-scrollbar-thumb) {
  background: #d1d5da;
  border-radius: 3px;
}

.markdown-content :deep(pre::-webkit-scrollbar-thumb:hover) {
  background: #c9d1d9;
}

.markdown-content :deep(code) {
  font-family: 'SF Mono', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', monospace;
  font-size: 13.5px;
  color: #24292e;
  line-height: 1.7;
  display: inline;
  white-space: pre-wrap;
  word-spacing: normal;
  word-break: break-word;
  padding: 0;
  background: transparent;
  border-radius: 0;
  max-width: 100%;
}

/* ========== 其他Markdown元素样式 ========== */

/* 标题 */
.markdown-content :deep(h1) {
  font-size: 24px;
  margin: 24px 0 16px;
  font-weight: 700;
  color: #24292e;
  line-height: 1.3;
  letter-spacing: -0.5px;
}

.markdown-content :deep(h2) {
  font-size: 20px;
  margin: 20px 0 14px;
  font-weight: 600;
  color: #24292e;
  line-height: 1.4;
  letter-spacing: -0.3px;
}

.markdown-content :deep(h3) {
  font-size: 18px;
  margin: 18px 0 12px;
  font-weight: 600;
  color: #24292e;
  line-height: 1.5;
}

/* 段落 */
.markdown-content :deep(p) {
  margin: 14px 0;
  line-height: 1.7;
  color: #24292e;
  word-break: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
}

/* 列表 */
.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 14px 0;
  padding-left: 28px;
  line-height: 1.7;
}

.markdown-content :deep(li) {
  margin: 6px 0;
  position: relative;
}

.markdown-content :deep(ul li::marker) {
  color: #0366d6;
}

.markdown-content :deep(ol li::marker) {
  color: #0366d6;
  font-weight: 600;
}

/* 链接 */
.markdown-content :deep(a) {
  color: #0366d6;
  text-decoration: none;
  font-weight: 500;
  border-bottom: 1px solid rgba(3, 102, 214, 0.2);
  transition: all 0.15s ease;
  padding-bottom: 1px;
  word-break: break-all;
  overflow-wrap: break-word;
  max-width: 100%;
}

.markdown-content :deep(a:hover) {
  color: #0366d6;
  border-bottom-color: #0366d6;
}

/* 强调 */
.markdown-content :deep(strong) {
  font-weight: 700;
  color: #24292e;
}

.markdown-content :deep(em) {
  font-style: italic;
  color: #586069;
}

/* 引用块 */
.markdown-content :deep(blockquote) {
  margin: 16px 0;
  padding: 12px 20px;
  border-left: 4px solid #d1d5da;
  background: #f6f8fa;
  border-radius: 6px;
}

.markdown-content :deep(blockquote p) {
  margin: 0;
  color: #586069;
  font-weight: 500;
}

/* 表格 */
.markdown-content :deep(table) {
  margin: 16px 0;
  border-collapse: collapse;
  width: 100%;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #e1e4e8;
  display: block;
  max-width: 100%;
  overflow-x: auto;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  border: 1px solid #e1e4e8;
  padding: 10px 14px;
  text-align: left;
  word-break: break-word;
}

.markdown-content :deep(th) {
  background: #f6f8fa;
  font-weight: 700;
  color: #24292e;
  font-size: 13px;
}

.markdown-content :deep(tr:nth-child(even) td) {
  background: #f6f8fa;
}

.markdown-content :deep(tr:hover td) {
  background: #f0f3f6;
}

/* 行内代码 */
.markdown-content :deep(code:not(pre code)) {
  padding: 3px 7px;
  background: rgba(27, 31, 35, 0.05);
  border-radius: 6px;
  font-size: 13px;
  color: #24292e;
  font-weight: 600;
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  border: 1px solid rgba(27, 31, 35, 0.1);
  word-break: break-word;
  overflow-wrap: break-word;
}

/* 分隔线 */
.markdown-content :deep(hr) {
  margin: 24px 0;
  border: none;
  height: 2px;
  background: #e1e4e8;
  border-radius: 2px;
}

/* 图片 */
.markdown-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 16px 0;
  border: 1px solid #e1e4e8;
  transition: all 0.2s ease;
}

.markdown-content :deep(img:hover) {
  border-color: #c9d1d9;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}
</style>