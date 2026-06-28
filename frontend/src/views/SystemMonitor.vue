<template>
  <div>
    <el-card>
      <template #header>
        <h3>系统监控</h3>
      </template>
      
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="在线Agent" :value="agentCount" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="Workflow总数" :value="workflowCount" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="告警数量" :value="alertCount" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="活跃用户" :value="userCount" />
        </el-col>
      </el-row>
    </el-card>
    
    <el-card style="margin-top: 20px">
      <template #header>
        <div class="header">
          <h3>Workflow执行统计</h3>
          <el-button size="small" @click="refreshStats">
            刷新统计
          </el-button>
        </div>
      </template>
      
      <el-table :data="workflowStats">
        <el-table-column prop="agent_id" label="Agent ID" />
        <el-table-column prop="task_type" label="任务类型" />
        <el-table-column prop="total" label="执行次数" width="100" />
        <el-table-column prop="success" label="成功次数" width="100">
          <template #default="{ row }">
            <el-tag type="success">{{ row.success }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="failed" label="失败次数" width="100">
          <template #default="{ row }">
            <el-tag type="danger">{{ row.failed }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="avg_time" label="平均耗时" width="120" />
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
          <el-tag type="success">已连接</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Redis">
          <el-tag type="success">已连接</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="LLM Provider">
          <el-tag type="success">阿里云百炼</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
    
    <el-card style="margin-top: 20px">
      <template #header>
        <h3>最近执行记录</h3>
      </template>
      
      <el-table :data="recentExecutions" v-loading="loadingExecutions">
        <el-table-column prop="workflow_id" label="Workflow ID" width="300" />
        <el-table-column prop="task_type" label="任务类型" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration" label="耗时" width="120" />
        <el-table-column prop="created_at" label="执行时间" />
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
const userCount = ref(1)

const workflowStats = ref<any[]>([])
const recentExecutions = ref<any[]>([])
const loadingExecutions = ref(false)

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

onMounted(() => {
  loadStats()
})

const loadStats = async () => {
  try {
    const [agentRes, alertRes] = await Promise.all([
      agentApi.list(),
      alertApi.list()
    ])
    
    if (agentRes && agentRes.code === 200) {
      agentCount.value = (agentRes.data || []).length
    }
    
    if (alertRes && alertRes.code === 200) {
      alertCount.value = (alertRes.data || []).length
    }
    
    loadWorkflowStats()
    loadRecentExecutions()
  } catch (error: any) {
    ElMessage.error('加载统计失败: ' + error.message)
  }
}

const loadWorkflowStats = () => {
  workflowStats.value = [
    {
      agent_id: 'monitor-agent-001',
      task_type: 'alert_analysis',
      total: 5,
      success: 5,
      failed: 0,
      avg_time: '2.5s'
    },
    {
      agent_id: 'analysis-agent-001',
      task_type: 'fault_diagnosis',
      total: 3,
      success: 3,
      failed: 0,
      avg_time: '3.2s'
    }
  ]
}

const loadRecentExecutions = () => {
  recentExecutions.value = [
    {
      workflow_id: 'workflow-monitor-agent-001-xxx',
      task_type: 'alert_analysis',
      status: 'Completed',
      duration: '2.3s',
      created_at: '2026-06-24 23:31:04'
    },
    {
      workflow_id: 'workflow-monitor-agent-001-yyy',
      task_type: 'alert_analysis',
      status: 'Completed',
      duration: '2.5s',
      created_at: '2026-06-24 23:31:06'
    }
  ]
}

const refreshStats = () => {
  loadStats()
  ElMessage.success('统计已刷新')
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
