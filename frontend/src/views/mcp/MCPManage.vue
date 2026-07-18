<template>
  <div class="mcp-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>MCP Server 管理</span>
          <el-button type="primary" size="small" @click="showAddDialog">
            <el-icon><Plus /></el-icon>
            添加 MCP Server
          </el-button>
        </div>
      </template>

      <el-table :data="servers" v-loading="loading" stripe>
        <el-table-column prop="name" label="名称" width="150">
          <template #default="{ row }">
            <div class="server-name">
              <el-icon class="server-icon"><Connection /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" />
        <el-table-column prop="url" label="URL" min-width="200">
          <template #default="{ row }">
            <el-link :href="row.url" target="_blank" type="primary">
              {{ row.url }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="auth_type" label="认证类型" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.auth_type" :type="row.auth_type === 'bearer' ? 'warning' : 'info'">
              {{ row.auth_type }}
            </el-tag>
            <el-tag v-else type="info">none</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="连接状态" width="120">
          <template #default="{ row }">
            <div class="connection-status" :class="connectionStatus[row.id]?.status">
              <el-icon v-if="connectionStatus[row.id]?.loading"><Loading /></el-icon>
              <el-icon v-else-if="connectionStatus[row.id]?.status === 'success'"><SuccessFilled /></el-icon>
              <el-icon v-else-if="connectionStatus[row.id]?.status === 'error'"><CircleCloseFilled /></el-icon>
              <el-icon v-else><WarningFilled /></el-icon>
              <span>{{ connectionStatus[row.id]?.text || '未测试' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="工具数量" width="100">
          <template #default="{ row }">
            <el-tag v-if="serverToolsCount[row.id]" type="success">
              {{ serverToolsCount[row.id] }}
            </el-tag>
            <el-tag v-else type="info">-</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              active-value="active"
              inactive-value="inactive"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_by" label="创建人" width="100" />
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="testConnection(row)" :loading="connectionStatus[row.id]?.loading">
              测试连接
            </el-button>
            <el-button size="small" type="info" @click="viewTools(row)">
              查看工具
            </el-button>
            <el-button size="small" type="primary" @click="showEditDialog(row)">
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadServers"
        @current-change="loadServers"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 MCP Server' : '添加 MCP Server'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="MCP Server 名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" placeholder="MCP Server 描述" />
        </el-form-item>
        <el-form-item label="URL" required>
          <el-input v-model="form.url" placeholder="HTTP/SSE endpoint URL，例如 http://localhost:8080/mcp" />
          <div class="form-tip">支持 HTTP/SSE 协议的 MCP Server endpoint</div>
        </el-form-item>
        <el-form-item label="认证类型">
          <el-select v-model="form.auth_type" placeholder="选择认证类型" clearable>
            <el-option label="无认证" value="" />
            <el-option label="API Key" value="api_key" />
            <el-option label="Bearer Token" value="bearer" />
            <el-option label="Basic Auth" value="basic" />
          </el-select>
          <div class="form-tip">
            Basic Auth: Jenkins 用户名和 API Token，格式为 base64 编码的用户名:token
          </div>
        </el-form-item>
        <el-form-item label="认证Token" v-if="form.auth_type">
          <el-input 
            v-model="form.auth_token" 
            type="password" 
            :placeholder="getAuthPlaceholder(form.auth_type)" 
            show-password 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 工具列表对话框 -->
    <el-dialog v-model="toolsDialogVisible" title="MCP Server 工具列表" width="800px">
      <div v-if="selectedServer" class="tools-header">
        <div class="server-info">
          <el-icon><Connection /></el-icon>
          <span>{{ selectedServer.name }}</span>
          <el-tag type="info">{{ selectedServer.url }}</el-tag>
        </div>
        <div class="tools-stats">
          <el-tag type="success">{{ tools.length }} 个工具</el-tag>
        </div>
      </div>
      
      <el-table :data="tools" v-loading="toolsLoading" stripe max-height="500">
        <el-table-column prop="name" label="工具名称" width="200">
          <template #default="{ row }">
            <div class="tool-name-cell">
              <el-icon><Operation /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="300" />
        <el-table-column label="参数" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.inputSchema?.properties" type="primary">
              {{ Object.keys(row.inputSchema.properties).length }} 个参数
            </el-tag>
            <el-tag v-else type="info">无参数</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button size="small" text @click="showToolDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 工具详情对话框 -->
    <el-dialog v-model="toolDetailVisible" title="工具详情" width="600px">
      <div v-if="currentTool" class="tool-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="工具名称">
            <el-tag type="primary">{{ currentTool.name }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="描述">
            {{ currentTool.description }}
          </el-descriptions-item>
          <el-descriptions-item label="参数类型">
            {{ currentTool.inputSchema?.type || 'object' }}
          </el-descriptions-item>
        </el-descriptions>
        
        <div v-if="currentTool.inputSchema?.properties" class="params-section">
          <h4>参数列表</h4>
          <el-table :data="getParamsList(currentTool.inputSchema)" stripe>
            <el-table-column prop="name" label="参数名" width="150" />
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="required" label="必填" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.required" type="danger" size="small">必填</el-tag>
                <el-tag v-else type="info" size="small">可选</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="200" />
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection, Loading, SuccessFilled, CircleCloseFilled, WarningFilled, Operation } from '@element-plus/icons-vue'
import { mcpApi } from '../../api'

const servers = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const submitting = ref(false)
const form = ref({
  name: '',
  description: '',
  url: '',
  auth_type: '',
  auth_token: ''
})

const toolsDialogVisible = ref(false)
const tools = ref<any[]>([])
const toolsLoading = ref(false)
const selectedServer = ref<any>(null)
const serverToolsCount = ref<Record<string, number>>({})

const connectionStatus = ref<Record<string, any>>({})

const toolDetailVisible = ref(false)
const currentTool = ref<any>(null)

const loadServers = async () => {
  loading.value = true
  try {
    const res = await mcpApi.listServers(currentPage.value, pageSize.value)
    if (res.code === 200) {
      servers.value = res.servers || []
      total.value = res.total || 0
      
      // 加载每个 server 的工具数量
      servers.value.forEach(server => {
        loadServerToolsCount(server.id)
        connectionStatus.value[server.id] = { status: 'unknown', text: '未测试', loading: false }
      })
    }
  } catch (error: any) {
    ElMessage.error('加载失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const loadServerToolsCount = async (serverId: string) => {
  try {
    const res = await mcpApi.getServerTools(serverId)
    if (res.code === 200) {
      serverToolsCount.value[serverId] = res.tools?.length || 0
    }
  } catch (error) {
    serverToolsCount.value[serverId] = 0
  }
}

const showAddDialog = () => {
  isEdit.value = false
  editId.value = ''
  form.value = { name: '', description: '', url: '', auth_type: '', auth_token: '' }
  dialogVisible.value = true
}

const showEditDialog = (row: any) => {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    name: row.name,
    description: row.description || '',
    url: row.url,
    auth_type: row.auth_type || '',
    auth_token: ''
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name || !form.value.url) {
    ElMessage.warning('请填写名称和URL')
    return
  }

  submitting.value = true
  try {
    if (isEdit.value) {
      await mcpApi.updateServer(editId.value, form.value)
      ElMessage.success('更新成功')
    } else {
      await mcpApi.createServer(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadServers()
  } catch (error: any) {
    ElMessage.error('操作失败: ' + error.message)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm('确定删除该 MCP Server?', '提示', { type: 'warning' })
    await mcpApi.deleteServer(row.id)
    ElMessage.success('删除成功')
    loadServers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const handleStatusChange = async (row: any) => {
  try {
    await mcpApi.updateServer(row.id, { status: row.status })
    ElMessage.success('状态更新成功')
  } catch (error: any) {
    ElMessage.error('状态更新失败: ' + error.message)
    row.status = row.status === 'active' ? 'inactive' : 'active'
  }
}

const testConnection = async (row: any) => {
  connectionStatus.value[row.id] = { status: 'testing', text: '测试中...', loading: true }
  
  try {
    const res = await mcpApi.testServer(row.id)
    if (res.success) {
      connectionStatus.value[row.id] = { status: 'success', text: '连接成功', loading: false }
      ElMessage.success('连接成功')
    } else {
      connectionStatus.value[row.id] = { status: 'error', text: '连接失败', loading: false }
      ElMessage.error('连接失败: ' + res.message)
    }
  } catch (error: any) {
    connectionStatus.value[row.id] = { status: 'error', text: '连接失败', loading: false }
    ElMessage.error('测试失败: ' + error.message)
  }
}

const viewTools = async (row: any) => {
  selectedServer.value = row
  toolsDialogVisible.value = true
  toolsLoading.value = true
  try {
    const res = await mcpApi.getServerTools(row.id)
    if (res.code === 200) {
      tools.value = res.tools || []
    }
  } catch (error: any) {
    ElMessage.error('获取工具失败: ' + error.message)
  } finally {
    toolsLoading.value = false
  }
}

const showToolDetail = (tool: any) => {
  currentTool.value = tool
  toolDetailVisible.value = true
}

const getParamsList = (inputSchema: any) => {
  if (!inputSchema?.properties) return []
  
  return Object.entries(inputSchema.properties).map(([name, prop]: [string, any]) => ({
    name,
    type: prop.type || 'any',
    required: inputSchema.required?.includes(name) || false,
    description: prop.description || ''
  }))
}

const getAuthPlaceholder = (authType: string): string => {
  switch (authType) {
    case 'api_key':
      return '输入 API Key'
    case 'bearer':
      return '输入 Bearer Token'
    case 'basic':
      return '输入 base64 编码的用户名:API_Token'
    default:
      return '输入认证 Token'
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

onMounted(() => {
  loadServers()
})
</script>

<style scoped>
.mcp-manage {
  padding: 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.server-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.server-icon {
  color: var(--el-color-primary);
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.connection-status.success {
  color: var(--el-color-success);
}

.connection-status.error {
  color: var(--el-color-danger);
}

.connection-status.testing {
  color: var(--el-color-primary);
}

.connection-status.unknown {
  color: var(--el-text-color-secondary);
}

.form-tip {
  margin-top: 5px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.tools-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.server-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tool-name-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tool-detail {
  padding: 10px;
}

.params-section {
  margin-top: 20px;
}

.params-section h4 {
  margin-bottom: 10px;
  color: var(--el-text-color-primary);
}
</style>