<template>
  <div>
    <el-row :gutter="20" class="header">
      <el-col :span="24">
        <h1>Kubernetes管理</h1>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="Pods" name="pods">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-form :inline="true">
                <el-form-item label="命名空间">
                  <el-select v-model="podFilter.namespace" placeholder="选择命名空间" @change="fetchPods">
                    <el-option label="全部" value="" />
                    <el-option label="default" value="default" />
                    <el-option label="kube-system" value="kube-system" />
                    <el-option label="monitoring" value="monitoring" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-input v-model="podFilter.search" placeholder="搜索Pod" clearable>
                    <template #prefix>
                      <el-icon><Search /></el-icon>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="fetchPods">刷新</el-button>
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="filteredPods" style="width: 100%">
            <el-table-column type="selection" width="55" />
            <el-table-column prop="name" label="名称" width="300">
              <template #default="scope">
                <el-link type="primary" @click="viewPodDetail(scope.row)">{{ scope.row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="scope">
                <el-tag :type="getPodStatusType(scope.row.status)">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ready" label="就绪" width="100" />
            <el-table-column prop="restarts" label="重启次数" width="100" />
            <el-table-column prop="cpu" label="CPU" width="120">
              <template #default="scope">
                {{ scope.row.cpu }}m
              </template>
            </el-table-column>
            <el-table-column prop="memory" label="内存" width="120">
              <template #default="scope">
                {{ scope.row.memory }}Mi
              </template>
            </el-table-column>
            <el-table-column prop="age" label="运行时长" width="120" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="scope">
                <el-button type="primary" size="small" text @click="viewPodLogs(scope.row)">
                  日志
                </el-button>
                <el-button type="warning" size="small" text @click="restartPod(scope.row)">
                  重启
                </el-button>
                <el-button type="danger" size="small" text @click="deletePod(scope.row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Deployments" name="deployments">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-form :inline="true">
                <el-form-item label="命名空间">
                  <el-select v-model="deploymentFilter.namespace" placeholder="选择命名空间" @change="fetchDeployments">
                    <el-option label="全部" value="" />
                    <el-option label="default" value="default" />
                    <el-option label="kube-system" value="kube-system" />
                    <el-option label="monitoring" value="monitoring" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="fetchDeployments">刷新</el-button>
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="deployments" style="width: 100%">
            <el-table-column prop="name" label="名称" width="300">
              <template #default="scope">
                <el-link type="primary" @click="viewDeploymentDetail(scope.row)">{{ scope.row.name }}</el-link>
              </template>
            </el-table-column>
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="replicas" label="副本数" width="120">
              <template #default="scope">
                {{ scope.row.readyReplicas }} / {{ scope.row.replicas }}
              </template>
            </el-table-column>
            <el-table-column prop="available" label="可用" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.available ? 'success' : 'danger'">
                  {{ scope.row.available ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="strategy" label="更新策略" width="120" />
            <el-table-column prop="age" label="运行时长" width="120" />
            <el-table-column label="操作" width="250" fixed="right">
              <template #default="scope">
                <el-button type="primary" size="small" text @click="scaleDeployment(scope.row)">
                  扩缩容
                </el-button>
                <el-button type="warning" size="small" text @click="restartDeployment(scope.row)">
                  重启
                </el-button>
                <el-button type="info" size="small" text @click="viewDeploymentYaml(scope.row)">
                  YAML
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Services" name="services">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-form :inline="true">
                <el-form-item label="命名空间">
                  <el-select v-model="serviceFilter.namespace" placeholder="选择命名空间" @change="fetchServices">
                    <el-option label="全部" value="" />
                    <el-option label="default" value="default" />
                    <el-option label="kube-system" value="kube-system" />
                    <el-option label="monitoring" value="monitoring" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="fetchServices">刷新</el-button>
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="services" style="width: 100%">
            <el-table-column prop="name" label="名称" width="250" />
            <el-table-column prop="namespace" label="命名空间" width="150" />
            <el-table-column prop="type" label="类型" width="120" />
            <el-table-column prop="clusterIP" label="ClusterIP" width="150" />
            <el-table-column prop="ports" label="端口" width="200">
              <template #default="scope">
                <div v-for="(port, index) in scope.row.ports" :key="index">
                  {{ port.port }}:{{ port.targetPort }}/{{ port.protocol }}
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="selector" label="选择器">
              <template #default="scope">
                <el-tag v-for="(value, key, index) in scope.row.selector" :key="index" size="small" style="margin-right: 5px">
                  {{ key }}={{ value }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="Events" name="events">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-form :inline="true">
                <el-form-item label="命名空间">
                  <el-select v-model="eventFilter.namespace" placeholder="选择命名空间" @change="fetchEvents">
                    <el-option label="全部" value="" />
                    <el-option label="default" value="default" />
                    <el-option label="kube-system" value="kube-system" />
                    <el-option label="monitoring" value="monitoring" />
                  </el-select>
                </el-form-item>
                <el-form-item label="类型">
                  <el-select v-model="eventFilter.type" placeholder="全部" @change="fetchEvents">
                    <el-option label="全部" value="" />
                    <el-option label="Normal" value="Normal" />
                    <el-option label="Warning" value="Warning" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="fetchEvents">刷新</el-button>
                </el-form-item>
              </el-form>
            </div>
          </template>

          <el-table :data="events" style="width: 100%">
            <el-table-column prop="lastSeen" label="时间" width="180">
              <template #default="scope">
                {{ formatTime(scope.row.lastSeen) }}
              </template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="100">
              <template #default="scope">
                <el-tag :type="scope.row.type === 'Warning' ? 'warning' : 'info'">
                  {{ scope.row.type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" width="150" />
            <el-table-column prop="object" label="对象" width="250" />
            <el-table-column prop="message" label="消息" />
            <el-table-column prop="count" label="次数" width="80" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="podDetailVisible" title="Pod详情" width="70%">
      <el-tabs v-model="podDetailTab">
        <el-tab-pane label="详情" name="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="名称">{{ selectedPod?.name }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ selectedPod?.namespace }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getPodStatusType(selectedPod?.status)">{{ selectedPod?.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="节点">{{ selectedPod?.node }}</el-descriptions-item>
            <el-descriptions-item label="Pod IP">{{ selectedPod?.podIP }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatTime(selectedPod?.createdAt) }}</el-descriptions-item>
          </el-descriptions>

          <h4 style="margin-top: 20px">容器</h4>
          <el-table :data="selectedPod?.containers || []" style="width: 100%">
            <el-table-column prop="name" label="名称" width="200" />
            <el-table-column prop="image" label="镜像" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="scope">
                <el-tag :type="scope.row.status === 'Running' ? 'success' : 'warning'">
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="restartCount" label="重启次数" width="100" />
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <pre class="yaml-content">{{ selectedPod?.yaml }}</pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="30%">
      <el-form label-width="100px">
        <el-form-item label="Deployment">
          <el-input :value="selectedDeployment?.name" disabled />
        </el-form-item>
        <el-form-item label="当前副本数">
          <el-input :value="selectedDeployment?.replicas" disabled />
        </el-form-item>
        <el-form-item label="目标副本数">
          <el-input-number v-model="targetReplicas" :min="0" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmScale">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logDialogVisible" :title="`日志 - ${logPodName}`" width="70%">
      <el-form :inline="true" style="margin-bottom: 15px">
        <el-form-item label="容器">
          <el-select v-model="selectedContainer" placeholder="选择容器">
            <el-option v-for="c in podContainers" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="行数">
          <el-input-number v-model="logLines" :min="100" :max="10000" :step="100" />
        </el-form-item>
        <el-form-item>
          <el-switch v-model="followLogs" active-text="实时" @change="toggleFollowLogs" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchPodLogs">刷新</el-button>
        </el-form-item>
      </el-form>
      <pre class="log-content">{{ podLogs }}</pre>
    </el-dialog>

    <el-dialog v-model="yamlDialogVisible" title="YAML" width="70%">
      <pre class="yaml-content">{{ deploymentYaml }}</pre>
      <template #footer>
        <el-button @click="yamlDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyYaml">复制</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { k8sApi } from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

const activeTab = ref('pods')

const podFilter = ref({
  namespace: '',
  search: ''
})

const deploymentFilter = ref({
  namespace: ''
})

const serviceFilter = ref({
  namespace: ''
})

const eventFilter = ref({
  namespace: '',
  type: ''
})

const pods = ref<any[]>([])
const deployments = ref<any[]>([])
const services = ref<any[]>([])
const events = ref<any[]>([])

const filteredPods = computed(() => {
  if (!podFilter.value.search) return pods.value
  const search = podFilter.value.search.toLowerCase()
  return pods.value.filter(p => p.name.toLowerCase().includes(search))
})

const podDetailVisible = ref(false)
const selectedPod = ref<any>(null)
const podDetailTab = ref('detail')

const scaleDialogVisible = ref(false)
const selectedDeployment = ref<any>(null)
const targetReplicas = ref(1)

const logDialogVisible = ref(false)
const logPodName = ref('')
const selectedContainer = ref('')
const podContainers = ref<string[]>([])
const logLines = ref(500)
const followLogs = ref(false)
const podLogs = ref('')
let logInterval: ReturnType<typeof setInterval> | null = null

const yamlDialogVisible = ref(false)
const deploymentYaml = ref('')

const fetchPods = async () => {
  try {
    const res = await k8sApi.getPods(podFilter.value.namespace)
    if (res.data?.pods) {
      pods.value = res.data.pods.map((p: any) => ({
        name: p.metadata.name,
        namespace: p.metadata.namespace,
        status: p.status.phase,
        ready: `${p.status.containerStatuses?.filter((c: any) => c.ready).length || 0}/${p.spec.containers.length}`,
        restarts: p.status.containerStatuses?.reduce((sum: number, c: any) => sum + c.restartCount, 0) || 0,
        cpu: p.metrics?.cpu || 0,
        memory: p.metrics?.memory || 0,
        age: getAge(p.metadata.creationTimestamp),
        node: p.spec.nodeName,
        podIP: p.status.podIP,
        createdAt: p.metadata.creationTimestamp,
        containers: p.spec.containers?.map((c: any, i: number) => ({
          name: c.name,
          image: c.image,
          status: p.status.containerStatuses?.[i]?.state?.running ? 'Running' : 'Not Running',
          restartCount: p.status.containerStatuses?.[i]?.restartCount || 0
        })) || [],
        yaml: ''
      }))
    }
  } catch (error) {
    ElMessage.error('获取Pod列表失败')
    console.error(error)
  }
}

const fetchDeployments = async () => {
  try {
    const res = await k8sApi.getDeployments(deploymentFilter.value.namespace)
    if (res.data?.deployments) {
      deployments.value = res.data.deployments.map((d: any) => ({
        name: d.metadata.name,
        namespace: d.metadata.namespace,
        replicas: d.spec.replicas,
        readyReplicas: d.status.readyReplicas || 0,
        available: d.status.availableReplicas > 0,
        strategy: d.spec.strategy.type,
        age: getAge(d.metadata.creationTimestamp),
        yaml: ''
      }))
    }
  } catch (error) {
    ElMessage.error('获取Deployment列表失败')
    console.error(error)
  }
}

const fetchServices = async () => {
  try {
    const res = await k8sApi.getServices(serviceFilter.value.namespace)
    if (res.data?.services) {
      services.value = res.data.services.map((s: any) => ({
        name: s.metadata.name,
        namespace: s.metadata.namespace,
        type: s.spec.type,
        clusterIP: s.spec.clusterIP,
        ports: s.spec.ports || [],
        selector: s.spec.selector || {}
      }))
    }
  } catch (error) {
    ElMessage.error('获取Service列表失败')
    console.error(error)
  }
}

const fetchEvents = async () => {
  try {
    const res = await k8sApi.getEvents(eventFilter.value.namespace, eventFilter.value.type)
    if (res.data?.events) {
      events.value = res.data.events.map((e: any) => ({
        lastSeen: e.lastTimestamp,
        type: e.type,
        reason: e.reason,
        object: `${e.involvedObject.kind}/${e.involvedObject.name}`,
        message: e.message,
        count: e.count
      }))
    }
  } catch (error) {
    ElMessage.error('获取事件列表失败')
    console.error(error)
  }
}

const viewPodDetail = async (pod: any) => {
  selectedPod.value = pod
  try {
    const res = await k8sApi.getPodYaml(pod.namespace, pod.name)
    selectedPod.value.yaml = res.data?.yaml || ''
  } catch (error) {
    console.error(error)
  }
  podDetailVisible.value = true
}

const restartPod = async (pod: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要重启Pod "${pod.name}" 吗？`,
      '确认重启',
      { confirmButtonText: '重启', cancelButtonText: '取消', type: 'warning' }
    )
    await k8sApi.deletePod(pod.namespace, pod.name)
    ElMessage.success('Pod已删除，正在重新创建')
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重启Pod失败')
      console.error(error)
    }
  }
}

const deletePod = async (pod: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除Pod "${pod.name}" 吗？此操作不可逆。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'error' }
    )
    await k8sApi.deletePod(pod.namespace, pod.name)
    ElMessage.success('Pod已删除')
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除Pod失败')
      console.error(error)
    }
  }
}

const viewPodLogs = (pod: any) => {
  logPodName.value = pod.name
  podContainers.value = pod.containers?.map((c: any) => c.name) || []
  selectedContainer.value = podContainers.value[0] || ''
  logDialogVisible.value = true
  fetchPodLogs()
}

const fetchPodLogs = async () => {
  try {
    const res = await k8sApi.getPodLogs(
      selectedPod.value.namespace,
      logPodName.value,
      selectedContainer.value,
      logLines.value
    )
    podLogs.value = res.data?.logs || ''
  } catch (error) {
    ElMessage.error('获取日志失败')
    console.error(error)
  }
}

const toggleFollowLogs = () => {
  if (followLogs.value) {
    logInterval = setInterval(fetchPodLogs, 2000)
  } else {
    if (logInterval) {
      clearInterval(logInterval)
      logInterval = null
    }
  }
}

const viewDeploymentDetail = async (deployment: any) => {
  selectedDeployment.value = deployment
  scaleDialogVisible.value = false
}

const scaleDeployment = (deployment: any) => {
  selectedDeployment.value = deployment
  targetReplicas.value = deployment.replicas
  scaleDialogVisible.value = true
}

const confirmScale = async () => {
  try {
    await k8sApi.scaleDeployment(
      selectedDeployment.value.namespace,
      selectedDeployment.value.name,
      targetReplicas.value
    )
    ElMessage.success('扩缩容成功')
    scaleDialogVisible.value = false
    fetchDeployments()
  } catch (error) {
    ElMessage.error('扩缩容失败')
    console.error(error)
  }
}

const restartDeployment = async (deployment: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要重启Deployment "${deployment.name}" 吗？`,
      '确认重启',
      { confirmButtonText: '重启', cancelButtonText: '取消', type: 'warning' }
    )
    await k8sApi.restartDeployment(deployment.namespace, deployment.name)
    ElMessage.success('重启命令已发送')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重启失败')
      console.error(error)
    }
  }
}

const viewDeploymentYaml = async (deployment: any) => {
  try {
    const res = await k8sApi.getDeploymentYaml(deployment.namespace, deployment.name)
    deploymentYaml.value = res.data?.yaml || ''
    yamlDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取YAML失败')
    console.error(error)
  }
}

const copyYaml = () => {
  navigator.clipboard.writeText(deploymentYaml.value)
  ElMessage.success('已复制到剪贴板')
}

const getPodStatusType = (status: string | undefined) => {
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

const formatTime = (timestamp: string | undefined) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString()
}

const getAge = (timestamp: string) => {
  const created = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - created.getTime()

  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days > 0) return `${days}d`

  const hours = Math.floor(diff / (1000 * 60 * 60))
  if (hours > 0) return `${hours}h`

  const minutes = Math.floor(diff / (1000 * 60))
  return `${minutes}m`
}

onMounted(() => {
  fetchPods()
  fetchDeployments()
  fetchServices()
  fetchEvents()
})

onBeforeUnmount(() => {
  if (logInterval) {
    clearInterval(logInterval)
  }
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

.yaml-content,
.log-content {
  background-color: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}
</style>