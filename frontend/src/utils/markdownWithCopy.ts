// Marked代码复制扩展（备选方案）
// 如果renderer方式有问题，可以使用这个扩展

import { marked } from 'marked'
import { ElMessage } from 'element-plus'

// 全局复制函数
if (typeof window !== 'undefined') {
  (window as any).copyCodeToClipboard = function(codeId: string) {
    const codeElement = document.getElementById(codeId)
    if (codeElement) {
      const code = codeElement.textContent || ''
      navigator.clipboard.writeText(code).then(() => {
        ElMessage.success('代码已复制到剪贴板')
      }).catch(() => {
        ElMessage.error('复制失败')
      })
    }
  }
}

// marked扩展 - 添加代码复制按钮
const codeCopyExtension = {
  name: 'codeCopy',
  renderer: {
    code(code: string, language: string) {
      const validLang = language || 'plaintext'
      const codeId = `code-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
      
      // HTML转义
      const escapedCode = code
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
      
      return `<div class="code-block-wrapper">
        <div class="code-header">
          <span class="code-lang">${validLang}</span>
          <button class="copy-btn" onclick="copyCodeToClipboard('${codeId}')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
              <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12v-2zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
            </svg>
            <span>复制</span>
          </button>
        </div>
        <pre id="${codeId}"><code class="language-${validLang}">${escapedCode}</code></pre>
      </div>`
    }
  }
}

// 使用扩展
marked.use(codeCopyExtension)

export function renderMarkdownWithCopy(content: string): string {
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