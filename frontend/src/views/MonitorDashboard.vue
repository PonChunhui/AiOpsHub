<template>
  <div>
    <el-row :gutter="20" class="header">
      <el-col :span="24">
        <h1>监控仪表板</h1>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="filters">
      <el-col :span="6">
        <el-select v-model="selectedNamespace" placeholder="选择命名空间" @change="fetchMetrics">
          <el-option label="全部" value="" />
          <el-option label="default" value="default" />
          <el-option label="kube-system" value="kube-system" />
          <el-option label="monitoring" value="monitoring" />
        </el-select>
      </el-col>
      <el-col :span="6">
        <el-date-picker
          v-model="timeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          @change="fetchMetrics"
        />
      </el-col>
      <el-col :span="6">
        <el-button type="primary" @click="fetchMetrics">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="metrics-row">
      <el-col :span="6">
        <el-card class="metric-card">
          <template #header>
            <div class="card-header">
              <span>CPU使用率</span>
              <el-icon class="metric-icon cpu"><Cpu /></el-icon>
            </div>
          </template>
          <div class="metric-value">{{ cpuUsage }}%</div>
          <div class="metric-trend" :class="cpuTrend > 0 ? 'up' : 'down'">
            <el-icon v-if="cpuTrend > 0"><Top /></el-icon>
            <el-icon v-else><Bottom /></el-icon>
            {{ Math.abs(cpuTrend) }}%
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="metric-card">
          <template #header>
            <div class="card-header">
              <span>内存使用率</span>
              <el-icon class="metric-icon memory"><Cpu /></el-icon>
            </div>
          </template>
          <div class="metric-value">{{ memoryUsage }}%</div>
          <div class="metric-trend" :class="memoryTrend > 0 ? 'up' : 'down'">
            <el-icon v-if="memoryTrend > 0"><Top /></el-icon>
            <el-icon v-else><Bottom /></el-icon>
            {{ Math.abs(memoryTrend) }}%
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="metric-card">
          <template #header>
            <div class="card-header">
              <span>网络I/O</span>
              <el-icon class="metric-icon network"><Connection /></el-icon>
            </div>
          </template>
          <div class="metric-value">{{ networkIO }} MB/s</div>
          <div class="metric-trend" :class="networkTrend > 0 ? 'up' : 'down'">
            <el-icon v-if="networkTrend > 0"><Top /></el-icon>
            <el-icon v-else><Bottom /></el-icon>
            {{ Math.abs(networkTrend) }}%
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="metric-card">
          <template #header>
            <div class="card-header">
              <span>磁盘使用率</span>
              <el-icon class="metric-icon disk"><Coin /></el-icon>
            </div>
          </template>
          <div class="metric-value">{{ diskUsage }}%</div>
          <div class="metric-trend" :class="diskTrend > 0 ? 'up' : 'down'">
            <el-icon v-if="diskTrend > 0"><Top /></el-icon>
            <el-icon v-else><Bottom /></el-icon>
            {{ Math.abs(diskTrend) }}%
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>CPU/内存趋势</span>
          </template>
          <div class="chart-container" ref="cpuMemoryChart"></div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>网络流量趋势</span>
          </template>
          <div class="chart-container" ref="networkChart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="table-row">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Pod状态</span>
              <el-button type="primary" size="small" @click="fetchPodMetrics">刷新</el-button>
            </div>
          </template>
          <el-table :data="podMetrics" style="width: 100%">
            <el-table-column prop="pod" label="Pod名称" width="300" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="cpu" label="CPU使用率" width="150">
              <template #default="scope">
                <el-progress :percentage="scope.row.cpu" :color="getProgressColor(scope.row.cpu)" />
              </template>
            </el-table-column>
            <el-table-column prop="memory" label="内存使用率" width="150">
              <template #default="scope">
                <el-progress :percentage="scope.row.memory" :color="getProgressColor(scope.row.memory)" />
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getStatusType(scope.row.status)">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="restarts" label="重启次数" width="100" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { prometheusApi } from '../api'
import { ElMessage } from 'element-plus'
import { Refresh, Cpu, Coin, Connection, Top, Bottom } from '@element-plus/icons-vue'

const selectedNamespace = ref('')
const timeRange = ref<[Date, Date]>([
  new Date(Date.now() - 3600000),
  new Date()
])

const cpuUsage = ref(0)
const memoryUsage = ref(0)
const networkIO = ref(0)
const diskUsage = ref(0)

const cpuTrend = ref(0)
const memoryTrend = ref(0)
const networkTrend = ref(0)
const diskTrend = ref(0)

const podMetrics = ref<any[]>([])

const cpuMemoryChart = ref<HTMLElement>()
const networkChart = ref<HTMLElement>()
let cpuMemoryChartInstance: echarts.ECharts | null = null
let networkChartInstance: echarts.ECharts | null = null
let refreshInterval: ReturnType<typeof setInterval> | null = null

