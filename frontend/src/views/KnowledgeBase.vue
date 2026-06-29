<template>
  <div class="knowledge-base-container">
    <el-card class="documents-card">
      <template #header>
        <div class="card-header">
          <span>知识库管理</span>
          <div>
            <el-button type="primary" size="small" @click="handleAddDocument">
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
        <el-form-item label="文档类型">
          <el-select v-model="filterDocType" placeholder="全部" clearable style="width: 120px">
            <el-option label="全部" value="" />
            <el-option label="SOP" value="sop" />
            <el-option label="FAQ" value="faq" />
            <el-option label="告警" value="alert" />
          </el-select>
        </el-form-item>
        <el-form-item label="组件">
          <el-input
            v-model="filterComponent"
            placeholder="组件名（如 mysql、k8s）"
            clearable
            style="width: 150px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilter">搜索</el-button>
          <el-button @click="handleSearchClear">清除</el-button>
        </el-form-item>
      </el-form>

<el-table :data="allDocuments" v-loading="docsLoading" stripe>
        <el-table-column prop="id" label="ID" width="160">
          <template #default="{ row }">
            <el-tooltip :content="row.id" placement="top" :disabled="!row.id || row.id.length <= 20">
              <span style="cursor: pointer">{{ truncateId(row.id) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="150">
          <template #default="{ row }">
            <div>{{ row.title }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="doc_type" label="文档类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getDocTypeColor(row.doc_type)">
              {{ formatDocType(row.doc_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="component" label="组件名" width="120">
          <template #default="{ row }">
            <el-tag type="info" v-if="row.component">{{ row.component }}</el-tag>
            <span v-else style="color: #999">-</span>
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
            <el-button size="small" type="primary" @click="handleEditDocument(row)">编辑</el-button>
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
        <el-descriptions-item label="文档类型">
          <el-tag :type="getDocTypeColor(currentDocument?.doc_type)">
            {{ formatDocType(currentDocument?.doc_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="组件">
          <el-tag type="info" v-if="currentDocument?.component">
            {{ currentDocument?.component }}
          </el-tag>
          <span v-else style="color: #999">-</span>
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
        <el-button type="warning" @click="handleEditFromDetail">编辑文档</el-button>
        <el-button type="primary" @click="copyContent">复制内容</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ragApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { marked } from 'marked'

const router = useRouter()
const searchQuery = ref('')
const filterDocType = ref('')
const filterComponent = ref('')
const allDocuments = ref<any[]>([])
const docsLoading = ref(false)
const detailDialogVisible = ref(false)
const currentDocument = ref<any>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const totalDocuments = ref(0)

const renderMarkdown = (content: string) => {
  if (!content) return '<p style="color: #999">暂无内容，请在左侧编辑器输入...</p>'
  
  marked.setOptions({
    breaks: true,
    gfm: true
  })
  
  return marked(content) as string
}

const loadDocuments = async () => {
  docsLoading.value = true
  try {
    const res = await ragApi.listDocuments(filterDocType.value, filterComponent.value, searchQuery.value, currentPage.value, pageSize.value)
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
  filterDocType.value = ''
  filterComponent.value = ''
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

const handleAddDocument = () => {
  router.push('/knowledge-base/edit')
}

const handleEditDocument = (doc: any) => {
  router.push(`/knowledge-base/edit/${doc.id}`)
}

const handleEditFromDetail = () => {
  if (!currentDocument.value) return
  detailDialogVisible.value = false
  handleEditDocument(currentDocument.value)
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

const truncateId = (id: string | undefined) => {
  if (!id) return '-'
  if (id.length <= 20) return id
  return id.substring(0, 8) + '...' + id.substring(id.length - 8)
}

onMounted(() => {
  loadDocuments()
})
</script>

<style scoped>
.knowledge-base-container {
  height: 100%;
  width: 100%;
}

.documents-card {
  margin-bottom: 20px;
  height: calc(100vh - 140px);
  width: 100%;
}

.documents-card :deep(.el-card__body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 15px 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 15px;
}

.documents-card :deep(.el-table) {
  flex: 1;
}

.documents-card :deep(.el-table__body-wrapper) {
  overflow-y: auto;
}

.content-preview {
  color: #666;
  font-size: 13px;
  line-height: 1.4;
}

.pagination {
  margin-top: 15px;
  margin-bottom: 0;
  display: flex;
  justify-content: center;
  padding: 10px 0;
}
</style>