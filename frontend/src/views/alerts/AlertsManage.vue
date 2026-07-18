<template>
  <div>
    <el-card class="page-card">
      <template #header>
        <div class="page-card-header">
          <h3 class="page-card-title">告警管理</h3>
          <el-button type="primary" @click="showCreateDialog">
            创建告警
          </el-button>
        </div>
      </template>
      
      <el-table v-loading="loading" :data="alerts">
        <el-table-column prop="id" label="ID" width="280" />
        <el-table-column prop="source" label="来源" width="120" />
        <el-table-column prop="severity" label="严重性" width="100">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button 
              size="small" 
              type="primary" 
              @click="analyzeAlert(row)"
            >
              {{ alertAnalysisCache.get(row.id) ? '查看分析' : '分析' }}
            </el-button>
            <el-tag 
              v-if="alertAnalysisCache.get(row.id)" 
              type="success" 
              size="small"
              style="margin-left: 5px"
            >
              已分析
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 创建告警对话框 -->
    <el-dialog v-model="dialogVisible" title="创建告警" width="500px">
      <el-form :model="alertForm" label-width="100px">
        <el-form-item label="来源">
          <el-input v-model="alertForm.source" placeholder="manual" />
        </el-form-item>
        
        <el-form-item label="严重性">
          <el-select v-model="alertForm.severity" style="width: 100%">
            <el-option label="Critical" value="critical" />
            <el-option label="Warning" value="warning" />
            <el-option label="Info" value="info" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="标题">
          <el-input v-model="alertForm.title" placeholder="告警标题" />
        </el-form-item>
        
        <el-form-item label="描述">
          <el-input 
            v-model="alertForm.description" 
            type="textarea"
            :rows="3"
            placeholder="告警详细描述"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitAlert">
          创建
        </el-button>
      </template>
    </el-dialog>
    
    <!-- 告警分析对话框 -->
    <el-dialog 
      v-model="analyzeDialogVisible" 
      title="告警AI分析" 
      width="70%"
    >
      <el-card>
        <el-descriptions :column="2" border style="margin-bottom: 20px">
          <el-descriptions-item label="告警标题">
            {{ currentAlert?.title || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="告警描述">
            {{ currentAlert?.description || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="严重性">
            <el-tag :type="getSeverityType(currentAlert?.severity)">
              {{ currentAlert?.severity || '-' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="来源">
            {{ currentAlert?.source || '-' }}
          </el-descriptions-item>
        </el-descriptions>
        
        <!-- 分析状态和进度 -->
        <el-card v-if="workflowStatus" style="margin-bottom: 20px">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="Workflow ID">
              {{ currentWorkflowId || '未启动' }}
            </el-descriptions-item>
            <el-descriptions-item label="执行状态">
              <el-tag :type="getWorkflowStatusType(workflowStatus)">
                {{ workflowStatus || '未启动' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="分析进度">
              <span v-if="workflowStatus === 'Running'" style="color: #E6A23C">
                正在分析中...
              </span>
              <span v-else-if="workflowStatus === 'Completed'" style="color: #67C23A">
                分析完成
              </span>
              <span v-else-if="workflowStatus === 'Failed'" style="color: #F56C6C">
                分析失败
              </span>
              <span v-else style="color: #909399">
                未启动
              </span>
            </el-descriptions-item>
          </el-descriptions>
          
          <el-card v-if="workflowStatus === 'Running'" style="text-align: center; padding: 20px; margin-top: 15px">
            <el-icon class="is-loading" size="30" style="margin-bottom: 10px">
              <Loading />
            </el-icon>
            <p style="margin: 0">AI正在分析告警内容，请稍候...</p>
          </el-card>
        </el-card>
        
        <!-- 分析结果 -->
        <el-card v-if="analysisResultText" style="margin-bottom: 20px">
          <template #header>
            <div class="header">
              <h4 style="margin: 0">AI分析结果</h4>
              <div>
                <el-tag type="success" size="small" style="margin-right: 10px">
                  {{ workflowStatus === 'Completed' ? '分析完成' : '历史结果' }}
                </el-tag>
                <el-tag type="info" size="small">
                  {{ analysisResultText.length }} 字
                </el-tag>
              </div>
            </div>
          </template>
          <el-input 
            v-model="analysisResultText"
            type="textarea"
            :rows="20"
            readonly
          />
        </el-card>
        
        <!-- 分析失败提示 -->
        <el-card v-if="workflowStatus === 'Failed'" style="background-color: #fef0f0; padding: 15px; margin-bottom: 20px">
          <el-alert type="error" :closable="false">
            <p style="margin: 0">分析失败，可能原因：</p>
            <p style="margin: 5px 0 0 0; font-size: 12px">
              - Agent配置错误<br>
              - LLM连接失败<br>
              - 网络超时
            </p>
          </el-alert>
        </el-card>
        
        <!-- 未分析提示 -->
        <el-card v-if="!workflowStatus && !analysisResultText" style="text-align: center; padding: 30px; margin-bottom: 20px">
          <el-icon size="50" style="color: #C0C4CC; margin-bottom: 15px">
            <Warning />
          </el-icon>
          <p style="margin: 0; color: #909399">该告警尚未进行AI分析</p>
          <p style="margin: 5px 0 0 0; font-size: 12px; color: #C0C4CC">
            点击下方"开始分析"按钮启动AI智能分析
          </p>
        </el-card>
      </el-card>
      
      <template #footer>
        <el-button @click="analyzeDialogVisible = false">关闭</el-button>
        
        <el-button 
          v-if="!workflowStatus || workflowStatus === 'Completed' || workflowStatus === 'Failed'"
          type="primary"
          :loading="analyzing"
          @click="startAnalysis"
        >
          {{ workflowStatus === 'Completed' ? '重新分析' : '开始分析' }}
        </el-button>
        
        <el-button 
          v-if="workflowStatus === 'Running'"
          type="primary" 
          :loading="checkingStatus"
          @click="checkAnalysisStatus"
        >
          刷新状态
        </el-button>
        
        <el-button 
          type="success" 
          :loading="gettingResult"
          @click="getAnalysisResult"
          v-if="workflowStatus === 'Completed' && !analysisResultText"
        >
          获取结果
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { alertApi } from '@/api'
import { ElMessage } from 'element-plus'
import { Loading, Warning } from '@element-plus/icons-vue'

const alerts = ref<any[]>([])
const loading = ref(false)
const submitting = ref(false)
const analyzing = ref(false)
const checkingStatus = ref(false)
const gettingResult = ref(false)

const dialogVisible = ref(false)
const analyzeDialogVisible = ref(false)

const alertForm = ref({
  source: 'manual',
  severity: 'warning',
  title: '',
  description: ''
})

const alertAnalysisCache = ref<Map<string, any>>(new Map())

const loadAnalysisCache = async () => {
  try {
    const res = await alertApi.listAnalysis(100, 0)
    if (res && res.code === 200) {
      const cacheMap = new Map<string, any>()
      const results = res.data || []
      for (const result of results) {
        cacheMap.set(result.alert_id, {
          workflowId: result.workflow_id,
          status: result.status,
          result: JSON.parse(result.result || '{}'),
          resultText: result.analysis_text,
          timestamp: result.created_at
        })
      }
      alertAnalysisCache.value = cacheMap
      console.log('从后端加载分析缓存:', cacheMap.size, '条')
    }
  } catch (error) {
    console.error('加载分析缓存失败:', error)
  }
}

const saveAnalysisToBackend = async (alertId: string, workflowId: string, status: string, result: any, resultText: string) => {
  try {
    await alertApi.saveAnalysis({
      alert_id: alertId,
      workflow_id: workflowId,
      status: status,
      result: result,
      analysis_text: resultText
    })
    console.log('分析结果已保存到后端')
  } catch (error) {
    console.error('保存分析结果失败:', error)
  }
}

onMounted(() => {
  loadAnalysisCache()
  loadAlerts()
})

const currentWorkflowId = ref('')
const workflowStatus = ref('')
const analysisResult = ref<any>(null)
const analysisResultText = ref('')
const currentAlert = ref<any>(null)

const getSeverityType = (severity: string) => {
  switch (severity) {
    case 'critical':
      return 'danger'
    case 'warning':
      return 'warning'
    case 'info':
      return 'info'
    default:
      return ''
  }
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'open':
      return 'warning'
    case 'resolved':
      return 'success'
    case 'acknowledged':
      return 'info'
    default:
      return ''
  }
}

const getWorkflowStatusType = (status: string) => {
  switch (status) {
    case 'Running':
      return 'warning'
    case 'Completed':
      return 'success'
    case 'Failed':
      return 'danger'
    default:
      return 'info'
  }
}

const loadAlerts = async () => {
  loading.value = true
  try {
    const res = await alertApi.list()
    if (res && res.code === 200) {
      alerts.value = res.data || []
      ElMessage.success(`已加载 ${alerts.value.length} 个告警`)
    }
  } catch (error: any) {
    ElMessage.error('加载告警失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  alertForm.value.source = 'manual'
  alertForm.value.severity = 'warning'
  alertForm.value.title = ''
  alertForm.value.description = ''
  dialogVisible.value = true
}

const submitAlert = async () => {
  if (!alertForm.value.title) {
    ElMessage.warning('请输入告警标题')
    return
  }
  
  submitting.value = true
  try {
    await alertApi.create({
      source: alertForm.value.source,
      severity: alertForm.value.severity,
      title: alertForm.value.title,
      description: alertForm.value.description,
      raw_data: '{}'
    })
    
    ElMessage.success('告警已创建')
    dialogVisible.value = false
    loadAlerts()
  } catch (error: any) {
    ElMessage.error('创建失败: ' + error.message)
  } finally {
    submitting.value = false
  }
}

const analyzeAlert = async (alert: any) => {
  currentAlert.value = alert
  
  const cachedAnalysis = alertAnalysisCache.value.get(alert.id)
  if (cachedAnalysis) {
    currentWorkflowId.value = cachedAnalysis.workflowId
    workflowStatus.value = cachedAnalysis.status
    analysisResult.value = cachedAnalysis.result
    analysisResultText.value = cachedAnalysis.resultText
    ElMessage.info('已加载历史分析结果')
  } else {
    try {
      const res = await alertApi.getAnalysis(alert.id)
      if (res && res.code === 200 && res.data) {
        const backendResult = res.data
        currentWorkflowId.value = backendResult.workflow_id
        workflowStatus.value = backendResult.status
        analysisResult.value = JSON.parse(backendResult.result || '{}')
        analysisResultText.value = backendResult.analysis_text
        
        alertAnalysisCache.value.set(alert.id, {
          workflowId: backendResult.workflow_id,
          status: backendResult.status,
          result: analysisResult.value,
          resultText: backendResult.analysis_text,
          timestamp: backendResult.created_at
        })
        
        ElMessage.info('已加载历史分析结果')
      } else {
        currentWorkflowId.value = ''
        workflowStatus.value = ''
        analysisResult.value = null
        analysisResultText.value = ''
      }
    } catch (error) {
      console.log('告警未分析或获取失败:', error)
      currentWorkflowId.value = ''
      workflowStatus.value = ''
      analysisResult.value = null
      analysisResultText.value = ''
    }
  }
  
  analyzeDialogVisible.value = true
}

const startAnalysis = async () => {
  if (!currentAlert.value) {
    ElMessage.warning('未选择告警')
    return
  }
  
  currentWorkflowId.value = ''
  workflowStatus.value = ''
  analysisResult.value = null
  analysisResultText.value = ''
  analyzing.value = true
  
  try {
    workflowStatus.value = 'Failed'
    analyzing.value = false
    ElMessage.error('AI分析功能已禁用 - workflowApi已移除')
  } catch (error: any) {
    ElMessage.error('启动分析失败: ' + error.message)
    analyzing.value = false
    workflowStatus.value = 'Failed'
  }
}

const checkAnalysisStatus = async () => {
  if (!currentWorkflowId.value) return
  
  checkingStatus.value = true
  try {
    ElMessage.warning('状态查询功能已禁用 - workflowApi已移除')
  } catch (error: any) {
    ElMessage.error('查询状态失败: ' + error.message)
  } finally {
    checkingStatus.value = false
  }
}

const getAnalysisResult = async () => {
  if (!currentWorkflowId.value) return
  
  gettingResult.value = true
  try {
    ElMessage.warning('结果获取功能已禁用 - workflowApi已移除')
  } catch (error: any) {
    ElMessage.error('获取结果失败: ' + error.message)
  } finally {
    gettingResult.value = false
  }
}
</script>

<style scoped>
.page-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>