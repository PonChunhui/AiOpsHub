<template>
  <div class="file-manage">
    <el-container>
      <el-aside width="260px" class="host-sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">
            <el-icon><Platform /></el-icon>
            主机列表
          </span>
          <el-button size="small" link @click="loadHostGroups">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
        <div class="sidebar-content">
          <el-tree
            :data="hostTreeData"
            node-key="id"
            default-expand-all
            highlight-current
            @node-click="handleNodeClick"
          >
            <template #default="{ data }">
              <div class="tree-node" :class="{ 'is-host': data.isHost }">
                <el-icon v-if="!data.isHost"><Folder /></el-icon>
                <el-icon v-else><Monitor /></el-icon>
                <span class="node-label">{{ data.name }}</span>
                <el-tag v-if="data.isHost" size="small" :type="data.status === 'active' ? 'success' : 'info'" class="host-status">
                  {{ data.ip }}
                </el-tag>
              </div>
            </template>
          </el-tree>
        </div>
      </el-aside>

      <el-main class="file-main">
        <template v-if="currentHost">
          <div class="file-toolbar">
            <div class="toolbar-left">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item
                  v-for="(crumb, index) in breadcrumbs"
                  :key="index"
                >
                  <a
                    v-if="index < breadcrumbs.length - 1"
                    @click.prevent="navigateTo(crumb.path)"
                    href="#"
                    class="breadcrumb-link"
                  >
                    {{ crumb.name }}
                  </a>
                  <span v-else>{{ crumb.name }}</span>
                </el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            <div class="toolbar-right">
              <el-button size="small" @click="loadFiles(currentPath)">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
              <el-button size="small" type="primary" @click="uploadDialogVisible = true">
                <el-icon><Upload /></el-icon>
                上传
              </el-button>
            </div>
          </div>

          <div class="file-content">
            <el-table
              :data="files"
              v-loading="loading"
              stripe
              highlight-current-row
              @row-dblclick="handleRowDblClick"
            >
              <el-table-column label="名称" min-width="300">
                <template #default="{ row }">
                  <div class="file-name" @click="handleFileClick(row)">
                    <el-icon class="file-icon" :class="{ 'is-dir': row.is_dir }">
                      <Folder v-if="row.is_dir" />
                      <Document v-else />
                    </el-icon>
                    <span>{{ row.name }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="size" label="大小" width="120">
                <template #default="{ row }">
                  {{ row.is_dir ? '-' : formatSize(row.size) }}
                </template>
              </el-table-column>
              <el-table-column prop="mode" label="权限" width="120" />
              <el-table-column prop="mod_time" label="修改时间" width="180" />
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="{ row }">
                  <el-button
                    v-if="!row.is_dir"
                    size="small"
                    type="primary"
                    link
                    @click="downloadFile(row)"
                  >
                    下载
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </template>

        <div v-else class="file-placeholder">
          <el-icon :size="64"><Folder /></el-icon>
          <p>请从左侧主机列表选择要管理的主机</p>
        </div>
      </el-main>
    </el-container>

    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="500px">
      <div class="upload-area">
        <p class="upload-path">目标路径: {{ currentPath }}</p>
        <el-upload
          ref="uploadRef"
          :auto-upload="false"
          :limit="1"
          drag
          @change="handleUploadChange"
        >
          <el-icon :size="48"><Upload /></el-icon>
          <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        </el-upload>
      </div>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpload" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Upload, Platform, Folder, Document } from '@element-plus/icons-vue'
import api from '@/api'

interface Breadcrumb {
  name: string
  path: string
}

const route = useRoute()

const currentHost = ref<any>(null)
const currentPath = ref('/')
const files = ref<any[]>([])
const loading = ref(false)
const hostGroups = ref<any[]>([])
const hostTreeData = ref<any[]>([])

const uploadDialogVisible = ref(false)
const uploading = ref(false)
const uploadFile = ref<File | null>(null)

const breadcrumbs = computed<Breadcrumb[]>(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const crumbs: Breadcrumb[] = [{ name: '/', path: '/' }]
  let path = ''
  for (const part of parts) {
    path += '/' + part
    crumbs.push({ name: part, path })
  }
  return crumbs
})

const loadHostGroups = async () => {
  try {
    const res = await api.get('/host-groups')
    if (res?.code === 200) {
      hostGroups.value = res?.data?.groups || []
      await loadAllHosts()
    }
  } catch (error: any) {
    ElMessage.error('加载主机分组失败: ' + error.message)
  }
}

