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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑 Agent' : '创建 Agent'" width="800px">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="基本信息" name="basic">
          <el-scrollbar max-height="500px">
            <el-form :model="form" label-width="100px">
              <el-divider content-position="left">
                <el-icon><User /></el-icon>
                基础信息
              </el-divider>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="名称" required>
                    <el-input v-model="form.name" placeholder="Agent 名称" clearable>
                      <template #prefix>
                        <el-icon><UserFilled /></el-icon>
                      </template>
                    </el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="头像">
                    <el-input v-model="form.avatar" placeholder="Emoji 头像" clearable>
                      <template #prefix>
                        <span style="font-size: 16px;">{{ form.avatar || '🤖' }}</span>
                      </template>
                    </el-input>
                    <div style="margin-top: 8px;">
                      <el-tag 
                        v-for="emoji in ['🤖', '🚨', '📊', '🔧', '🔍', '⚡', '🎯', '💡']"
                        :key="emoji"
                        @click="form.avatar = emoji"
                        style="cursor: pointer; margin-right: 4px;"
                      >
                        {{ emoji }}
                      </el-tag>
                    </div>
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="角色">
                    <el-input 
                      v-model="form.role" 
                      placeholder="角色描述，如 告警分析与处理专家"
                      show-word-limit
                      maxlength="100"
                    />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="分类">
                    <el-select v-model="form.category" placeholder="选择分类" style="width: 100%;">
                      <el-option label="告警处理" value="告警处理">
                        <el-icon style="margin-right: 8px;"><Bell /></el-icon>告警处理
                      </el-option>
                      <el-option label="故障诊断" value="故障诊断">
                        <el-icon style="margin-right: 8px;"><Warning /></el-icon>故障诊断
                      </el-option>
                      <el-option label="日志分析" value="日志分析">
                        <el-icon style="margin-right: 8px;"><Document /></el-icon>日志分析
                      </el-option>
                      <el-option label="系统巡检" value="系统巡检">
                        <el-icon style="margin-right: 8px;"><Monitor /></el-icon>系统巡检
                      </el-option>
                      <el-option label="变更执行" value="变更执行">
                        <el-icon style="margin-right: 8px;"><Edit /></el-icon>变更执行
                      </el-option>
                      <el-option label="文档生成" value="文档生成">
                        <el-icon style="margin-right: 8px;"><Folder /></el-icon>文档生成
                      </el-option>
                      <el-option label="合规检查" value="合规检查">
                        <el-icon style="margin-right: 8px;"><CircleCheck /></el-icon>合规检查
                      </el-option>
                      <el-option label="服务器命令" value="服务器命令">
                        <el-icon style="margin-right: 8px;"><Promotion /></el-icon>服务器命令
                      </el-option>
                      <el-option label="自动巡检" value="自动巡检">
                        <el-icon style="margin-right: 8px;"><Refresh /></el-icon>自动巡检
                      </el-option>
                      <el-option label="其他" value="其他">
                        <el-icon style="margin-right: 8px;"><More /></el-icon>其他
                      </el-option>
                    </el-select>
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-form-item label="描述">
                <el-input 
                  v-model="form.description" 
                  type="textarea" 
                  :rows="3"
                  placeholder="详细描述 Agent 的功能和应用场景"
                  show-word-limit
                  maxlength="500"
                />
              </el-form-item>
              
              <el-divider content-position="left">
                <el-icon><Setting /></el-icon>
                模型配置
              </el-divider>
              
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="绑定模型">
                    <el-select v-model="form.model" placeholder="选择 LLM 模型" style="width: 100%;">
                      <el-option label="qwen3.7-max (推荐)" value="qwen3.7-max">
                        <div style="display: flex; justify-content: space-between; align-items: center;">
                          <span>qwen3.7-max</span>
                          <el-tag size="small" type="success">推荐</el-tag>
                        </div>
                      </el-option>
                      <el-option label="gpt-3.5-turbo" value="gpt-3.5-turbo" />
                      <el-option label="gpt-4" value="gpt-4" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="温度参数">
                    <div style="display: flex; align-items: center; gap: 10px;">
                      <el-slider 
                        v-model="form.temperature" 
                        :min="0" 
                        :max="1" 
                        :step="0.1" 
                        style="width: 200px;"
                      />
                      <el-tag type="info" size="small">
                        {{ form.temperature < 0.3 ? '精确' : form.temperature < 0.7 ? '平衡' : '创意' }}
                      </el-tag>
                    </div>
                    <div style="color: #999; font-size: 12px; margin-top: 4px;">
                      温度越高回复越有创意，温度越低回复越精确
                    </div>
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-divider content-position="left">
                <el-icon><Document /></el-icon>
                系统提示词
              </el-divider>
              
              <el-form-item label="提示词内容">
                <el-input 
                  v-model="form.system_prompt" 
                  type="textarea" 
                  :rows="8"
                  placeholder="定义 Agent 的行为、能力、角色定位和工作流程"
                  show-word-limit
                  maxlength="5000"
                />
                <div style="color: #999; font-size: 12px; margin-top: 4px;">
                  系统提示词决定了 Agent 的行为特征和专业能力
                </div>
              </el-form-item>
              
              <el-divider content-position="left">
                <el-icon><Lock /></el-icon>
                其他设置
              </el-divider>
              
              <el-form-item label="预设标记">
                <el-switch v-model="form.is_preset" />
                <el-tag :type="form.is_preset ? 'warning' : 'info'" size="small" style="margin-left: 10px;">
                  {{ form.is_preset ? '预设 Agent（不可删除）' : '自定义 Agent' }}
                </el-tag>
              </el-form-item>
            </el-form>
          </el-scrollbar>
        </el-tab-pane>
        
        <el-tab-pane label="工具挂载" name="tools">
          <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
            <template #title>
              <div style="display: flex; align-items: center; gap: 8px;">
                <el-icon><Setting /></el-icon>
                <span>工具挂载配置</span>
              </div>
            </template>
            <template #default>
              <div style="line-height: 1.6;">
                选择要挂载到该 Agent 的工具，可为每个工具配置特定参数。
                <strong>已选择 {{ Object.keys(selectedTools).filter(id => selectedTools[id]).length }} 个工具</strong>
              </div>
            </template>
          </el-alert>
          
          <div v-if="availableTools.length === 0" style="text-align: center; padding: 40px;">
            <el-empty description="暂无可用工具">
              <el-button type="primary" @click="loadTools">刷新工具列表</el-button>
            </el-empty>
          </div>
          
          <div v-else>
            <el-scrollbar max-height="450px">
              <div 
                v-for="tool in availableTools" 
                :key="tool.id"
                style="margin-bottom: 12px;"
              >
                <el-card 
                  :body-style="{ padding: '16px' }"
                  :class="selectedTools[tool.id] ? 'tool-card-selected' : 'tool-card'"
                  shadow="hover"
                >
                  <div style="display: flex; align-items: start; gap: 12px;">
                    <el-checkbox 
                      v-model="selectedTools[tool.id]" 
                      @change="handleToolSelect(tool)"
                      size="large"
                    />
                    
                    <div style="font-size: 32px; min-width: 48px; text-align: center;">
                      {{ tool.icon || '🔧' }}
                    </div>
                    
                    <div style="flex: 1;">
                      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
                        <strong style="font-size: 16px;">{{ tool.name }}</strong>
                        <el-tag :type="tool.type === 'builtin' ? 'primary' : 'warning'" size="small">
                          {{ tool.type === 'builtin' ? '内置' : 'MCP' }}
                        </el-tag>
                        <el-tag type="info" size="small">{{ tool.category }}</el-tag>
                        <el-tag 
                          :type="getRiskLevelType(tool.risk_level)" 
                          size="small"
                        >
                          {{ tool.risk_level === 'low' ? '低风险' : tool.risk_level === 'medium' ? '中风险' : '高风险' }}
                        </el-tag>
                      </div>
                      
                      <div style="color: #666; font-size: 14px; margin-bottom: 8px;">
                        {{ tool.description }}
                      </div>
                      
                      <div v-if="selectedTools[tool.id]" style="display: flex; gap: 8px;">
                        <el-button 
                          type="primary" 
                          size="small"
                          link
                          @click="showToolConfig(tool)"
                        >
                          <el-icon><Setting /></el-icon>
                          配置参数
                        </el-button>
                        <el-tag size="small">
                          {{ toolBindings[tool.id]?.enabled ? '已启用' : '已禁用' }}
                        </el-tag>
                      </div>
                    </div>
                    
                    <el-tag 
                      v-if="tool.is_preset" 
                      type="success" 
                      size="small"
                    >
                      预设
                    </el-tag>
                  </div>
                </el-card>
              </div>
            </el-scrollbar>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 工具配置对话框 -->
    <el-dialog v-model="toolConfigVisible" title="工具配置" width="600px">
      <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
        <template #default>
          <div style="line-height: 1.6;">
            <strong>{{ currentTool?.name }}</strong> - {{ currentTool?.description }}
          </div>
        </template>
      </el-alert>
      
      <el-form label-width="120px">
        <el-form-item label="启用状态">
          <el-switch v-model="toolConfig.enabled" />
          <span style="color: #999; font-size: 12px; margin-left: 10px;">
            {{ toolConfig.enabled ? '该工具已启用' : '该工具已禁用' }}
          </span>
        </el-form-item>
        
        <el-divider content-position="left">工具参数配置</el-divider>
        
        <el-form-item label="超时时间">
          <el-input-number v-model="toolConfig.timeout" :min="5" :max="300" />
          <span style="color: #999; font-size: 12px; margin-left: 10px;">秒</span>
        </el-form-item>
        
        <div v-if="parsedSchema && parsedSchema.properties">
          <div v-for="(prop, key) in parsedSchema.properties" :key="key" style="margin-bottom: 16px;">
            <el-form-item :label="formatLabel(key, prop)">
              <div style="display: flex; align-items: center; gap: 8px;">
                <span style="color: #999; font-size: 12px;">默认值:</span>
                <el-tag size="small" type="info">{{ formatDefaultValue(key) }}</el-tag>
              </div>
              
              <div style="margin-top: 8px;">
                <component
                  :is="getInputComponent(prop)"
                  v-model="toolConfig.configFields[key]"
                  :placeholder="prop.description || '请输入'"
                  :type="prop.type === 'string' ? 'text' : undefined"
                  :min="prop.type === 'number' ? 0 : undefined"
                  :max="prop.type === 'number' ? 1000 : undefined"
                  style="width: 100%;"
                />
                
                <el-button 
                  v-if="toolConfig.configFields[key] !== undefined"
                  type="text"
                  size="small"
                  @click="resetToDefault(key)"
                  style="margin-top: 4px;"
                >
                  恢复默认值
                </el-button>
              </div>
              
              <div v-if="prop.description" style="color: #999; font-size: 12px; margin-top: 4px;">
                {{ prop.description }}
              </div>
            </el-form-item>
          </div>
        </div>
        
        <el-divider content-position="left">高级配置</el-divider>
        
        <el-form-item label="自定义配置">
          <el-input 
            v-model="toolConfig.advancedConfig" 
            type="textarea" 
            :rows="4"
            placeholder="JSON 格式的额外配置，用于覆盖未列出的参数"
          />
          <div style="color: #999; font-size: 12px; margin-top: 4px;">
            示例: {"custom_param": "value"}
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="toolConfigVisible = false">取消</el-button>
        <el-button type="primary" @click="saveToolConfig">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, User, UserFilled, Bell, Warning, Document, Monitor, Edit, 
  Folder, CircleCheck, Promotion, Refresh, More, Setting, Lock 
} from '@element-plus/icons-vue'
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
const activeTab = ref('basic')
const submitting = ref(false)

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

