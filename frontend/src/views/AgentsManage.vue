<template>
  <div class="agent-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>Agent 管理</span>
          <el-button type="primary" size="small" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            创建 Agent
          </el-button>
        </div>
      </template>

      <el-table :data="agents" v-loading="loading" stripe>
        <el-table-column prop="avatar" label="头像" width="80">
          <template #default="{ row }">
            <span class="agent-avatar">{{ row.avatar || '🤖' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="150">
          <template #default="{ row }">
            <div class="agent-name">
              <span>{{ row.name }}</span>
              <el-tag v-if="row.is_preset" type="success" size="small">预设</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" min-width="200" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model" label="模型" width="120" />
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="info" @click="showDetailDialog(row)">详情</el-button>
            <el-button size="small" type="primary" @click="showEditDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.is_preset">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="Agent 详情" width="800px">
      <el-descriptions :column="2" border v-if="currentAgent">
        <el-descriptions-item label="名称">{{ currentAgent.name }}</el-descriptions-item>
        <el-descriptions-item label="模型">{{ currentAgent.model }}</el-descriptions-item>
        <el-descriptions-item label="角色">{{ currentAgent.role }}</el-descriptions-item>
        <el-descriptions-item label="温度">{{ currentAgent.temperature }}</el-descriptions-item>
        <el-descriptions-item label="系统提示词" :span="2">
          <el-scrollbar max-height="400px">
            <pre style="white-space: pre-wrap; font-size: 13px;">{{ currentAgent.system_prompt }}</pre>
          </el-scrollbar>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 Agent' : '创建 Agent'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="系统提示词">
          <el-input v-model="form.system_prompt" type="textarea" :rows="5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
})

const agents = ref<any[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const detailDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref('')
const currentAgent = ref<any>(null)
const form = ref({
  name: '',
  avatar: '',
  role: '',
  category: '',
  description: '',
  system_prompt: '',
  model: 'qwen-turbo',
  temperature: 0.7,
  is_preset: false
})

const loadAgents = async () => {
  loading.value = true
  try {
    const res = await api.get('/agents/presets')
    if (res.data.code === 200) {
      agents.value = res.data.data.agents || []
    }
  } catch (error: any) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

function showDetailDialog(agent: any) {
  currentAgent.value = agent
  detailDialogVisible.value = true
}

function showCreateDialog() {
  isEdit.value = false
  editId.value = ''
  form.value = { name: '', avatar: '', role: '', category: '', description: '', system_prompt: '', model: 'qwen-turbo', temperature: 0.7, is_preset: false }
  dialogVisible.value = true
}

function showEditDialog(row: any) {
  isEdit.value = true
  editId.value = row.id
  form.value = { ...row }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    if (isEdit.value) {
      await api.put(`/agents/${editId.value}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/agents', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadAgents()
  } catch (error: any) {
    ElMessage.error('操作失败')
  }
}

const handleDelete = async (row: any) => {
  if (row.is_preset) return
  try {
    await ElMessageBox.confirm('确定删除?', '提示', { type: 'warning' })
    await api.delete(`/agents/${row.id}`)
    ElMessage.success('删除成功')
    loadAgents()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleToggleEnabled = async (row: any) => {
  try {
    await api.post(`/agents/${row.id}/toggle`)
    ElMessage.success('状态已切换')
  } catch (error: any) {
    ElMessage.error('切换失败')
    row.enabled = !row.enabled
  }
}

onMounted(() => {
  loadAgents()
})
</script>

<style scoped>
.agent-avatar {
  font-size: 24px;
}
.agent-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
