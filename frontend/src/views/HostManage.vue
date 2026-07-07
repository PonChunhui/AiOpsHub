<template>
  <div class="host-manage">
    <el-container>
      <el-aside width="300px" class="group-aside">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>主机分组</span>
              <el-button type="primary" size="small" @click="showCreateGroupDialog">
                <el-icon><Plus /></el-icon>
                添加分组
              </el-button>
            </div>
          </template>

          <el-tree
            ref="groupTreeRef"
            :data="groupTree"
            node-key="id"
            default-expand-all
            highlight-current
            @node-click="handleGroupClick"
          >
            <template #default="{ node, data }">
              <div class="tree-node">
                <span>{{ data.name }}</span>
                <div class="tree-node-actions">
                  <el-button 
                    type="primary" 
                    size="small" 
                    link
                    @click.stop="showCreateChildGroupDialog(data)"
                  >
                    <el-icon><Plus /></el-icon>
                  </el-button>
                  <el-button 
                    type="primary" 
                    size="small" 
                    link
                    @click.stop="showEditGroupDialog(data)"
                  >
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button 
                    type="danger" 
                    size="small" 
                    link
                    @click.stop="handleDeleteGroup(data)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </template>
          </el-tree>
        </el-card>
      </el-aside>

      <el-main class="host-main">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>主机列表 {{ currentGroup ? `- ${currentGroup.name}` : '' }}</span>
              <div>
                <el-button type="primary" size="small" @click="showBatchImportDialog">
                  <el-icon><Upload /></el-icon>
                  批量导入
                </el-button>
                <el-button type="danger" size="small" @click="handleBatchDelete" :disabled="selectedHosts.length === 0">
                  <el-icon><Delete /></el-icon>
                  批量删除
                </el-button>
                <el-button type="primary" size="small" @click="showCreateHostDialog">
                  <el-icon><Plus /></el-icon>
                  添加主机
                </el-button>
              </div>
            </div>
          </template>

          <el-table
            :data="hosts"
            v-loading="loading"
            stripe
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="name" label="主机名称" min-width="120" />
            <el-table-column prop="host_type" label="主机类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.host_type === 'linux' ? 'primary' : 'success'">
                  {{ row.host_type === 'linux' ? 'Linux' : 'Windows' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ip" label="主机IP" width="130" />
            <el-table-column prop="port" label="端口" width="80" />
            <el-table-column prop="username" label="用户名" width="100" />
            <el-table-column prop="auth_type" label="认证类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.auth_type === 'password' ? 'primary' : 'warning'">
                  {{ row.auth_type === 'password' ? '密码' : '密钥' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" min-width="150" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="success" @click="openTerminal(row)">
                  <el-icon><Monitor /></el-icon>
                  终端
                </el-button>
                <el-button size="small" type="info" @click="testConnection(row)">
                  测试
                </el-button>
                <el-button size="small" type="primary" @click="showEditHostDialog(row)">
                  编辑
                </el-button>
                <el-button size="small" type="danger" @click="handleDeleteHost(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top: 20px; justify-content: flex-end"
            @size-change="loadHosts"
            @current-change="loadHosts"
          />
        </el-card>
      </el-main>
    </el-container>

    <el-dialog v-model="groupDialogVisible" :title="groupDialogTitle" width="500px">
      <el-form :model="groupForm" label-width="100px">
        <el-form-item label="分组名称" required>
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="父分组">
          <el-tree-select
            v-model="groupForm.parent_id"
            :data="groupTreeSelect"
            check-strictly
            placeholder="请选择父分组"
            clearable
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="groupForm.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitGroup">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="hostDialogVisible" :title="hostDialogTitle" width="600px">
      <el-form :model="hostForm" label-width="120px">
        <el-form-item label="所属分组" required>
          <el-tree-select
            v-model="hostForm.group_id"
            :data="groupTreeSelect"
            check-strictly
            placeholder="请选择分组"
          />
        </el-form-item>
        <el-form-item label="主机名称" required>
          <el-input v-model="hostForm.name" placeholder="请输入主机名称" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="主机类型" required>
              <el-select v-model="hostForm.host_type" placeholder="请选择主机类型">
                <el-option label="Linux" value="linux" />
                <el-option label="Windows" value="windows" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="认证类型" required>
              <el-select v-model="hostForm.auth_type" placeholder="请选择认证类型">
                <el-option label="密码" value="password" />
                <el-option label="密钥" value="key" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="主机IP" required>
              <el-input v-model="hostForm.ip" placeholder="请输入主机IP" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="SSH端口" required>
              <el-input-number v-model="hostForm.port" :min="1" :max="65535" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="用户名" required>
          <el-input v-model="hostForm.username" placeholder="请输入SSH用户名" />
        </el-form-item>
        <el-form-item label="密码" v-if="hostForm.auth_type === 'password'" required>
          <el-input v-model="hostForm.password" type="password" show-password placeholder="请输入SSH密码" />
        </el-form-item>
        <el-form-item label="SSH私钥" v-if="hostForm.auth_type === 'key'" required>
          <el-input v-model="hostForm.private_key" type="textarea" :rows="5" placeholder="请输入SSH私钥内容" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="hostForm.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="hostDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitHost">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchImportDialogVisible" title="批量导入主机" width="600px">
      <el-form :model="batchImportForm" label-width="120px">
        <el-form-item label="目标分组" required>
          <el-tree-select
            v-model="batchImportForm.group_id"
            :data="groupTreeSelect"
            check-strictly
            placeholder="请选择目标分组"
          />
        </el-form-item>
        <el-form-item label="CSV文件" required>
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".csv"
            @change="handleFileChange"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                CSV格式：主机名称,主机类型,主机IP,端口,用户名,认证类型,密码/密钥,备注
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchImportDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchImport">导入</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="terminalDialogVisible" title="SSH终端" width="800px" top="5vh">
      <div class="terminal-container" ref="terminalContainer">
        <div class="terminal-output" ref="terminalOutput"></div>
      </div>
      <template #footer>
        <el-button @click="closeTerminal">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Upload, Monitor } from '@element-plus/icons-vue'
import api from '@/api'

const groupTree = ref<any[]>([])
const groupTreeSelect = computed(() => {
  return buildGroupTreeSelect(groupTree.value)
})

const hosts = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const selectedHosts = ref<any[]>([])
const currentGroup = ref<any>(null)

const groupDialogVisible = ref(false)
const groupDialogTitle = ref('')
const isEditGroup = ref(false)
const editGroupId = ref('')
const groupForm = ref({
  name: '',
  parent_id: '',
  description: ''
})

const hostDialogVisible = ref(false)
const hostDialogTitle = ref('')
const isEditHost = ref(false)
const editHostId = ref('')
const hostForm = ref({
  group_id: '',
  name: '',
  host_type: 'linux',
  ip: '',
  port: 22,
  username: '',
  auth_type: 'password',
  password: '',
  private_key: '',
  remark: ''
})

const batchImportDialogVisible = ref(false)
const batchImportForm = ref({
  group_id: '',
  file: null as any
})

const terminalDialogVisible = ref(false)
const terminalContainer = ref()
const terminalOutput = ref()
const currentTerminalHost = ref<any>(null)
const wsConnection = ref<any>(null)

const loadGroupTree = async () => {
  try {
    const res = await api.get('/host-groups')
    if (res?.code === 200) {
      groupTree.value = res?.data?.groups || []
    }
  } catch (error: any) {
    ElMessage.error('加载分组失败: ' + error.message)
  }
}

const loadHosts = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value,
      group_id: currentGroup.value?.id || ''
    }
    const res = await api.get('/hosts', { params })
    if (res?.code === 200) {
      hosts.value = res?.data?.hosts || []
      total.value = res?.data?.total || 0
    }
  } catch (error: any) {
    ElMessage.error('加载主机失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const buildGroupTreeSelect = (groups: any[]) => {
  return groups.map(group => ({
    value: group.id,
    label: group.name,
    children: group.children ? buildGroupTreeSelect(group.children) : []
  }))
}

const handleGroupClick = (data: any) => {
  currentGroup.value = data
  currentPage.value = 1
  loadHosts()
}

const showCreateGroupDialog = () => {
  isEditGroup.value = false
  editGroupId.value = ''
  groupDialogTitle.value = '添加分组'
  groupForm.value = {
    name: '',
    parent_id: '',
    description: ''
  }
  groupDialogVisible.value = true
}

const showCreateChildGroupDialog = (parentGroup: any) => {
  isEditGroup.value = false
  editGroupId.value = ''
  groupDialogTitle.value = `添加子分组 - ${parentGroup.name}`
  groupForm.value = {
    name: '',
    parent_id: parentGroup.id,
    description: ''
  }
  groupDialogVisible.value = true
}

const showEditGroupDialog = (group: any) => {
  isEditGroup.value = true
  editGroupId.value = group.id
  groupDialogTitle.value = '编辑分组'
  groupForm.value = {
    name: group.name,
    parent_id: group.parent_id || '',
    description: group.description || ''
  }
  groupDialogVisible.value = true
}

const handleSubmitGroup = async () => {
  if (!groupForm.value.name) {
    ElMessage.warning('请输入分组名称')
    return
  }

  try {
    if (isEditGroup.value) {
      await api.put(`/host-groups/${editGroupId.value}`, groupForm.value)
      ElMessage.success('分组更新成功')
    } else {
      await api.post('/host-groups', groupForm.value)
      ElMessage.success('分组创建成功')
    }
    groupDialogVisible.value = false
    loadGroupTree()
  } catch (error: any) {
    ElMessage.error('操作失败: ' + error.message)
  }
}

const handleDeleteGroup = async (group: any) => {
  try {
    const res = await api.get(`/host-groups/${group.id}/check-cascade`)
    if (res?.data?.has_children) {
      ElMessage.warning('该分组下存在子分组或主机，无法删除')
      return
    }

    await ElMessageBox.confirm('确定删除该分组?', '提示', { type: 'warning' })
    await api.delete(`/host-groups/${group.id}`)
    ElMessage.success('分组删除成功')
    loadGroupTree()
    if (currentGroup.value?.id === group.id) {
      currentGroup.value = null
      loadHosts()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const showCreateHostDialog = () => {
  isEditHost.value = false
  editHostId.value = ''
  hostDialogTitle.value = '添加主机'
  hostForm.value = {
    group_id: currentGroup.value?.id || '',
    name: '',
    host_type: 'linux',
    ip: '',
    port: 22,
    username: '',
    auth_type: 'password',
    password: '',
    private_key: '',
    remark: ''
  }
  hostDialogVisible.value = true
}

const showEditHostDialog = (host: any) => {
  isEditHost.value = true
  editHostId.value = host.id
  hostDialogTitle.value = '编辑主机'
  hostForm.value = {
    group_id: host.group_id,
    name: host.name,
    host_type: host.host_type,
    ip: host.ip,
    port: host.port,
    username: host.username,
    auth_type: host.auth_type,
    password: '',
    private_key: '',
    remark: host.remark || ''
  }
  hostDialogVisible.value = true
}

const handleSubmitHost = async () => {
  if (!hostForm.value.name || !hostForm.value.ip || !hostForm.value.username) {
    ElMessage.warning('请填写必填项')
    return
  }

  try {
    if (isEditHost.value) {
      await api.put(`/hosts/${editHostId.value}`, hostForm.value)
      ElMessage.success('主机更新成功')
    } else {
      await api.post('/hosts', hostForm.value)
      ElMessage.success('主机创建成功')
    }
    hostDialogVisible.value = false
    loadHosts()
  } catch (error: any) {
    ElMessage.error('操作失败: ' + error.message)
  }
}

const handleDeleteHost = async (host: any) => {
  try {
    await ElMessageBox.confirm('确定删除该主机?', '提示', { type: 'warning' })
    await api.delete(`/hosts/${host.id}`)
    ElMessage.success('主机删除成功')
    loadHosts()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const handleSelectionChange = (selection: any[]) => {
  selectedHosts.value = selection
}

const handleBatchDelete = async () => {
  if (selectedHosts.value.length === 0) {
    ElMessage.warning('请选择要删除的主机')
    return
  }

  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedHosts.value.length} 个主机?`, '提示', { type: 'warning' })
    const ids = selectedHosts.value.map(h => h.id)
    await api.post('/hosts/batch-delete', { ids })
    ElMessage.success('批量删除成功')
    loadHosts()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const showBatchImportDialog = () => {
  batchImportForm.value = {
    group_id: currentGroup.value?.id || '',
    file: null
  }
  batchImportDialogVisible.value = true
}

const handleFileChange = (file: any) => {
  batchImportForm.value.file = file.raw
}

const handleBatchImport = async () => {
  if (!batchImportForm.value.group_id) {
    ElMessage.warning('请选择目标分组')
    return
  }
  if (!batchImportForm.value.file) {
    ElMessage.warning('请选择CSV文件')
    return
  }

  const formData = new FormData()
  formData.append('group_id', batchImportForm.value.group_id)
  formData.append('file', batchImportForm.value.file)

  try {
    const res = await api.post('/hosts/batch-import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    if (res?.code === 200) {
      ElMessage.success(`导入成功: ${res?.data?.import_count} 个主机`)
      if (res?.data?.error_count > 0) {
        ElMessage.warning(`导入失败: ${res?.data?.error_count} 行`)
      }
      batchImportDialogVisible.value = false
      loadHosts()
    }
  } catch (error: any) {
    ElMessage.error('导入失败: ' + error.message)
  }
}

const testConnection = async (host: any) => {
  try {
    const res = await api.post(`/hosts/${host.id}/test-connection`)
    if (res?.data?.success) {
      ElMessage.success('连接测试成功')
    } else {
      ElMessage.error('连接测试失败: ' + res?.data?.message)
    }
  } catch (error: any) {
    ElMessage.error('测试失败: ' + error.message)
  }
}

const openTerminal = (host: any) => {
  currentTerminalHost.value = host
  terminalDialogVisible.value = true

  setTimeout(() => {
    connectWebSocket(host)
  }, 100)
}

const connectWebSocket = (host: any) => {
  const token = localStorage.getItem('token')
  const wsUrl = `ws://localhost:8080/ws/ssh/${host.id}?token=${token}`

  wsConnection.value = new WebSocket(wsUrl)

  wsConnection.value.onopen = () => {
    terminalOutput.value.innerHTML += '<div style="color: green;">已连接到主机: ' + host.name + '</div>'
  }

  wsConnection.value.onmessage = (event: any) => {
    const data = JSON.parse(event.data)
    if (data.type === 'data') {
      terminalOutput.value.innerHTML += '<pre>' + data.data + '</pre>'
    } else if (data.type === 'error') {
      terminalOutput.value.innerHTML += '<div style="color: red;">错误: ' + data.data + '</div>'
    }
  }

  wsConnection.value.onerror = (error: any) => {
    terminalOutput.value.innerHTML += '<div style="color: red;">WebSocket连接错误</div>'
  }

  wsConnection.value.onclose = () => {
    terminalOutput.value.innerHTML += '<div style="color: gray;">连接已关闭</div>'
  }
}

const closeTerminal = () => {
  if (wsConnection.value) {
    wsConnection.value.close()
    wsConnection.value = null
  }
  terminalDialogVisible.value = false
  terminalOutput.value.innerHTML = ''
}

const getStatusType = (status: string) => {
  return status === 'active' ? 'success' : status === 'error' ? 'danger' : 'info'
}

const getStatusText = (status: string) => {
  return status === 'active' ? '正常' : status === 'error' ? '错误' : '停用'
}

onMounted(() => {
  loadGroupTree()
  loadHosts()
})
</script>

<style scoped>
.host-manage {
  height: calc(100vh - 60px);
}

.group-aside {
  padding: 0 0 0 20px;
}

.host-main {
  padding: 0 20px 20px 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tree-node {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.tree-node-actions {
  display: flex;
  gap: 4px;
}

.terminal-container {
  width: 100%;
  height: 500px;
  background: #1e1e1e;
  color: #ffffff;
  padding: 10px;
  border-radius: 4px;
}

.terminal-output {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  white-space: pre-wrap;
  overflow-y: auto;
  height: 100%;
}

.terminal-output pre {
  margin: 0;
  white-space: pre-wrap;
}
</style>