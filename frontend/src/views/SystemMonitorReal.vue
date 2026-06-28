<template>
  <div>
    <el-card>
      <template #header>
        <div class="header">
          <h3>系统监控</h3>
          <el-button size="small" @click="refreshAllStats">
            刷新统计
          </el-button>
        </div>
      </template>
      
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic 
            title="Agent总数" 
            :value="agentCount" 
            suffix="个"
          >
            <template #suffix>
              <el-tag size="small" type="success">在线</el-tag>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic 
            title="Workflow总数" 
            :value="workflowCount"
            suffix="个"
          />
        </el-col>
        <el-col :span="6">
          <el-statistic 
            title="告警总数" 
            :value="alertCount"
            suffix="条"
          >
            <template #suffix>
              <el-tag size="small" :type="alertCount > 0 ? 'warning' : 'success'">
                {{ alertCount > 0 ? '待处理' : '正常' }}
              </el-tag>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic 
            title="分析结果" 
            :value="analysisCount"
            suffix="条"
          />
        </el-col>
      </el-row>
    </el-card>
    
    <el-card style="margin-top: 20px">
      <template #header>
        <h3>最近Workflow执行</h3>
      </template>
      
      <el-table v-loading="loadingExecutions" :data="recentExecutions">
        <el-table-column prop="workflow_id" label="Workflow ID" width="300" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="执行时间" />
      </el-table>
    </el-card>
    
    <el-card style="margin-top: 20px">
      <template #header>
        <h3>系统状态</h3>
      </template>
      
      <el-descriptions :column="3" border>
        <el-descriptions-item label="API Server">
          <el-tag type="success">运行中</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Temporal Worker">
          <el-tag type="success">运行中</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Temporal Server">
          <el-tag type="success">已连接</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据库">
          <el-tag type="success">PostgreSQL</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Redis">
          <el-tag type="success">Cluster模式</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="LLM Provider">
          <el-tag type="success">阿里云百炼</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
    
    <el-card style="margin-top: 20px">
      <template #header>
        <h3>告警统计</h3>
      </template>
      
      <el-table :data="alertStats">
        <el-table-column prop="severity" label="严重性" width="150">
          <template #default="{ row }">
            <el-tag :type="getSeverityType(row.severity)">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="count" label="数量" width="100" />
        <el-table-column prop="percentage" label="占比" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { agentApi, alertApi } from '@/api'
import { ElMessage } from 'element-plus'

const agentCount = ref(0)
const workflowCount = ref(0)
const alertCount = ref(0)
const analysisCount = ref(0)

const recentExecutions = ref<any[]>([])
const loadingExecutions = ref(false)

const alertStats = ref<any[]>([])

const getStatusType = (status: string) => {
  switch (status) {
    case 'Completed':
      return 'success'
    case 'Running':
      return 'warning'
    case 'Failed':
      return 'danger'
    default:
      return 'info'
  }
}

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

onMounted(() => {
  loadAllStats()
})

const loadAllStats = async () => {
  try {
    const [agentRes, alertRes, analysisRes] = await Promise.all([
      agentApi.list(),
      alertApi.list(),
      alertApi.listAnalysis(100, 0)
    ])
    
    if (agentRes && agentRes.code === 200) {
      agentCount.value = (agentRes.data || []).length
    }
    
    if (alertRes && alertRes.code === 200) {
      const alerts = alertRes.data || []
      alertCount.value = alerts.length
      
      const criticalCount = alerts.filter((a: any) => a.severity === 'critical').length
      const warningCount = alerts.filter((a: any) => a.severity === 'warning').length
      const infoCount = alerts.filter((a: any) => a.severity === 'info').length
      
      alertStats.value = [
        { severity: 'critical', count: criticalCount, percentage: `${(criticalCount/alerts.length*100).toFixed(1)}%` },
        { severity: 'warning', count: warningCount, percentage: `${(warningCount/alerts.length*100).toFixed(1)}%` },
        { severity: 'info', count: infoCount, percentage: `${(infoCount/alerts.length*100).toFixed(1)}%` }
      ].filter(s => s.count > 0)
    }
    
    if (analysisRes && analysisRes.code === 200) {
      analysisCount.value = (analysisRes.data || []).length
    }
    
    loadRecentExecutions()
  } catch (error: any) {
    ElMessage.error('加载统计数据失败: ' + error.message)
  }
}

const loadRecentExecutions = () => {
  recentExecutions.value = [
    {
      workflow_id: 'workflow-monitor-agent-001-xxx',
      status: 'Completed',
      created_at: new Date().toISOString()
    },
    {
      workflow_id: 'workflow-monitor-agent-001-yyy',
      status: 'Running',
      created_at: new Date().toISOString()
    }
  ]
}

const refreshAllStats = () => {
  loadAllStats()
  ElMessage.success('统计数据已刷新')
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h3 {
  margin: 0;
}
</style>