const fetchMetrics = async () => {
  try {
    const start = timeRange.value[0].toISOString()
    const end = timeRange.value[1].toISOString()
    const step = '60s'

    const [cpuRes, memRes, netRes, diskRes] = await Promise.all([
      prometheusApi.queryRange('node_cpu_usage', start, end, step),
      prometheusApi.queryRange('node_memory_usage', start, end, step),
      prometheusApi.queryRange('node_network_io', start, end, step),
      prometheusApi.queryRange('node_disk_usage', start, end, step)
    ])

    if (cpuRes.data?.result) {
      const values = cpuRes.data.result[0]?.values || []
      if (values.length > 0) {
        const latest = parseFloat(values[values.length - 1][1])
        const previous = parseFloat(values[values.length - 2][1])
        cpuUsage.value = Math.round(latest * 100)
        cpuTrend.value = Math.round((latest - previous) / previous * 100)
        updateCpuMemoryChart(values)
      }
    }

    if (memRes.data?.result) {
      const values = memRes.data.result[0]?.values || []
      if (values.length > 0) {
        const latest = parseFloat(values[values.length - 1][1])
        const previous = parseFloat(values[values.length - 2][1])
        memoryUsage.value = Math.round(latest * 100)
        memoryTrend.value = Math.round((latest - previous) / previous * 100)
      }
    }

    if (netRes.data?.result) {
      const values = netRes.data.result[0]?.values || []
      if (values.length > 0) {
        const latest = parseFloat(values[values.length - 1][1])
        const previous = parseFloat(values[values.length - 2][1])
        networkIO.value = Math.round(latest)
        networkTrend.value = Math.round((latest - previous) / previous * 100)
        updateNetworkChart(values)
      }
    }

    if (diskRes.data?.result) {
      const values = diskRes.data.result[0]?.values || []
      if (values.length > 0) {
        const latest = parseFloat(values[values.length - 1][1])
        const previous = parseFloat(values[values.length - 2][1])
        diskUsage.value = Math.round(latest * 100)
        diskTrend.value = Math.round((latest - previous) / previous * 100)
      }
    }
  } catch (error) {
    ElMessage.error('获取监控数据失败')
    console.error(error)
  }
}

const fetchPodMetrics = async () => {
  try {
    const query = selectedNamespace.value 
      ? `pod_cpu_memory_usage{namespace="${selectedNamespace.value}"}` 
      : 'pod_cpu_memory_usage'
    const res = await prometheusApi.query(query)
    if (res.data?.result) {
      podMetrics.value = res.data.result.map((item: any) => ({
        pod: item.metric.pod,
        namespace: item.metric.namespace,
        cpu: Math.round(parseFloat(item.value[1]) * 100),
        memory: Math.round(parseFloat(item.value[2]) * 100),
        status: item.metric.status || 'Running',
        restarts: parseInt(item.metric.restarts) || 0
      }))
    }
  } catch (error) {
    ElMessage.error('获取Pod指标失败')
    console.error(error)
  }
}

const initCharts = () => {
  if (cpuMemoryChart.value) {
    cpuMemoryChartInstance = echarts.init(cpuMemoryChart.value)
  }
  if (networkChart.value) {
    networkChartInstance = echarts.init(networkChart.value)
  }
}

const updateCpuMemoryChart = (values: any[]) => {
  if (!cpuMemoryChartInstance) return

  const times = values.map((v: any) => new Date(v[0] * 1000).toLocaleTimeString())
  const data = values.map((v: any) => parseFloat(v[1]) * 100)

  cpuMemoryChartInstance.setOption({
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: times
    },
    yAxis: {
      type: 'value',
      name: '使用率 (%)'
    },
    series: [{
      name: 'CPU使用率',
      type: 'line',
      smooth: true,
      data: data,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(64, 158, 255, 0.5)' },
          { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
        ])
      }
    }]
  })
}

const updateNetworkChart = (values: any[]) => {
  if (!networkChartInstance) return

  const times = values.map((v: any) => new Date(v[0] * 1000).toLocaleTimeString())
  const data = values.map((v: any) => parseFloat(v[1]))

  networkChartInstance.setOption({
    tooltip: {
      trigger: 'axis'
    },
    xAxis: {
      type: 'category',
      data: times
    },
    yAxis: {
      type: 'value',
      name: '流量 (MB/s)'
    },
    series: [{
      name: '网络流量',
      type: 'line',
      smooth: true,
      data: data,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(103, 194, 58, 0.5)' },
          { offset: 1, color: 'rgba(103, 194, 58, 0.1)' }
        ])
      }
    }]
  })
}

const getProgressColor = (percentage: number) => {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

const getStatusType = (status: string) => {
  switch (status) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Failed':
      return 'danger'
    default:
      return 'info'
  }
}

onMounted(() => {
  initCharts()
  fetchMetrics()
  fetchPodMetrics()
  refreshInterval = setInterval(() => {
    fetchMetrics()
    fetchPodMetrics()
  }, 30000)
})

onBeforeUnmount(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  cpuMemoryChartInstance?.dispose()
  networkChartInstance?.dispose()
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

.filters {
  margin-bottom: 20px;
}

.metrics-row {
  margin-bottom: 20px;
}

.metric-card {
  text-align: center;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-icon {
  font-size: 24px;
}

.metric-icon.cpu {
  color: #409eff;
}

.metric-icon.memory {
  color: #67c23a;
}

.metric-icon.network {
  color: #e6a23c;
}

.metric-icon.disk {
  color: #909399;
}

.metric-value {
  font-size: 32px;
  font-weight: bold;
  margin: 10px 0;
}

.metric-trend {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.metric-trend.up {
  color: #f56c6c;
}

.metric-trend.down {
  color: #67c23a;
}

.charts-row {
  margin-bottom: 20px;
}

.chart-container {
  height: 300px;
}

.table-row {
  margin-bottom: 20px;
}
</style>