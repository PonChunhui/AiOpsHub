<template>
  <div class="tool-manage">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工具管理</span>
          <el-button type="primary" size="small" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            创建工具
          </el-button>
          <el-button type="success" size="small" @click="initPresets" :loading="initLoading">
            初始化预设工具
          </el-button>
        </div>
      </template>

      <el-table :data="tools" v-loading="loading" stripe>
        <el-table-column prop="icon" label="图标" width="80">
          <template #default="{ row }">
            <span style="font-size: 24px;">{{ row.icon || '🔧' }}</span>
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
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'builtin' ? 'primary' : 'warning'">
              {{ row.type === 'builtin' ? '内置' : 'MCP' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="risk_level" label="风险等级" width="100">
          <template #default="{ row }">
            <el-tag :type="getRiskLevelType(row.risk_level)">
              {{ row.risk_level === 'low' ? '低' : row.risk_level === 'medium' ? '中' : '高' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'danger'">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="showEditDialog(row)">编辑</el-button>
            <el-button size="small" @click="showConfigDialog(row)">配置</el-button>
            <el-button 
              size="small" 
              :type="row.enabled ? 'warning' : 'success'"
              @click="toggleEnabled(row)"
            >
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteTool(row)"
              :disabled="row.is_preset"
            >
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
        @size-change="loadTools"
        @current-change="loadTools"
        style="margin-top: 20px; justify-content: flex-end;"
      />
    </el-card>

    <el-dialog v-model="createDialogVisible" title="创建工具" width="700px">
      <el-scrollbar max-height="500px">
        <el-form :model="toolForm" label-width="120px">
          <el-divider content-position="left">
            <el-icon><Tools /></el-icon>
            基础信息
          </el-divider>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="名称" required>
                <el-input 
                  v-model="toolForm.name" 
                  placeholder="如: ssh_exec"
                  clearable
                >
                  <template #prefix>
                    <el-icon><Tools /></el-icon>
                  </template>
                </el-input>
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  工具名称应简洁明了，使用下划线分隔
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="类型">
                <el-select v-model="toolForm.type" style="width: 100%;">
                  <el-option label="内置工具" value="builtin">
                    <el-icon style="margin-right: 8px;"><Cpu /></el-icon>内置工具
                  </el-option>
                  <el-option label="MCP工具" value="mcp">
                    <el-icon style="margin-right: 8px;"><Connection /></el-icon>MCP工具
                  </el-option>
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="分类">
                <el-select v-model="toolForm.category" style="width: 100%;">
                  <el-option label="服务器操作" value="服务器操作">
                    <el-icon style="margin-right: 8px;"><Monitor /></el-icon>服务器操作
                  </el-option>
                  <el-option label="监控查询" value="监控查询">
                    <el-icon style="margin-right: 8px;"><DataLine /></el-icon>监控查询
                  </el-option>
                  <el-option label="容器管理" value="容器管理">
                    <el-icon style="margin-right: 8px;"><Box /></el-icon>容器管理
                  </el-option>
                  <el-option label="日志分析" value="日志分析">
                    <el-icon style="margin-right: 8px;"><Document /></el-icon>日志分析
                  </el-option>
                  <el-option label="CI/CD" value="CI/CD">
                    <el-icon style="margin-right: 8px;"><Refresh /></el-icon>CI/CD
                  </el-option>
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="图标">
                <el-input 
                  v-model="toolForm.icon" 
                  placeholder="Emoji图标"
                  clearable
                >
                  <template #prefix>
                    <span style="font-size: 16px;">{{ toolForm.icon || '🔧' }}</span>
                  </template>
                </el-input>
                <div style="margin-top: 8px;">
                  <el-tag 
                    v-for="emoji in ['💻', '📊', '🚢', '📝', '🔧', '🔍', '⚡', '🎯']"
                    :key="emoji"
                    @click="toolForm.icon = emoji"
                    style="cursor: pointer; margin-right: 4px;"
                    size="small"
                  >
                    {{ emoji }}
                  </el-tag>
                </div>
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="描述">
            <el-input 
              v-model="toolForm.description" 
              type="textarea" 
              rows="3"
              placeholder="详细描述工具的功能和使用场景"
              show-word-limit
              maxlength="500"
            />
          </el-form-item>
          
          <el-divider content-position="left">
            <el-icon><Warning /></el-icon>
            安全配置
          </el-divider>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="风险等级">
                <el-select v-model="toolForm.risk_level" style="width: 100%;">
                  <el-option label="低风险" value="low">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                      <span>低风险</span>
                      <el-tag type="success" size="small">安全</el-tag>
                    </div>
                  </el-option>
                  <el-option label="中风险" value="medium">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                      <span>中风险</span>
                      <el-tag type="warning" size="small">需审核</el-tag>
                    </div>
                  </el-option>
                  <el-option label="高风险" value="high">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                      <span>高风险</span>
                      <el-tag type="danger" size="small">谨慎使用</el-tag>
                    </div>
                  </el-option>
                </el-select>
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  高风险工具需管理员审批才能执行
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="超时时间">
                <el-input-number 
                  v-model="toolForm.execution_timeout" 
                  :min="10" 
                  :max="300"
                  style="width: 100%;"
                />
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  工具执行的最大等待时间（秒）
                </div>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-scrollbar>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createTool" :loading="createLoading">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editDialogVisible" title="编辑工具" width="800px">
      <el-tabs v-model="editActiveTab">
        <el-tab-pane label="基本信息" name="basic">
          <el-scrollbar max-height="400px">
            <el-form :model="editForm" label-width="120px">
              <el-divider content-position="left">基础设置</el-divider>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="图标">
                    <el-input 
                      v-model="editForm.icon" 
                      clearable
                    >
                      <template #prefix>
                        <span style="font-size: 16px;">{{ editForm.icon || '🔧' }}</span>
                      </template>
                    </el-input>
                    <div style="margin-top: 8px;">
                      <el-tag 
                        v-for="emoji in ['💻', '📊', '🚢', '📝', '🔧', '🔍', '⚡', '🎯']"
                        :key="emoji"
                        @click="editForm.icon = emoji"
                        style="cursor: pointer; margin-right: 4px;"
                        size="small"
                      >
                        {{ emoji }}
                      </el-tag>
                    </div>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="风险等级">
                    <el-select v-model="editForm.risk_level" style="width: 100%;">
                      <el-option label="低风险" value="low">
                        <el-tag type="success" size="small">安全</el-tag>
                      </el-option>
                      <el-option label="中风险" value="medium">
                        <el-tag type="warning" size="small">需审核</el-tag>
                      </el-option>
                      <el-option label="高风险" value="high">
                        <el-tag type="danger" size="small">谨慎使用</el-tag>
                      </el-option>
                    </el-select>
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-form-item label="描述">
                <el-input 
                  v-model="editForm.description" 
                  type="textarea" 
                  rows="3"
                  show-word-limit
                  maxlength="500"
                />
              </el-form-item>
              
              <el-form-item label="超时时间">
                <el-input-number 
                  v-model="editForm.execution_timeout" 
                  :min="10" 
                  :max="300"
                />
                <span style="color: #999; font-size: 12px; margin-left: 10px;">秒</span>
              </el-form-item>
              
              <el-divider content-position="left">参数定义</el-divider>
              
              <el-form-item label="参数Schema">
                <el-input 
                  v-model="editForm.parameters_schema" 
                  type="textarea" 
                  rows="8"
                  placeholder="JSON Schema格式定义工具参数"
                />
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  定义工具接受的参数结构，包括类型、描述和验证规则
                </div>
              </el-form-item>
              
              <el-form-item label="默认配置">
                <el-input 
                  v-model="editForm.default_config" 
                  type="textarea" 
                  rows="8"
                  placeholder="JSON格式的默认配置"
                />
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  设置工具的默认参数值，Agent可覆盖这些配置
                </div>
              </el-form-item>
            </el-form>
          </el-scrollbar>
        </el-tab-pane>
        
        <el-tab-pane label="参数预览" name="preview">
          <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
            根据参数定义自动生成的表单预览
          </el-alert>
          
          <el-card v-if="parsedSchema && parsedSchema.properties">
            <div v-for="(prop, key) in parsedSchema.properties" :key="key" style="margin-bottom: 16px;">
              <el-form-item :label="key">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-tag size="small">{{ prop.type }}</el-tag>
                  <span v-if="prop.description" style="color: #666;">{{ prop.description }}</span>
                </div>
                <div style="margin-top: 8px;">
                  <el-tag v-if="parsedDefaultConfig[key]" type="info" size="small">
                    默认值: {{ formatValue(parsedDefaultConfig[key]) }}
                  </el-tag>
                </div>
              </el-form-item>
            </div>
          </el-card>
          
          <el-empty v-else description="暂无参数定义" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="updateTool" :loading="editLoading">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="configDialogVisible" title="工具配置详情" width="700px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="工具名称">{{ currentTool?.name }}</el-descriptions-item>
        <el-descriptions-item label="工具类型">{{ currentTool?.type }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ currentTool?.category }}</el-descriptions-item>
        <el-descriptions-item label="风险等级">{{ currentTool?.risk_level }}</el-descriptions-item>
        <el-descriptions-item label="超时时间">{{ currentTool?.execution_timeout }}秒</el-descriptions-item>
        <el-descriptions-item label="状态">{{ currentTool?.enabled ? '启用' : '禁用' }}</el-descriptions-item>
      </el-descriptions>
      
      <div style="margin-top: 20px;">
        <h4>参数定义</h4>
        <pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; max-height: 200px; overflow: auto;">{{ formatJSON(currentTool?.parameters_schema) }}</pre>
      </div>
      
      <div style="margin-top: 20px;">
        <h4>默认配置</h4>
        <pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; max-height: 200px; overflow: auto;">{{ formatJSON(currentTool?.default_config) }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Tools, Cpu, Connection, Monitor, DataLine, Box, 
  Document, Refresh, Warning 
} from '@element-plus/icons-vue'
import axios from 'axios'

interface Tool {
  id: string
  name: string
  type: string
  category: string
  icon: string
  description: string
  parameters_schema: string
  default_config: string
  enabled: boolean
  is_preset: boolean
  risk_level: string
  execution_timeout: number
  created_at: string
  updated_at: string
}

const tools = ref<Tool[]>([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const createDialogVisible = ref(false)
const editDialogVisible = ref(false)
const configDialogVisible = ref(false)
const currentTool = ref<Tool | null>(null)
const editActiveTab = ref('basic')

const createLoading = ref(false)
const editLoading = ref(false)
const initLoading = ref(false)

const parsedSchema = computed(() => {
  if (!editForm.value.parameters_schema) return null
  try {
    return JSON.parse(editForm.value.parameters_schema)
  } catch {
    return null
  }
})

const parsedDefaultConfig = computed(() => {
  if (!editForm.value.default_config) return {}
  try {
    return JSON.parse(editForm.value.default_config)
  } catch {
    return {}
  }
})

const toolForm = ref({
  name: '',
  type: 'builtin',
  category: '',
  icon: '',
  description: '',
  risk_level: 'low',
  execution_timeout: 60,
  parameters_schema: '',
  default_config: '',
  enabled: true,
})

const editForm = ref({
  icon: '',
  description: '',
  risk_level: 'low',
  execution_timeout: 60,
  parameters_schema: '',
  default_config: '',
})

const loadTools = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('token')
    const response = await axios.get('/api/v1/tools', {
      headers: { Authorization: `Bearer ${token}` },
      params: {
        page: currentPage.value,
        pageSize: pageSize.value,
      },
    })
    
    if (response.data?.data) {
      tools.value = response.data.data.tools || []
      total.value = response.data.data.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '加载工具列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  toolForm.value = {
    name: '',
    type: 'builtin',
    category: '',
    icon: '',
    description: '',
    risk_level: 'low',
    execution_timeout: 60,
    parameters_schema: '',
    default_config: '',
    enabled: true,
  }
  createDialogVisible.value = true
}

const createTool = async () => {
  createLoading.value = true
  try {
    const token = localStorage.getItem('token')
    await axios.post('/api/v1/tools', toolForm.value, {
      headers: { Authorization: `Bearer ${token}` },
    })
    
    ElMessage.success('工具创建成功')
    createDialogVisible.value = false
    loadTools()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '创建工具失败')
  } finally {
    createLoading.value = false
  }
}

const showEditDialog = (tool: Tool) => {
  currentTool.value = tool
  editForm.value = {
    icon: tool.icon,
    description: tool.description,
    risk_level: tool.risk_level,
    execution_timeout: tool.execution_timeout,
    parameters_schema: tool.parameters_schema || '',
    default_config: tool.default_config || '',
  }
  editDialogVisible.value = true
}

const updateTool = async () => {
  editLoading.value = true
  try {
    const token = localStorage.getItem('token')
    await axios.put(`/api/v1/tools/${currentTool.value?.id}`, editForm.value, {
      headers: { Authorization: `Bearer ${token}` },
    })
    
    ElMessage.success('工具更新成功')
    editDialogVisible.value = false
    loadTools()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '更新工具失败')
  } finally {
    editLoading.value = false
  }
}

const showConfigDialog = (tool: Tool) => {
  currentTool.value = tool
  configDialogVisible.value = true
}

const toggleEnabled = async (tool: Tool) => {
  try {
    const action = tool.enabled ? '禁用' : '启用'
    await ElMessageBox.confirm(`确定要${action}工具 "${tool.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    const token = localStorage.getItem('token')
    await axios.put(`/api/v1/tools/${tool.id}`, {
      enabled: !tool.enabled,
    }, {
      headers: { Authorization: `Bearer ${token}` },
    })
    
    ElMessage.success(`工具已${action}`)
    loadTools()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '操作失败')
    }
  }
}

const deleteTool = async (tool: Tool) => {
  try {
    await ElMessageBox.confirm(`确定要删除工具 "${tool.name}" 吗？此操作不可恢复。`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error',
    })
    
    const token = localStorage.getItem('token')
    await axios.delete(`/api/v1/tools/${tool.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    
    ElMessage.success('工具删除成功')
    loadTools()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '删除失败')
    }
  }
}

const initPresets = async () => {
  initLoading.value = true
  try {
    const token = localStorage.getItem('token')
    await axios.post('/api/v1/tools/init-presets', {}, {
      headers: { Authorization: `Bearer ${token}` },
    })
    
    ElMessage.success('预设工具初始化成功')
    loadTools()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '初始化失败')
  } finally {
    initLoading.value = false
  }
}

const getRiskLevelType = (level: string) => {
  return level === 'low' ? 'success' : level === 'medium' ? 'warning' : 'danger'
}

const formatJSON = (jsonStr: string | undefined) => {
  if (!jsonStr) return '{}'
  try {
    return JSON.stringify(JSON.parse(jsonStr), null, 2)
  } catch {
    return jsonStr
  }
}

const formatValue = (value: any) => {
  if (Array.isArray(value)) {
    return value.length > 3 ? `${value.slice(0, 3).join(', ')}...` : value.join(', ')
  } else if (typeof value === 'object') {
    return JSON.stringify(value)
  } else if (value === undefined || value === null) {
    return '未设置'
  }
  return String(value)
}

onMounted(() => {
  loadTools()
})
</script>

<style scoped>
.tool-manage {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

pre {
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

h4 {
  margin: 0 0 10px 0;
  font-weight: bold;
}
</style>