const loadAllHosts = async () => {
  try {
    const res = await api.get('/hosts', { params: { pageSize: 1000 } })
    if (res?.code === 200) {
      const hosts = res?.data?.hosts || []
      hostTreeData.value = buildHostTree(hostGroups.value, hosts)
    }
  } catch (error: any) {
    ElMessage.error('加载主机列表失败: ' + error.message)
  }
}

const buildHostTree = (groups: any[], hosts: any[]) => {
  return groups.map(group => ({
    id: group.id,
    name: group.name,
    isHost: false,
    children: [
      ...(group.children ? buildHostTree(group.children, hosts) : []),
      ...hosts
        .filter(host => host.group_id === group.id)
        .map(host => ({
          id: host.id,
          name: host.name,
          ip: host.ip,
          status: host.status,
          isHost: true,
          hostData: host
        }))
    ]
  }))
}

const handleNodeClick = (data: any) => {
  if (!data.isHost || !data.hostData) return
  currentHost.value = data.hostData
  currentPath.value = '/'
  loadFiles('/')
}

const loadFiles = async (path: string) => {
  if (!currentHost.value) return
  loading.value = true
  try {
    const res = await api.get(`/hosts/${currentHost.value.id}/files`, { params: { path } })
    if (res?.code === 200) {
      files.value = res?.data?.files || []
      currentPath.value = res?.data?.path || path
    }
  } catch (error: any) {
    ElMessage.error('加载文件列表失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const handleFileClick = (row: any) => {
  if (row.is_dir) {
    loadFiles(row.path)
  }
}

const handleRowDblClick = (row: any) => {
  if (row.is_dir) {
    loadFiles(row.path)
  }
}

const navigateTo = (path: string) => {
  loadFiles(path)
}

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const handleUploadChange = (file: any) => {
  uploadFile.value = file.raw
}

const handleUpload = async () => {
  if (!uploadFile.value || !currentHost.value) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('path', currentPath.value + '/' + uploadFile.value.name)
    formData.append('file', uploadFile.value)

    await api.post(`/hosts/${currentHost.value.id}/files/upload`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    ElMessage.success('文件上传成功')
    uploadDialogVisible.value = false
    uploadFile.value = null
    loadFiles(currentPath.value)
  } catch (error: any) {
    ElMessage.error('文件上传失败: ' + error.message)
  } finally {
    uploading.value = false
  }
}

const downloadFile = (row: any) => {
  if (!currentHost.value) return
  const token = localStorage.getItem('token')
  const url = `/api/v1/hosts/${currentHost.value.id}/files/download?path=${encodeURIComponent(row.path)}&token=${token}`
  const a = document.createElement('a')
  a.href = url
  a.download = row.name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

onMounted(async () => {
  await loadHostGroups()

  const hostId = route.params.id as string
  if (hostId) {
    try {
      const res = await api.get('/hosts', { params: { pageSize: 1000 } })
      if (res?.code === 200) {
        const hosts = res?.data?.hosts || []
        const host = hosts.find((h: any) => h.id === hostId)
        if (host) {
          currentHost.value = host
          loadFiles('/')
        }
      }
    } catch {
      ElMessage.warning('获取主机信息失败')
    }
  }
})

onBeforeUnmount(() => {})
</script>

<style scoped>
.file-manage {
  height: calc(100vh - 60px);
}

.host-sidebar {
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
}

.sidebar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
}

.tree-node.is-host .el-icon {
  color: #409eff;
}

.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.host-status {
  margin-left: auto;
  flex-shrink: 0;
}

.file-main {
  padding: 0;
  display: flex;
  flex-direction: column;
}

.file-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #fff;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.breadcrumb-link {
  color: #409eff;
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.file-content {
  flex: 1;
  padding: 0;
  overflow: auto;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.file-icon {
  color: #909399;
}

.file-icon.is-dir {
  color: #e6a23c;
}

.file-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.file-placeholder .el-icon {
  margin-bottom: 16px;
  color: #c0c4cc;
}

.file-placeholder p {
  margin: 0;
  font-size: 14px;
}

.upload-area {
  text-align: center;
}

.upload-path {
  color: #606266;
  margin-bottom: 16px;
}
</style>
