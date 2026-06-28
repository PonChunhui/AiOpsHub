<template>
  <div>
    <el-card class="documents-card">
      <template #header>
        <div class="card-header">
          <span>知识库管理</span>
          <div>
            <el-button type="primary" size="small" @click="showAddDialog">
              <el-icon><Plus /></el-icon>
              添加文档
            </el-button>
            <el-button type="primary" size="small" @click="loadDocuments">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-form :inline="true" class="filter-form">
        <el-form-item label="搜索">
          <el-input
            v-model="searchQuery"
            placeholder="标题或内容关键词"
            clearable
            style="width: 250px"
            @clear="handleSearchClear"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="filterCategory" placeholder="全部" clearable style="width: 120px">
            <el-option label="全部" value="" />
            <el-option label="故障排查" value="troubleshooting" />
            <el-option label="优化建议" value="optimization" />
            <el-option label="最佳实践" value="best_practice" />
            <el-option label="Kubernetes" value="kubernetes" />
            <el-option label="监控告警" value="monitoring" />
            <el-option label="自动化" value="automation" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilter">搜索</el-button>
          <el-button @click="handleSearchClear">清除</el-button>
        </el-form-item>
      </el-form>

<el-table :data="allDocuments" v-loading="docsLoading" stripe>
        <el-table-column prop="id" label="ID" width="160" />
        <el-table-column prop="title" label="标题" min-width="150">
          <template #default="{ row }">
            <div>{{ row.title }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建人" width="120">
          <template #default="{ row }">
            {{ row.metadata?.created_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="更新人" width="120">
          <template #default="{ row }">
            {{ row.metadata?.updated_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.metadata?.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.metadata?.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewDocument(row)">查看</el-button>
            <el-button size="small" type="primary" @click="editDocument(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteDocument(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="totalDocuments"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <!-- 文档详情对话框 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      :title="currentDocument?.title || '文档详情'"
      width="70%"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="文档ID">
          {{ currentDocument?.id }}
        </el-descriptions-item>
        <el-descriptions-item label="分类">
          <el-tag :type="getCategoryType(currentDocument?.category)">
            {{ currentDocument?.category || '未分类' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="标签">
          <el-tag 
            v-for="tag in (currentDocument?.tags || [])" 
            :key="tag" 
            size="small"
            style="margin-right: 5px"
          >
            {{ tag }}
          </el-tag>
          <span v-if="!currentDocument?.tags || currentDocument?.tags.length === 0">无标签</span>
        </el-descriptions-item>
<el-descriptions-item label="创建时间">
            {{ formatTime(currentDocument?.created_at) }}
          </el-descriptions-item>
        </el-descriptions>

      <el-divider content-position="left">文档内容</el-divider>
      
      <div class="document-content markdown-preview" v-html="renderMarkdown(currentDocument?.content || '暂无内容')"></div>

      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="warning" @click="editFromDetail">编辑文档</el-button>
        <el-button type="primary" @click="copyContent">复制内容</el-button>
      </template>
    </el-dialog>

    <el-dialog 
      v-model="addDialogVisible" 
      :title="isEditing ? '编辑知识文档' : '添加知识文档'"
      width="90%"
      top="3vh"
      :close-on-click-modal="false"
    >
      <el-row :gutter="20">
        <el-col :span="4">
          <el-form label-width="60px" size="small">
            <el-form-item label="标题">
              <el-input v-model="newDocument.title" placeholder="文档标题" />
            </el-form-item>
            
            <el-form-item label="分类">
              <el-select v-model="newDocument.category" style="width: 100%">
                <el-option label="故障排查" value="troubleshooting" />
                <el-option label="优化建议" value="optimization" />
                <el-option label="最佳实践" value="best_practice" />
                <el-option label="Kubernetes" value="kubernetes" />
                <el-option label="监控告警" value="monitoring" />
                <el-option label="自动化" value="automation" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="标签">
              <el-input v-model="tagsInput" placeholder="逗号分隔" />
            </el-form-item>
          </el-form>
          
          <el-divider />
          
          <el-card shadow="never" class="syntax-help">
            <template #header>
              <span style="font-size: 12px">Markdown语法</span>
            </template>
            <div style="font-size: 12px; line-height: 1.6">
              <p><code># 标题</code> - 一级标题</p>
              <p><code>## 标题</code> - 二级标题</p>
              <p><code>**粗体**</code> - 粗体文本</p>
              <p><code>*斜体*</code> - 斜体文本</p>
              <p><code>`代码`</code> - 行内代码</p>
              <p><code>```代码块```</code> - 代码块</p>
              <p><code>- 列表项</code> - 无序列表</p>
              <p><code>1. 列表项</code> - 有序列表</p>
              <p><code>[链接](url)</code> - 超链接</p>
              <p><code>![图片](url)</code> - 图片</p>
              <p><code>> 引用</code> - 引用块</p>
              <p><code>---</code> - 分隔线</p>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="10">
          <div class="editor-panel">
            <div class="panel-header">
              <span>Markdown编辑器</span>
              <el-button-group size="small">
                <el-button @click="insertMarkdown('**', '**')">B</el-button>
                <el-button @click="insertMarkdown('*', '*')">I</el-button>
                <el-button @click="insertMarkdown('`', '`')">Code</el-button>
                <el-button @click="insertMarkdown('\n```\n', '\n```\n')">Block</el-button>
                <el-button @click="insertMarkdown('# ', '')">H1</el-button>
                <el-button @click="insertMarkdown('## ', '')">H2</el-button>
              </el-button-group>
            </div>
            <textarea 
              v-model="newDocument.content"
              class="markdown-textarea"
              placeholder="在此输入Markdown内容..."
              @input="updatePreview"
            ></textarea>
          </div>
        </el-col>
        
        <el-col :span="10">
          <div class="preview-panel">
            <div class="panel-header">
              <span>实时预览</span>
            </div>
            <div class="markdown-preview-container" v-html="renderMarkdown(newDocument.content)"></div>
          </div>
        </el-col>
      </el-row>
      
      <template #footer>
        <div style="text-align: center">
          <el-button @click="addDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSave" :loading="saving">
            {{ isEditing ? '保存' : '添加' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ragApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { marked } from 'marked'

const searchQuery = ref('')
const filterCategory = ref('')
const allDocuments = ref<any[]>([])
const docsLoading = ref(false)
const addDialogVisible = ref(false)
const saving = ref(false)
const isEditing = ref(false)
const editingDocId = ref('')
const detailDialogVisible = ref(false)
const currentDocument = ref<any>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const totalDocuments = ref(0)

const newDocument = ref({
  title: '',
  content: '',
  category: 'troubleshooting',
})

const tagsInput = ref('')

const filteredDocuments = computed(() => {
  return allDocuments.value
})

const renderMarkdown = (content: string) => {
  if (!content) return '<p style="color: #999">暂无内容，请在左侧编辑器输入...</p>'
  
  marked.setOptions({
    breaks: true,
    gfm: true
  })
  
  return marked(content) as string
}

const insertMarkdown = (prefix: string, suffix: string) => {
  const textarea = document.querySelector('.markdown-textarea') as HTMLTextAreaElement
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const text = newDocument.value.content
  const selectedText = text.substring(start, end)
  
  const newText = text.substring(0, start) + prefix + selectedText + suffix + text.substring(end)
  newDocument.value.content = newText
  
  setTimeout(() => {
    textarea.focus()
    textarea.setSelectionRange(start + prefix.length, start + prefix.length + selectedText.length)
  }, 0)
}

const updatePreview = () => {
}

const loadDocuments = async () => {
  docsLoading.value = true
  try {
    const res = await ragApi.listDocuments(filterCategory.value, searchQuery.value, currentPage.value, pageSize.value)
    console.log('Documents response:', res)
    if (res.code === 200 || res.data) {
      const docsData = res.data || res
      allDocuments.value = docsData.documents || []
      totalDocuments.value = docsData.total || 0
      ElMessage.success(`已加载 ${allDocuments.value.length} 个文档，总共 ${totalDocuments.value} 个`)
    } else {
      allDocuments.value = []
      totalDocuments.value = 0
    }
  } catch (error: any) {
    ElMessage.error('加载文档失败: ' + (error.message || '未知错误'))
    console.error('Load documents error:', error)
  } finally {
    docsLoading.value = false
  }
}

const handleFilter = () => {
  currentPage.value = 1
  loadDocuments()
}

const handleSearchClear = () => {
  searchQuery.value = ''
  filterCategory.value = ''
  currentPage.value = 1
  loadDocuments()
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  currentPage.value = 1
  loadDocuments()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  loadDocuments()
}

const showAddDialog = () => {
  isEditing.value = false
  editingDocId.value = ''
  newDocument.value = {
    title: '',
    content: '',
    category: 'troubleshooting',
  }
  tagsInput.value = ''
  addDialogVisible.value = true
}

const editDocument = (doc: any) => {
  isEditing.value = true
  editingDocId.value = doc.id
  newDocument.value = {
    title: doc.title || '',
    content: doc.content || '',
    category: doc.category || 'troubleshooting',
  }
  tagsInput.value = (doc.tags || []).join(', ')
  addDialogVisible.value = true
}

const editFromDetail = () => {
  if (!currentDocument.value) return
  detailDialogVisible.value = false
  editDocument(currentDocument.value)
}

const handleSave = async () => {
  if (!newDocument.value.title) {
    ElMessage.warning('请输入标题')
    return
  }
  if (!newDocument.value.content) {
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
      const res = await ragApi.updateDocument(editingDocId.value, {
        title: newDocument.value.title,
        content: newDocument.value.content,
        category: newDocument.value.category,
        tags: tags
      })

      if (res.code === 200 || res.data) {
        ElMessage.success('更新成功')
        addDialogVisible.value = false
        loadDocuments()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } else {
      const res = await ragApi.addDocument({
        title: newDocument.value.title,
        content: newDocument.value.content,
        category: newDocument.value.category,
        tags: tags
      })

      if (res.code === 200 || res.data) {
        ElMessage.success('添加成功')
        addDialogVisible.value = false
        loadDocuments()
      } else {
        ElMessage.error(res.message || '添加失败')
      }
    }
  } catch (error: any) {
    ElMessage.error(isEditing.value ? '更新失败: ' : '添加失败: ' + (error.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

const viewDocument = (doc: any) => {
  currentDocument.value = doc
  detailDialogVisible.value = true
}

const deleteDocument = async (doc: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除文档 "${doc.title || doc.id}" 吗？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    
    const res = await ragApi.deleteDocument(doc.id)
    if (res.code === 200 || res.data) {
      ElMessage.success('删除成功')
      loadDocuments()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.message || '未知错误'))
    }
  }
}

const copyContent = () => {
  if (!currentDocument.value?.content) {
    ElMessage.warning('文档内容为空')
    return
  }
  navigator.clipboard.writeText(currentDocument.value.content)
  ElMessage.success('内容已复制到剪贴板')
}

const truncateContent = (content: string, maxLength: number) => {
  if (!content) return '暂无内容'
  if (content.length <= maxLength) return content
  return content.substring(0, maxLength) + '...'
}

const formatTime = (timestamp: string | number | undefined) => {
  if (!timestamp) return '-'
  
  try {
    if (typeof timestamp === 'number') {
      if (timestamp > 10000000000) {
        return new Date(timestamp / 1000).toLocaleString()
      }
      return new Date(timestamp * 1000).toLocaleString()
    }
    
    if (typeof timestamp === 'string') {
      const parsed = new Date(timestamp)
      if (!isNaN(parsed.getTime())) {
        return parsed.toLocaleString()
      }
      
      const numTimestamp = parseInt(timestamp)
      if (!isNaN(numTimestamp)) {
        if (numTimestamp > 10000000000) {
          return new Date(numTimestamp / 1000).toLocaleString()
        }
        return new Date(numTimestamp * 1000).toLocaleString()
      }
    }
    
    return '-'
  } catch {
    return '-'
  }
}

const getCategoryType = (category: string | undefined) => {
  const types: Record<string, string> = {
    troubleshooting: 'danger',
    optimization: 'success',
    best_practice: 'warning',
    kubernetes: 'info',
    monitoring: 'primary',
    automation: '',
  }
  return types[category || ''] || 'info'
}

onMounted(() => {
  loadDocuments()
})
</script>

<style scoped>
.documents-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 15px;
}

.content-preview {
  color: #666;
  font-size: 13px;
  line-height: 1.4;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.syntax-help {
  margin-top: 10px;
}

.syntax-help p {
  margin: 8px 0;
}

.syntax-help code {
  background-color: #f5f5f5;
  padding: 2px 4px;
  border-radius: 2px;
  font-size: 11px;
}

.editor-panel, .preview-panel {
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #fff;
}

.panel-header {
  padding: 10px 15px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #ddd;
  font-weight: 500;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.markdown-textarea {
  width: 100%;
  min-height: 500px;
  padding: 15px;
  border: none;
  outline: none;
  resize: none;
  font-family: 'Courier New', Consolas, 'Monaco', monospace;
  font-size: 14px;
  line-height: 1.6;
  background-color: #fff;
}

.markdown-textarea:focus {
  background-color: #f9f9f9;
}

.markdown-preview-container {
  min-height: 500px;
  max-height: 500px;
  overflow-y: auto;
  padding: 20px;
  background-color: #fff;
}

.document-content {
  background-color: #f5f7fa;
  padding: 20px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
}

.markdown-preview, .markdown-preview-container {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
}

.markdown-preview h1, .markdown-preview-container h1 {
  font-size: 28px;
  font-weight: bold;
  margin: 20px 0 15px 0;
  border-bottom: 2px solid #eee;
  padding-bottom: 10px;
}

.markdown-preview h2, .markdown-preview-container h2 {
  font-size: 24px;
  font-weight: bold;
  margin: 18px 0 12px 0;
  border-bottom: 1px solid #eee;
  padding-bottom: 8px;
}

.markdown-preview h3, .markdown-preview-container h3 {
  font-size: 20px;
  font-weight: bold;
  margin: 15px 0 10px 0;
}

.markdown-preview h4, .markdown-preview-container h4 {
  font-size: 18px;
  font-weight: bold;
  margin: 12px 0 8px 0;
}

.markdown-preview p, .markdown-preview-container p {
  margin: 10px 0;
}

.markdown-preview ul, .markdown-preview-container ul, 
.markdown-preview ol, .markdown-preview-container ol {
  margin: 10px 0;
  padding-left: 30px;
}

.markdown-preview li, .markdown-preview-container li {
  margin: 5px 0;
}

.markdown-preview code, .markdown-preview-container code {
  background-color: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 13px;
  color: #c7254e;
}

.markdown-preview pre, .markdown-preview-container pre {
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 15px 0;
}

.markdown-preview pre code, .markdown-preview-container pre code {
  background-color: transparent;
  padding: 0;
  color: #333;
}

.markdown-preview blockquote, .markdown-preview-container blockquote {
  border-left: 4px solid #ddd;
  padding: 10px 15px;
  margin: 15px 0;
  background-color: #f9f9f9;
  color: #666;
}

.markdown-preview table, .markdown-preview-container table {
  border-collapse: collapse;
  width: 100%;
  margin: 15px 0;
}

.markdown-preview th, .markdown-preview-container th, 
.markdown-preview td, .markdown-preview-container td {
  border: 1px solid #ddd;
  padding: 8px 12px;
  text-align: left;
}

.markdown-preview th, .markdown-preview-container th {
  background-color: #f5f5f5;
  font-weight: bold;
}

.markdown-preview tr:nth-child(even), .markdown-preview-container tr:nth-child(even) {
  background-color: #f9f9f9;
}

.markdown-preview a, .markdown-preview-container a {
  color: #409EFF;
  text-decoration: none;
}

.markdown-preview a:hover, .markdown-preview-container a:hover {
  text-decoration: underline;
}

.markdown-preview strong, .markdown-preview-container strong {
  font-weight: bold;
}

.markdown-preview em, .markdown-preview-container em {
  font-style: italic;
}

.markdown-preview hr, .markdown-preview-container hr {
  border: none;
  border-top: 1px solid #eee;
  margin: 20px 0;
}
</style>