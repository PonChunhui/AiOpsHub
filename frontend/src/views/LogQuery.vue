<template>
  <div>
    <el-row :gutter="20" class="header">
      <el-col :span="24">
        <h1>日志查询</h1>
      </el-col>
    </el-row>

    <el-card class="query-card">
      <el-form :model="queryForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="命名空间">
              <el-select v-model="queryForm.namespace" placeholder="选择命名空间">
                <el-option label="全部" value="" />
                <el-option label="default" value="default" />
                <el-option label="kube-system" value="kube-system" />
                <el-option label="monitoring" value="monitoring" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Pod名称">
              <el-input v-model="queryForm.podName" placeholder="输入Pod名称" clearable />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="容器名称">
              <el-input v-model="queryForm.container" placeholder="输入容器名称" clearable />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="queryForm.timeRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="日志级别">
              <el-select v-model="queryForm.level" placeholder="选择日志级别">
                <el-option label="全部" value="" />
                <el-option label="INFO" value="info" />
                <el-option label="WARN" value="warn" />
                <el-option label="ERROR" value="error" />
                <el-option label="DEBUG" value="debug" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="搜索关键词">
              <el-input v-model="queryForm.keyword" placeholder="输入关键词" clearable />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="Loki查询">
              <el-input
                v-model="queryForm.lokiQuery"
                type="textarea"
                :rows="3"
                placeholder='例如: {namespace="default", pod="my-app-xxx"} |= "error"'
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <el-button type="primary" @click="fetchLogs" :loading="loading">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetQuery">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
          <el-button @click="exportLogs" :disabled="logs.length === 0">
            <el-icon><Download /></el-icon>
            导出
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="stats-card" v-if="logs.length > 0">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="总日志数" :value="logStats.total" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="错误数" :value="logStats.error" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="警告数" :value="logStats.warn" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="信息数" :value="logStats.info" />
        </el-col>
      </el-row>
    </el-card>

    <el-card class="logs-card">
      <el-table
        :data="paginatedLogs"
        style="width: 100%"
        :row-class-name="getRowClassName"
        max-height="600"
      >
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.timestamp) }}
          </template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="80">
          <template #default="scope">
            <el-tag :type="getLevelType(scope.row.level)" size="small">
              {{ scope.row.level }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pod" label="Pod" width="200" />
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="message" label="日志内容">
          <template #default="scope">
            <div class="log-message" v-html="highlightKeyword(scope.row.message)"></div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button type="primary" size="small" text @click="viewDetail(scope.row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="logs.length"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
      />
    </el-card>

    <el-dialog v-model="detailDialogVisible" title="日志详情" width="60%">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="时间">{{ formatTime(selectedLog?.timestamp) }}</el-descriptions-item>
        <el-descriptions-item label="级别">
          <el-tag :type="getLevelType(selectedLog?.level)">{{ selectedLog?.level }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Pod">{{ selectedLog?.pod }}</el-descriptions-item>
        <el-descriptions-item label="命名空间">{{ selectedLog?.namespace }}</el-descriptions-item>
        <el-descriptions-item label="容器">{{ selectedLog?.container }}</el-descriptions-item>
        <el-descriptions-item label="日志流">{{ selectedLog?.stream }}</el-descriptions-item>
      </el-descriptions>
      <div class="log-detail-message">
        <h4>日志内容:</h4>
        <pre>{{ selectedLog?.message }}</pre>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyLog">复制</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { logApi } from '../api'
import { ElMessage } from 'element-plus'
import { Search, Refresh, Download } from '@element-plus/icons-vue'

const queryForm = ref({
  namespace: '',
  podName: '',
  container: '',
  timeRange: [new Date(Date.now() - 3600000), new Date()] as [Date, Date],
  level: '',
  keyword: '',
  lokiQuery: ''
})

const loading = ref(false)
const logs = ref<any[]>([])
const currentPage = ref(1)
const pageSize = ref(50)
const detailDialogVisible = ref(false)
const selectedLog = ref<any>(null)

const logStats = computed(() => {
  return {
    total: logs.value.length,
    error: logs.value.filter(l => l.level === 'error').length,
    warn: logs.value.filter(l => l.level === 'warn').length,
    info: logs.value.filter(l => l.level === 'info').length
  }
})

const paginatedLogs = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return logs.value.slice(start, end)
})

