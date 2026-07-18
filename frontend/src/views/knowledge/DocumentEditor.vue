<template>
  <div class="document-editor-page">
    <div class="editor-header">
      <div class="header-left">
        <el-button size="small" @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <span class="header-title">{{ isEditing ? '编辑文档：' + document.title || '未命名' : '新建知识文档' }}</span>
        <el-tag v-if="document.doc_type" size="small" style="margin-left: 10px" :type="getDocTypeColor(document.doc_type)">
          {{ formatDocType(document.doc_type) }}
        </el-tag>
        <el-tag v-if="document.component" size="small" type="info" style="margin-left: 10px">
          {{ document.component }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-button size="small" @click="handlePreview" :disabled="!document.content">
          <el-icon><View /></el-icon>
          预览
        </el-button>
        <el-button size="small" type="primary" @click="handleSave" :loading="saving">
          <el-icon><Check /></el-icon>
          {{ isEditing ? '保存' : '创建' }}
        </el-button>
      </div>
    </div>

    <div class="editor-main">
      <el-row :gutter="20" style="height: 100%">
        <el-col :span="4" style="height: 100%; overflow-y: auto">
          <el-form label-width="60px" size="small">
            <el-form-item label="标题">
              <el-input v-model="document.title" placeholder="文档标题" />
            </el-form-item>
            
            <el-form-item label="文档类型">
              <el-select v-model="document.doc_type" style="width: 100%">
                <el-option label="SOP" value="sop" />
                <el-option label="FAQ" value="faq" />
                <el-option label="告警" value="alert" />
              </el-select>
            </el-form-item>

            <el-form-item label="组件">
              <el-input v-model="document.component" placeholder="组件名（如 mysql、k8s、redis）" />
            </el-form-item>
            
            <el-form-item label="标签">
              <el-input v-model="tagsInput" placeholder="逗号分隔" />
            </el-form-item>
          </el-form>
          
          <el-divider />
          
          <el-card shadow="never" class="editor-tips">
            <template #header>
              <span style="font-size: 12px">编辑器提示</span>
            </template>
            <div style="font-size: 12px; line-height: 1.6">
              <p>工具栏包含完整格式按钮</p>
              <p>支持快捷键：Ctrl+B粗体</p>
              <p>Ctrl+I斜体，Ctrl+D删除线</p>
              <p>代码块支持语言选择</p>
              <p>表格、链接、图片一键插入</p>
              <p>右侧实时预览渲染效果</p>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="20">
          <MarkdownEditor 
            v-model="document.content"
            :height="'calc(100vh - 140px)'"
            placeholder="在此输入Markdown内容，右侧实时预览..."
          />
        </el-col>
      </el-row>
    </div>

    <!-- 文档预览对话框 -->
    <el-dialog 
      v-model="previewDialogVisible" 
      title="文档预览"
      width="80%"
      class="preview-dialog"
    >
      <div class="preview-header">
        <h2>{{ document.title || '未命名文档' }}</h2>
        <el-tag :type="getDocTypeColor(document.doc_type)">
          {{ formatDocType(document.doc_type) }}
        </el-tag>
        <el-tag type="info" v-if="document.component" style="margin-left: 8px">
          {{ document.component }}
        </el-tag>
      </div>
      <el-divider />
      <div class="markdown-preview" v-html="renderMarkdown(document.content || '暂无内容')"></div>
      <template #footer>
        <el-button @click="previewDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ragApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, View, Check } from '@element-plus/icons-vue'
import { marked } from 'marked'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const router = useRouter()
const route = useRoute()
const docId = route.params.id as string | undefined

const isEditing = computed(() => !!docId)
const saving = ref(false)
const previewDialogVisible = ref(false)

const document = ref({
  title: '',
  content: '',
  doc_type: 'sop',
  component: '',
})

const tagsInput = ref('')

const loadDocument = async () => {
  if (!docId) return
  
  saving.value = true
  try {
    const res = await ragApi.getDocument(docId)
    console.log('Load document response:', res)
    
    if (res && res.document) {
      const docData = res.document
      console.log('Document data:', docData)
      
      document.value = {
        title: docData.title || '',
        content: docData.content || '',
        doc_type: docData.doc_type || 'sop',
        component: docData.component || '',
      }
      tagsInput.value = (docData.tags || []).join(', ')
      
      console.log('Loaded document:', document.value)
      console.log('Tags input:', tagsInput.value)
      
      ElMessage.success('文档加载成功')
    } else if (res && res.data) {
      const docData = res.data.document || res.data
      console.log('Document data (alternative):', docData)
      
      document.value = {
        title: docData.title || '',
        content: docData.content || '',
        doc_type: docData.doc_type || 'sop',
        component: docData.component || '',
      }
      tagsInput.value = (docData.tags || []).join(', ')
      
      ElMessage.success('文档加载成功')
    } else {
      ElMessage.error('加载文档失败: 未返回数据')
    }
  } catch (error: any) {
    console.error('Load document error:', error)
    ElMessage.error('加载文档失败: ' + (error.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

const handleSave = async () => {
  if (!document.value.title) {
    ElMessage.warning('请输入标题')
    return
  }
  if (!document.value.content) {
    ElMessage.warning('请输入内容')
    return
  }

  saving.value = true
  try {
    const tags = tagsInput.value
      .split(',')
      .map(t => t.trim())
      .filter(t => t)

    if (isEditing.value) {
      const res = await ragApi.updateDocument(docId!, {
        title: document.value.title,
        content: document.value.content,
        doc_type: document.value.doc_type,
        component: document.value.component,
        tags: tags
      })

      if (res.code === 200 || res.data) {
        ElMessage.success('更新成功')
        router.push('/knowledge-base')
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } else {
      const res = await ragApi.addDocument({
        title: document.value.title,
        content: document.value.content,
        doc_type: document.value.doc_type,
        component: document.value.component,
        tags: tags
      })

      if (res.code === 200 || res.data) {
        ElMessage.success('创建成功')
        router.push('/knowledge-base')
      } else {
        ElMessage.error(res.message || '创建失败')
      }
    }
  } catch (error: any) {
    ElMessage.error(isEditing.value ? '更新失败: ' : '创建失败: ' + (error.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

const handleBack = async () => {
  const hasChanges = document.value.title || document.value.content
  
  if (hasChanges) {
    try {
      await ElMessageBox.confirm(
        '文档未保存，确定要离开吗？',
        '确认离开',
        { confirmButtonText: '离开', cancelButtonText: '取消', type: 'warning' }
      )
      router.push('/knowledge-base')
    } catch {
      // 用户取消
    }
  } else {
    router.push('/knowledge-base')
  }
}

const handlePreview = () => {
  previewDialogVisible.value = true
}

const renderMarkdown = (content: string) => {
  if (!content) return '<p style="color: #999">暂无内容</p>'
  
  marked.setOptions({
    breaks: true,
    gfm: true
  })
  
  return marked(content) as string
}

const getDocTypeColor = (docType: string | undefined) => {
  const types: Record<string, string> = {
    sop: 'primary',
    faq: 'success',
    alert: 'danger',
  }
  return types[docType || ''] || 'info'
}

const formatDocType = (docType: string | undefined) => {
  if (!docType) return '未知'
  const labels: Record<string, string> = {
    sop: 'SOP',
    faq: 'FAQ',
    alert: '告警',
  }
  return labels[docType] || docType
}

onMounted(() => {
  if (docId) {
    loadDocument()
  }
})
</script>

<style scoped>
.document-editor-page {
  height: 100vh;
  width: 100%;
  background-color: #f5f7fa;
}

.editor-header {
  height: 60px;
  background-color: #fff;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  flex: 1;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-left: 15px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.editor-main {
  height: calc(100vh - 60px);
  padding: 0;
  overflow: hidden;
}

.editor-tips {
  margin-top: 10px;
}

.editor-tips p {
  margin: 8px 0;
  color: #666;
}
</style>