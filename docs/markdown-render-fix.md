# AI对话窗口优化问题修复指南

## 问题现象
代码块显示"plaintext 复制 [object Object]"或markdown内容显示异常

## 修复方案

### 方案1：使用后处理方式（当前实现）✓

**原理**：先用默认marked渲染，再后处理添加复制按钮

**优点**：
- Markdown内容正确渲染（标题、列表、表格等）
- 代码块正确显示（不会出现[object Object]）
- 复制按钮稳定可用

**实现**：见当前AIAssistant.vue的renderMarkdown函数

### 方案2：移除复制按钮（最简单）

如果复制按钮仍有问题，可以完全移除：

```vue
<script setup lang="ts">
import { marked } from 'marked'

// 配置marked
marked.setOptions({
  breaks: true,
  gfm: true
})

// 渲染Markdown（不添加复制按钮）
function renderMarkdown(content: string): string {
  try {
    return marked.parse(content, {
      breaks: true,
      gfm: true
    }) as string
  } catch (error) {
    console.error('Markdown渲染失败:', error)
    return content
  }
}
</script>
```

### 方案3：使用highlight.js（最专业）

需要安装依赖：
```bash
npm install highlight.js
```

实现代码：
```vue
<script setup lang="ts">
import { marked } from 'marked'
import hljs from 'highlight.js'

// 配置marked使用highlight.js
marked.setOptions({
  breaks: true,
  gfm: true,
  highlight: function(code: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      return hljs.highlight(code, { language: lang }).value
    }
    return hljs.highlightAuto(code).value
  }
})

function renderMarkdown(content: string): string {
  return marked.parse(content) as string
}
</script>
```

## 样式说明

### 当前代码块样式（深色主题）
- 背景：#282c34（类似VSCode One Dark）
- 头部：#21252b（分隔明显）
- 文字：#abb2bf（清晰易读）
- 复制按钮：悬停高亮，边界明显

### 调整样式（如果需要）

**改为浅色主题**：
```css
.code-block-wrapper {
  background: #f6f8fa;
  border: 1px solid #e1e4e8;
}

.code-block-wrapper .code-header {
  background: #f1f3f5;
}

.code-block-wrapper .code-header .code-lang {
  color: #586069;
}

.code-block-wrapper pre code {
  color: #24292e;
}
```

**调整字体大小**：
```css
.code-block-wrapper pre code {
  font-size: 13px; /* 或14px */
}
```

**调整复制按钮**：
```css
.copy-btn {
  padding: 8px 16px;
  font-size: 13px;
}
```

## 测试步骤

1. **刷新浏览器**（Ctrl+Shift+R）
2. **发送带代码的问题**：
   ```
   如何查看k8s pod日志？给出命令示例
   ```
3. **验证**：
   - ✓ Markdown标题、列表正确渲染
   - ✓ 代码块显示正确（不是[object Object]）
   - ✓ 复制按钮可用（点击显示"代码已复制")

## 常见问题

### Q1: 还是显示[object Object]
**解决**：使用方案2移除复制按钮，或方案3使用highlight.js

### Q2: 复制按钮点击无效
**解决**：
- 检查浏览器控制台是否有错误
- 确认navigator.clipboard API可用（HTTPS环境）
- 改用document.execCommand('copy')：

```typescript
(window as any).copyCode = function(codeId: string) {
  const codeElement = document.getElementById(codeId)
  if (codeElement) {
    const textarea = document.createElement('textarea')
    textarea.value = codeElement.textContent || ''
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    ElMessage.success('代码已复制')
  }
}
```

### Q3: Markdown其他内容不显示
**解决**：检查marked配置，确保breaks和gfm启用

### Q4: 代码高亮不显示
**解决**：需要使用highlight.js（方案3）或CSS样式手动高亮

## 文件位置

- 主文件：`frontend/src/views/AIAssistant.vue`
- 备选：`frontend/src/utils/markdownWithCopy.ts`

---

**当前状态**：使用后处理方式，应该能正常工作
**刷新浏览器后测试**，如果还有问题请提供具体截图或错误信息