const fetchLogs = async () => {
  loading.value = true
  try {
    const startTime = queryForm.value.timeRange[0].toISOString()
    const endTime = queryForm.value.timeRange[1].toISOString()

    let query = queryForm.value.lokiQuery
    if (!query) {
      const labels: string[] = []
      if (queryForm.value.namespace) {
        labels.push(`namespace="${queryForm.value.namespace}"`)
      }
      if (queryForm.value.podName) {
        labels.push(`pod="${queryForm.value.podName}"`)
      }
      if (queryForm.value.container) {
        labels.push(`container="${queryForm.value.container}"`)
      }
      query = `{${labels.join(', ')}}`
      if (queryForm.value.keyword) {
        query += ` |= "${queryForm.value.keyword}"`
      }
      if (queryForm.value.level) {
        query += ` | level="${queryForm.value.level}"`
      }
    }

    const res = await logApi.queryLoki(query, startTime, endTime, 1000)
    if (res.data?.result) {
      logs.value = res.data.result.flatMap((stream: any) => {
        return stream.values.map((val: any) => ({
          timestamp: parseInt(val[0]),
          message: val[1],
          level: extractLevel(val[1]),
          pod: stream.stream.pod || '-',
          namespace: stream.stream.namespace || '-',
          container: stream.stream.container || '-',
          stream: stream.stream.stream || '-'
        }))
      })
    }
    ElMessage.success(`查询到 ${logs.value.length} 条日志`)
  } catch (error) {
    ElMessage.error('查询日志失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const extractLevel = (message: string): string => {
  if (message.includes('ERROR') || message.includes('error')) return 'error'
  if (message.includes('WARN') || message.includes('warn')) return 'warn'
  if (message.includes('DEBUG') || message.includes('debug')) return 'debug'
  return 'info'
}

const resetQuery = () => {
  queryForm.value = {
    namespace: '',
    podName: '',
    container: '',
    timeRange: [new Date(Date.now() - 3600000), new Date()] as [Date, Date],
    level: '',
    keyword: '',
    lokiQuery: ''
  }
  logs.value = []
}

const exportLogs = () => {
  const content = logs.value
    .map(l => `[${formatTime(l.timestamp)}] [${l.level.toUpperCase()}] ${l.pod} ${l.namespace}: ${l.message}`)
    .join('\n')

  const blob = new Blob([content], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `logs-${Date.now()}.txt`
  a.click()
  window.URL.revokeObjectURL(url)
  ElMessage.success('日志已导出')
}

const formatTime = (timestamp: number | undefined) => {
  if (!timestamp) return '-'
  return new Date(timestamp / 1000000).toLocaleString()
}

const getLevelType = (level: string | undefined) => {
  switch (level) {
    case 'error':
      return 'danger'
    case 'warn':
      return 'warning'
    case 'debug':
      return 'info'
    default:
      return 'success'
  }
}

const getRowClassName = ({ row }: { row: any }) => {
  if (row.level === 'error') return 'error-row'
  if (row.level === 'warn') return 'warning-row'
  return ''
}

const highlightKeyword = (message: string) => {
  if (!queryForm.value.keyword) return message
  const regex = new RegExp(`(${queryForm.value.keyword})`, 'gi')
  return message.replace(regex, '<span class="highlight">$1</span>')
}

const viewDetail = (log: any) => {
  selectedLog.value = log
  detailDialogVisible.value = true
}

const copyLog = () => {
  if (!selectedLog.value) return
  const text = `[${formatTime(selectedLog.value.timestamp)}] [${selectedLog.value.level.toUpperCase()}] ${selectedLog.value.pod} ${selectedLog.value.namespace}: ${selectedLog.value.message}`
  navigator.clipboard.writeText(text)
  ElMessage.success('日志已复制到剪贴板')
}

onMounted(() => {
  fetchLogs()
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

.query-card {
  margin-bottom: 20px;
}

.stats-card {
  margin-bottom: 20px;
}

.logs-card {
  margin-bottom: 20px;
}

.log-message {
  word-break: break-all;
}

.log-message :deep(.highlight) {
  background-color: #ffc107;
  padding: 2px 4px;
  border-radius: 2px;
}

:deep(.error-row) {
  background-color: #fef0f0;
}

:deep(.warning-row) {
  background-color: #fdf6ec;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.log-detail-message {
  margin-top: 20px;
}

.log-detail-message pre {
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>