const availableTools = ref<any[]>([])
const toolsLoading = ref(false)
const selectedTools = ref<Record<string, boolean>>({})
const toolBindings = ref<Record<string, any>>({})

const toolConfigVisible = ref(false)
const currentTool = ref<any>(null)
const toolConfig = ref({
  enabled: true,
  timeout: 30,
  configFields: {} as Record<string, any>,
  advancedConfig: ''
})

const parsedSchema = computed(() => {
  if (!currentTool.value?.parameters_schema) return null
  try {
    return JSON.parse(currentTool.value.parameters_schema)
  } catch {
    return null
  }
})

const parsedDefaultConfig = computed(() => {
  if (!currentTool.value?.default_config) return {}
  try {
    return JSON.parse(currentTool.value.default_config)
  } catch {
    return {}
  }
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

const loadTools = async () => {
  toolsLoading.value = true
  try {
    const res = await api.get('/tools?page=1&pageSize=100')
    if (res?.code === 200) {
      availableTools.value = res?.data?.tools || []
    }
  } catch (error: any) {
    ElMessage.error('加载工具失败: ' + error.message)
    availableTools.value = []
  } finally {
    toolsLoading.value = false
  }
}

const loadAgentTools = async (agentId: string) => {
  try {
    const res = await api.get(`/agents/${agentId}/tools`)
    if (res?.code === 200) {
      const bindings = res?.data?.bindings || []
      bindings.forEach((binding: any) => {
        selectedTools.value[binding.tool_id] = true
        toolBindings.value[binding.tool_id] = {
          enabled: binding.enabled,
          config_override: typeof binding.config_override === 'string' 
            ? JSON.parse(binding.config_override) 
            : binding.config_override || {}
        }
      })
    }
  } catch (error: any) {
    console.error('加载 Agent 工具失败:', error.message)
  }
}

const handleToolSelect = (tool: any) => {
  if (selectedTools.value[tool.id]) {
    toolBindings.value[tool.id] = {
      enabled: true,
      config_override: {}
    }
  } else {
    delete toolBindings.value[tool.id]
  }
}

const showToolConfig = (tool: any) => {
  currentTool.value = tool
  const binding = toolBindings.value[tool.id] || {}
  const defaultConfig = parsedDefaultConfig.value
  
  const configFields: Record<string, any> = {}
  
  if (parsedSchema.value?.properties) {
    Object.keys(parsedSchema.value.properties).forEach(key => {
      const overrideValue = binding.config_override?.[key]
      configFields[key] = overrideValue !== undefined ? overrideValue : defaultConfig[key]
    })
  }
  
  toolConfig.value = {
    enabled: binding.enabled !== undefined ? binding.enabled : true,
    timeout: binding.config_override?.timeout || tool.execution_timeout || 30,
    configFields,
    advancedConfig: JSON.stringify(
      Object.keys(binding.config_override || {})
        .filter(k => !parsedSchema.value?.properties?.[k] && k !== 'timeout')
        .reduce((obj, k) => {
          obj[k] = binding.config_override[k]
          return obj
        }, {} as Record<string, any>),
      null, 2
    )
  }
  
  toolConfigVisible.value = true
}

const getInputComponent = (prop: any) => {
  if (prop.type === 'boolean') {
    return 'el-switch'
  } else if (prop.type === 'number' || prop.type === 'integer') {
    return 'el-input-number'
  } else if (prop.type === 'array') {
    return 'el-select'
  } else {
    return 'el-input'
  }
}

const formatLabel = (key: string, prop: any) => {
  const labelMap: Record<string, string> = {
    'allowed_commands': '允许命令',
    'allowed_hosts': '允许主机',
    'url': '服务地址',
    'kubeconfig': 'KubeConfig',
    'datasource': '数据源',
    'query': '查询语句',
    'time_range': '时间范围',
    'host': '主机地址',
    'command': '执行命令',
    'namespace': '命名空间',
    'name': '资源名称',
    'resource_type': '资源类型',
    'service': '服务名称',
    'level': '日志级别'
  }
  
  return labelMap[key] || key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' ')
}

const formatDefaultValue = (key: string) => {
  const defaultConfig = parsedDefaultConfig.value
  const value = defaultConfig[key]
  
  if (Array.isArray(value)) {
    return value.length > 3 ? `${value.slice(0, 3).join(', ')}...` : value.join(', ')
  } else if (typeof value === 'object') {
    return JSON.stringify(value)
  } else if (value === undefined || value === null) {
    return '未设置'
  }
  
  return String(value)
}

const resetToDefault = (key: string) => {
  const defaultConfig = parsedDefaultConfig.value
  toolConfig.value.configFields[key] = defaultConfig[key]
}

const saveToolConfig = () => {
  if (currentTool.value) {
    const configOverride: Record<string, any> = {
      timeout: toolConfig.value.timeout
    }
    
    Object.keys(toolConfig.value.configFields).forEach(key => {
      const value = toolConfig.value.configFields[key]
      const defaultValue = parsedDefaultConfig.value[key]
      
      if (value !== undefined && value !== defaultValue) {
        configOverride[key] = value
      }
    })
    
    if (toolConfig.value.advancedConfig) {
      try {
        const advancedObj = JSON.parse(toolConfig.value.advancedConfig)
        Object.assign(configOverride, advancedObj)
      } catch (error) {
        ElMessage.error('高级配置 JSON 格式错误')
        return
      }
    }
    
    toolBindings.value[currentTool.value.id] = {
      enabled: toolConfig.value.enabled,
      config_override: configOverride
    }
    
    toolConfigVisible.value = false
    ElMessage.success('配置已保存')
  }
}

watch(dialogVisible, async (visible) => {
  if (visible) {
    activeTab.value = 'basic'
    selectedTools.value = {}
    toolBindings.value = {}
    await loadTools()
    
    if (isEdit.value && editId.value) {
      await loadAgentTools(editId.value)
    }
  }
})

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

async function showEditDialog(row: any) {
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

  submitting.value = true
  try {
    let agentId = editId.value
    
    if (isEdit.value) {
      const updates: any = {}
      Object.keys(form.value).forEach(key => {
        updates[key] = form.value[key as keyof typeof form.value]
      })
      await api.put(`/agents/${editId.value}`, updates)
      ElMessage.success('更新成功')
    } else {
      const res = await api.post('/agents', form.value)
      if (res?.data?.id) {
        agentId = res.data.id
      }
      ElMessage.success('创建成功')
    }
    
    if (agentId) {
      const selectedToolIds = Object.keys(selectedTools.value).filter(id => selectedTools.value[id])
      
      for (const toolId of selectedToolIds) {
        const binding = toolBindings.value[toolId]
        try {
          await api.post(`/agents/${agentId}/tools/${toolId}`, {
            config_override: binding?.config_override || {}
          })
        } catch (error) {
          console.error(`绑定工具 ${toolId} 失败:`, error)
        }
      }
      
      const allToolIds = availableTools.value.map(t => t.id)
      const unselectedToolIds = allToolIds.filter(id => !selectedTools.value[id])
      
      for (const toolId of unselectedToolIds) {
        try {
          await api.delete(`/agents/${agentId}/tools/${toolId}`)
        } catch (error) {
          console.error(`解绑工具 ${toolId} 失败:`, error)
        }
      }
    }
    
    dialogVisible.value = false
    loadAgents()
  } catch (error: any) {
    ElMessage.error('操作失败: ' + error.message)
  } finally {
    submitting.value = false
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

const getRiskLevelType = (level: string) => {
  return level === 'low' ? 'success' : level === 'medium' ? 'warning' : 'danger'
}
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

.tool-card {
  border: 1px solid #e4e7ed;
  transition: all 0.3s;
}

.tool-card:hover {
  border-color: #409eff;
}

.tool-card-selected {
  border: 2px solid #409eff;
  background: #ecf5ff;
}

.tool-card-selected:hover {
  border-color: #66b1ff;
  background: #d9ecff;
}
</style>
