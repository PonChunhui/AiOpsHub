<template>
  <div>
    <el-row :gutter="20" class="header">
      <el-col :span="24">
        <h1>自动修复</h1>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="修复规则" name="rules">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>修复规则列表</span>
              <el-button type="primary" @click="showCreateDialog">
                <el-icon><Plus /></el-icon>
                新建规则
              </el-button>
            </div>
          </template>

          <el-table :data="rules" style="width: 100%">
            <el-table-column prop="name" label="规则名称" width="200" />
            <el-table-column prop="trigger" label="触发条件" width="250">
              <template #default="scope">
                <el-tag>{{ scope.row.trigger.metric }}</el-tag>
                <span>{{ scope.row.trigger.operator }} {{ scope.row.trigger.threshold }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="action" label="修复动作" width="150">
              <template #default="scope">
                <el-tag type="success">{{ scope.row.action.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-switch v-model="scope.row.enabled" @change="toggleRule(scope.row)" />
              </template>
            </el-table-column>
            <el-table-column prop="lastTriggered" label="最后触发" width="180">
              <template #default="scope">
                {{ scope.row.lastTriggered ? formatTime(scope.row.lastTriggered) : '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="triggerCount" label="触发次数" width="100" />
            <el-table-column label="操作" width="150">
              <template #default="scope">
                <el-button type="primary" size="small" text @click="editRule(scope.row)">
                  编辑
                </el-button>
                <el-button type="danger" size="small" text @click="deleteRule(scope.row.id)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="执行历史" name="history">
        <el-card>
          <el-form :inline="true" class="filter-form">
            <el-form-item label="状态">
              <el-select v-model="historyFilter.status" placeholder="全部" clearable>
                <el-option label="成功" value="success" />
                <el-option label="失败" value="failed" />
                <el-option label="执行中" value="running" />
              </el-select>
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="historyFilter.timeRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchHistory">查询</el-button>
            </el-form-item>
          </el-form>

          <el-table :data="history" style="width: 100%">
            <el-table-column prop="id" label="执行ID" width="200" />
            <el-table-column prop="ruleName" label="规则名称" width="200" />
            <el-table-column prop="trigger" label="触发原因" width="250" />
            <el-table-column prop="action" label="执行动作" width="150" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getStatusType(scope.row.status)">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="startTime" label="开始时间" width="180">
              <template #default="scope">
                {{ formatTime(scope.row.startTime) }}
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="耗时" width="100">
              <template #default="scope">
                {{ scope.row.duration }}ms
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="scope">
                <el-button type="primary" size="small" text @click="viewExecution(scope.row)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="historyPage"
            :page-size="20"
            :total="historyTotal"
            layout="total, prev, pager, next"
            class="pagination"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="手动修复" name="manual">
        <el-card>
          <el-form :model="manualForm" label-width="120px">
            <el-form-item label="修复类型">
              <el-select v-model="manualForm.type" placeholder="选择修复类型">
                <el-option label="重启Pod" value="restart_pod" />
                <el-option label="扩容Deployment" value="scale_deployment" />
                <el-option label="清理Pod" value="delete_pod" />
                <el-option label="更新配置" value="update_config" />
                <el-option label="执行命令" value="exec_command" />
              </el-select>
            </el-form-item>

            <el-form-item label="命名空间">
              <el-select v-model="manualForm.namespace" placeholder="选择命名空间">
                <el-option label="default" value="default" />
                <el-option label="kube-system" value="kube-system" />
                <el-option label="monitoring" value="monitoring" />
              </el-select>
            </el-form-item>

            <el-form-item label="资源名称">
              <el-input v-model="manualForm.resourceName" placeholder="输入Pod或Deployment名称" />
            </el-form-item>

            <el-form-item label="参数配置" v-if="manualForm.type === 'scale_deployment'">
              <el-input-number v-model="manualForm.replicas" :min="1" :max="100" />
            </el-form-item>

            <el-form-item label="执行命令" v-if="manualForm.type === 'exec_command'">
              <el-input v-model="manualForm.command" type="textarea" :rows="3" placeholder="输入要执行的命令" />
            </el-form-item>

            <el-form-item label="确认方式">
              <el-radio-group v-model="manualForm.confirmMode">
                <el-radio label="auto">自动执行</el-radio>
                <el-radio label="manual">人工确认</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="executeManual" :loading="executing">
                <el-icon><VideoPlay /></el-icon>
                执行修复
              </el-button>
              <el-button @click="dryRun">
                <el-icon><View /></el-icon>
                预览影响
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="ruleDialogVisible" :title="editingRule ? '编辑规则' : '新建规则'" width="60%">
      <el-form :model="ruleForm" label-width="120px">
        <el-form-item label="规则名称">
          <el-input v-model="ruleForm.name" placeholder="输入规则名称" />
        </el-form-item>

        <el-divider content-position="left">触发条件</el-divider>

        <el-form-item label="监控指标">
          <el-select v-model="ruleForm.trigger.metric" placeholder="选择指标">
            <el-option label="CPU使用率" value="cpu_usage" />
            <el-option label="内存使用率" value="memory_usage" />
            <el-option label="Pod重启次数" value="pod_restart_count" />
            <el-option label="HTTP错误率" value="http_error_rate" />
            <el-option label="响应时间" value="response_time" />
          </el-select>
        </el-form-item>

        <el-form-item label="条件">
          <el-select v-model="ruleForm.trigger.operator" style="width: 100px">
            <el-option label=">" value=">" />
            <el-option label="<" value="<" />
            <el-option label="=" value="=" />
            <el-option label=">=" value=">=" />
            <el-option label="<=" value="<=" />
          </el-select>
          <el-input-number v-model="ruleForm.trigger.threshold" :min="0" style="width: 200px" />
          <span style="margin-left: 10px">{{ getMetricUnit(ruleForm.trigger.metric) }}</span>
        </el-form-item>

        <el-form-item label="持续时间">
          <el-input-number v-model="ruleForm.trigger.duration" :min="1" />
          <el-select v-model="ruleForm.trigger.durationUnit" style="width: 100px">
            <el-option label="秒" value="s" />
            <el-option label="分" value="m" />
            <el-option label="时" value="h" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">修复动作</el-divider>

        <el-form-item label="动作类型">
          <el-select v-model="ruleForm.action.type" placeholder="选择动作类型">
            <el-option label="重启Pod" value="restart_pod" />
            <el-option label="扩容" value="scale_up" />
            <el-option label="缩容" value="scale_down" />
            <el-option label="发送通知" value="notify" />
            <el-option label="执行脚本" value="execute_script" />
          </el-select>
        </el-form-item>

        <el-form-item label="目标资源" v-if="['restart_pod', 'scale_up', 'scale_down'].includes(ruleForm.action.type)">
          <el-input v-model="ruleForm.action.target" placeholder="例如: deployment/my-app" />
        </el-form-item>

        <el-form-item label="副本数" v-if="['scale_up', 'scale_down'].includes(ruleForm.action.type)">
          <el-input-number v-model="ruleForm.action.replicas" :min="1" :max="100" />
        </el-form-item>

        <el-form-item label="通知渠道" v-if="ruleForm.action.type === 'notify'">
          <el-select v-model="ruleForm.action.channel" multiple>
            <el-option label="邮件" value="email" />
            <el-option label="企业微信" value="wechat" />
            <el-option label="钉钉" value="dingtalk" />
            <el-option label="Slack" value="slack" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">安全设置</el-divider>

        <el-form-item label="需要审批">
          <el-switch v-model="ruleForm.security.requiresApproval" />
        </el-form-item>

        <el-form-item label="最大执行次数">
          <el-input-number v-model="ruleForm.security.maxExecution" :min="1" :max="10" />
          <span style="margin-left: 10px">（每小时）</span>
        </el-form-item>

        <el-form-item label="静默期">
          <el-input-number v-model="ruleForm.security.cooldown" :min="0" />
          <span style="margin-left: 10px">分钟</span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="executionDialogVisible" title="执行详情" width="60%">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="执行ID">{{ selectedExecution?.id }}</el-descriptions-item>
        <el-descriptions-item label="规则名称">{{ selectedExecution?.ruleName }}</el-descriptions-item>
        <el-descriptions-item label="触发原因">{{ selectedExecution?.trigger }}</el-descriptions-item>
        <el-descriptions-item label="执行动作">{{ selectedExecution?.action }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(selectedExecution?.status)">
            {{ selectedExecution?.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(selectedExecution?.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(selectedExecution?.endTime) }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ selectedExecution?.duration }}ms</el-descriptions-item>
      </el-descriptions>

      <div class="execution-log">
        <h4>执行日志:</h4>
        <el-timeline>
          <el-timeline-item
            v-for="(log, index) in executionLogs"
            :key="index"
            :timestamp="formatTime(log.timestamp)"
            :type="log.type"
          >
            {{ log.message }}
          </el-timeline-item>
        </el-timeline>
      </div>

      <template #footer>
        <el-button @click="executionDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { remediationApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, View } from '@element-plus/icons-vue'

const activeTab = ref('rules')

const rules = ref<any[]>([])
const history = ref<any[]>([])
const historyPage = ref(1)
const historyTotal = ref(0)

const historyFilter = ref({
  status: '',
  timeRange: [] as Date[]
})

const ruleDialogVisible = ref(false)
const editingRule = ref<any>(null)
const ruleForm = ref({
  name: '',
  trigger: {
    metric: 'cpu_usage',
    operator: '>',
    threshold: 80,
    duration: 5,
    durationUnit: 'm'
  },
  action: {
    type: 'restart_pod',
    target: '',
    replicas: 3,
    channel: []
  },
  security: {
    requiresApproval: false,
    maxExecution: 3,
    cooldown: 30
  },
  enabled: true
})

const manualForm = ref({
  type: 'restart_pod',
  namespace: 'default',
  resourceName: '',
  replicas: 3,
  command: '',
  confirmMode: 'auto'
})

const executing = ref(false)
const executionDialogVisible = ref(false)
const selectedExecution = ref<any>(null)
const executionLogs = ref<any[]>([])

const fetchRules = async () => {
  try {
    const res = await remediationApi.getRules()
    if (res.data?.rules) {
      rules.value = res.data.rules
    }
  } catch (error) {
    ElMessage.error('获取规则列表失败')
    console.error(error)
  }
}

const fetchHistory = async () => {
  try {
    const params: any = {
      page: historyPage.value,
      pageSize: 20
    }
    if (historyFilter.value.status) {
      params.status = historyFilter.value.status
    }
    if (historyFilter.value.timeRange && historyFilter.value.timeRange.length === 2) {
      params.startTime = historyFilter.value.timeRange[0]?.toISOString() || ''
      params.endTime = historyFilter.value.timeRange[1]?.toISOString() || ''
    }

    const res = await remediationApi.getHistory(params)
    if (res.data) {
      history.value = res.data.history || []
      historyTotal.value = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取执行历史失败')
    console.error(error)
  }
}

const showCreateDialog = () => {
  editingRule.value = null
  ruleForm.value = {
    name: '',
    trigger: {
      metric: 'cpu_usage',
      operator: '>',
      threshold: 80,
      duration: 5,
      durationUnit: 'm'
    },
    action: {
      type: 'restart_pod',
      target: '',
      replicas: 3,
      channel: []
    },
    security: {
      requiresApproval: false,
      maxExecution: 3,
      cooldown: 30
    },
    enabled: true
  }
  ruleDialogVisible.value = true
}

const editRule = (rule: any) => {
  editingRule.value = rule
  ruleForm.value = { ...rule }
  ruleDialogVisible.value = true
}

const saveRule = async () => {
  try {
    if (editingRule.value) {
      await remediationApi.updateRule(editingRule.value.id, ruleForm.value)
      ElMessage.success('规则已更新')
    } else {
      await remediationApi.createRule(ruleForm.value)
      ElMessage.success('规则已创建')
    }
    ruleDialogVisible.value = false
    fetchRules()
  } catch (error) {
    ElMessage.error('保存规则失败')
    console.error(error)
  }
}

const deleteRule = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await remediationApi.deleteRule(id)
    ElMessage.success('规则已删除')
    fetchRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除规则失败')
      console.error(error)
    }
  }
}

const toggleRule = async (rule: any) => {
  try {
    await remediationApi.updateRule(rule.id, { enabled: rule.enabled })
    ElMessage.success(rule.enabled ? '规则已启用' : '规则已禁用')
  } catch (error) {
    rule.enabled = !rule.enabled
    ElMessage.error('更新规则状态失败')
    console.error(error)
  }
}

const executeManual = async () => {
  if (!manualForm.value.resourceName) {
    ElMessage.warning('请输入资源名称')
    return
  }

  try {
    executing.value = true
    const res = await remediationApi.executeManual({
      type: manualForm.value.type,
      namespace: manualForm.value.namespace,
      resourceName: manualForm.value.resourceName,
      replicas: manualForm.value.replicas,
      command: manualForm.value.command,
      confirmMode: manualForm.value.confirmMode
    })
    ElMessage.success(`修复任务已提交，执行ID: ${res.data.executionId}`)
  } catch (error) {
    ElMessage.error('执行修复失败')
    console.error(error)
  } finally {
    executing.value = false
  }
}

const dryRun = async () => {
  if (!manualForm.value.resourceName) {
    ElMessage.warning('请输入资源名称')
    return
  }

  try {
    const res = await remediationApi.dryRun({
      type: manualForm.value.type,
      namespace: manualForm.value.namespace,
      resourceName: manualForm.value.resourceName
    })
    ElMessageBox.alert(JSON.stringify(res.data, null, 2), '预览影响', {
      confirmButtonText: '确定'
    })
  } catch (error) {
    ElMessage.error('预览失败')
    console.error(error)
  }
}

const viewExecution = async (execution: any) => {
  selectedExecution.value = execution
  try {
    const res = await remediationApi.getExecutionLogs(execution.id)
    executionLogs.value = res.data?.logs || []
  } catch (error) {
    executionLogs.value = []
  }
  executionDialogVisible.value = true
}

const formatTime = (timestamp: string | undefined) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

const getStatusType = (status: string | undefined) => {
  switch (status) {
    case 'success':
      return 'success'
    case 'failed':
      return 'danger'
    case 'running':
      return 'warning'
    default:
      return 'info'
  }
}

const getMetricUnit = (metric: string) => {
  const units: any = {
    cpu_usage: '%',
    memory_usage: '%',
    pod_restart_count: '次',
    http_error_rate: '%',
    response_time: 'ms'
  }
  return units[metric] || ''
}

onMounted(() => {
  fetchRules()
  fetchHistory()
})
</script>

<style scoped>
.header {
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.execution-log {
  margin-top: 20px;
}

.execution-log h4 {
  margin-bottom: 15px;
}
</style>