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

      <!-- 分类筛选 -->
      <div class="filter-bar" style="margin-bottom: 20px;">
        <el-select v-model="filterCategory" placeholder="按分类筛选" clearable @change="loadAgents" style="width: 200px; margin-right: 10px;">
          <el-option label="全部" value="" />
          <el-option label="告警处理" value="告警处理" />
          <el-option label="故障诊断" value="故障诊断" />
          <el-option label="日志分析" value="日志分析" />
          <el-option label="系统巡检" value="系统巡检" />
          <el-option label="变更执行" value="变更执行" />
          <el-option label="文档生成" value="文档生成" />
          <el-option label="合规检查" value="合规检查" />
          <el-option label="服务器命令" value="服务器命令" />
          <el-option label="自动巡检" value="自动巡检" />
        </el-select>
        <el-checkbox v-model="showPresets" @change="loadAgents" style="margin-right: 10px;">只显示预设</el-checkbox>
        <el-checkbox v-model="showEnabled" @change="loadAgents">只显示启用</el-checkbox>
      </div>

      <el-table :data="agents" v-loading="loading" stripe>
        <el-table-column prop="avatar" label="头像" width="80">
          <template #default="{ row }">
            <span style="font-size: 24px;">{{ row.avatar || '🤖' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="150">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 8px;">
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
        <el-table-column prop="temperature" label="温度" width="80">
          <template #default="{ row }">
            {{ row.temperature }}
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="handleToggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
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

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="loadAgents"
        @current-change="loadAgents"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="Agent 详情" width="800px">
      <el-descriptions :column="2" border v-if="currentAgent">
        <el-descriptions-item label="ID">{{ currentAgent.id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ currentAgent.name }}</el-descriptions-item>
        <el-descriptions-item label="头像">{{ currentAgent.avatar }}</el-descriptions-item>
        <el-descriptions-item label="角色">{{ currentAgent.role }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ currentAgent.category }}</el-descriptions-item>
        <el-descriptions-item label="模型">{{ currentAgent.model }}</el-descriptions-item>
        <el-descriptions-item label="温度">{{ currentAgent.temperature }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentAgent.enabled ? 'success' : 'danger'">
            {{ currentAgent.enabled ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="预设">
          <el-tag :type="currentAgent.is_preset ? 'warning' : 'info'">
            {{ currentAgent.is_preset ? '预设 Agent' : '自定义 Agent' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ currentAgent.description }}</el-descriptions-item>
        <el-descriptions-item label="系统提示词" :span="2">
          <el-scrollbar max-height="400px">
            <pre style="white-space: pre-wrap; font-size: 13px; line-height: 1.6;">{{ currentAgent.system_prompt }}</pre>
          </el-scrollbar>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="showEditDialog(currentAgent)">编辑</el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 Agent' : '创建 Agent'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="Agent 名称" />
        </el-form-item>
        <el-form-item label="头像">
          <el-input v-model="form.avatar" placeholder="Emoji 头像，如 🤖 🚨 📊" />
        </el-form-item>
        <el-form-item label="角色">
          <el-input v-model="form.role" placeholder="角色描述，如 告警分析与处理专家" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.category" placeholder="选择分类">
            <el-option label="告警处理" value="告警处理" />
            <el-option label="故障诊断" value="故障诊断" />
            <el-option label="日志分析" value="日志分析" />
            <el-option label="系统巡检" value="系统巡检" />
            <el-option label="变更执行" value="变更执行" />
            <el-option label="文档生成" value="文档生成" />
            <el-option label="合规检查" value="合规检查" />
            <el-option label="服务器命令" value="服务器命令" />
            <el-option label="自动巡检" value="自动巡检" />
            <el-option label="其他" value="其他" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="功能描述" />
        </el-form-item>
        <el-form-item label="系统提示词">
          <el-input v-model="form.system_prompt" type="textarea" :rows="5" placeholder="系统提示词，定义 Agent 的行为和能力" />
        </el-form-item>
        <el-form-item label="绑定模型">
          <el-select v-model="form.model" placeholder="选择 LLM 模型">
            <el-option label="qwen3.7-max" value="qwen3.7-max" />
            <el-option label="gpt-3.5-turbo" value="gpt-3.5-turbo" />
            <el-option label="gpt-4" value="gpt-4" />
          </el-select>
        </el-form-item>
        <el-form-item label="温度参数">
          <el-slider v-model="form.temperature" :min="0" :max="1" :step="0.1" show-input />
        </el-form-item>
        <el-form-item label="预设">
          <el-switch v-model="form.is_preset" />
          <span style="color: #999; font-size: 12px; margin-left: 10px;">预设 Agent 不可删除</span>
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
import api from '@/api'

const agents = ref<any[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const filterCategory = ref('')
const showPresets = ref(false)
const showEnabled = ref(false)

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
  model: 'qwen3.7-max',
  temperature: 0.7,
  is_preset: false
})

const loadAgents = async () => {
  loading.value = true
  try {
    let url = `/agents?page=${currentPage.value}&pageSize=${pageSize.value}`
    if (showPresets.value) {
      url = '/agents/presets'
    } else if (showEnabled.value) {
      url = '/agents/enabled'
    }

    const res = await api.get(url)
    // res 已经是 response.data，不需要再访问 .data
    if (res?.code === 200) {
      let agentList = res?.data?.agents || res?.data || []
      
      if (!Array.isArray(agentList)) {
        agentList = []
      }
      
      agents.value = agentList
      
      if (filterCategory.value && !showPresets.value && !showEnabled.value) {
        agents.value = agents.value.filter(a => a.category === filterCategory.value)
      }
      
      total.value = showPresets.value || showEnabled.value ? agents.value.length : (res?.data?.total || agents.value.length)
    }
  } catch (error: any) {
    ElMessage.error('加载失败: ' + error.message)
    agents.value = []
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
  form.value = {
    name: '',
    avatar: '🤖',
    role: '',
    category: '',
    description: '',
    system_prompt: '',
    model: 'qwen3.7-max',
    temperature: 0.7,
    is_preset: false
  }
  dialogVisible.value = true
}

function showEditDialog(row: any) {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    name: row.name,
    avatar: row.avatar,
    role: row.role,
    category: row.category,
    description: row.description,
    system_prompt: row.system_prompt,
    model: row.model,
    temperature: row.temperature,
    is_preset: row.is_preset
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name) {
    ElMessage.warning('请填写名称')
    return
  }

  try {
    if (isEdit.value) {
      const updates: any = {}
      Object.keys(form.value).forEach(key => {
        updates[key] = form.value[key as keyof typeof form.value]
      })
      await api.put(`/agents/${editId.value}`, updates)
      ElMessage.success('更新成功')
    } else {
      await api.post('/agents', form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadAgents()
  } catch (error: any) {
    ElMessage.error('操作失败: ' + error.message)
  }
}

const handleDelete = async (row: any) => {
  if (row.is_preset) {
    ElMessage.warning('预设 Agent 不能删除')
    return
  }

  try {
    await ElMessageBox.confirm('确定删除该 Agent?', '提示', { type: 'warning' })
    await api.delete(`/agents/${row.id}`)
    ElMessage.success('删除成功')
    loadAgents()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

const handleToggleEnabled = async (row: any) => {
  try {
    await api.post(`/agents/${row.id}/toggle`)
    ElMessage.success('状态已切换')
  } catch (error: any) {
    ElMessage.error('切换失败: ' + error.message)
    row.enabled = !row.enabled
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

onMounted(() => {
  loadAgents()
})
</script>

<style scoped>
.agent-manage {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
