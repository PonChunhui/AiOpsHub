<template>
  <div class="terminal-page">
    <!-- 左侧主机列表 -->
    <div class="host-sidebar">
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
    </div>

    <!-- 右侧终端区域 -->
    <div class="terminal-main">
      <!-- 顶部工具栏 -->
      <div class="terminal-toolbar">
        <div class="toolbar-left">
          <el-button @click="closeWindow" size="small">
            <el-icon><Close /></el-icon>
            关闭窗口
          </el-button>
        </div>
        <div class="toolbar-right" v-if="activeTab">
          <el-tag :type="activeTab.status === 'connected' ? 'success' : activeTab.status === 'error' ? 'danger' : 'info'" size="small">
            {{ statusText(activeTab.status) }}
          </el-tag>
          <el-button @click="reconnectTab(activeTab)" size="small" :disabled="activeTab.status === 'connected'">
            <el-icon><Refresh /></el-icon>
            重连
          </el-button>
        </div>
      </div>

      <!-- 选项卡栏 -->
      <div class="tab-bar" v-if="tabs.length > 0">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          class="tab-item"
          :class="{ active: activeTabId === tab.id }"
          @click="switchTab(tab.id)"
        >
          <el-icon class="tab-icon"><Monitor /></el-icon>
          <span class="tab-title">{{ tab.host.name }}</span>
          <span class="tab-ip">{{ tab.host.ip }}</span>
          <el-icon class="tab-close" @click.stop="closeTab(tab.id)"><Close /></el-icon>
        </div>
      </div>

      <!-- 终端容器区域 -->
      <div class="terminals-wrapper">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          class="terminal-container"
          :class="{ hidden: activeTabId !== tab.id }"
          :ref="el => setTerminalRef(tab.id, el as HTMLElement)"
        ></div>
        <!-- 空状态 -->
        <div v-if="tabs.length === 0" class="terminal-placeholder">
          <el-icon :size="64"><Platform /></el-icon>
          <p>请从左侧主机列表选择要连接的主机</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Close, Monitor, Refresh, Platform, Folder } from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import api from '@/api'

// ========== 类型定义 ==========
interface TabContext {
  id: string
  host: any
  terminal: Terminal
  fitAddon: FitAddon
  ws: WebSocket | null
  status: 'connecting' | 'connected' | 'disconnected' | 'error'
}

// ========== 选项卡状态 ==========
const tabs = ref<TabContext[]>([])
const activeTabId = ref<string>('')
const terminalRefs = new Map<string, HTMLElement>()
let tabIdCounter = 0

const activeTab = () => tabs.value.find(t => t.id === activeTabId.value)

// ========== 主机列表数据 ==========
const hostGroups = ref<any[]>([])
const hostTreeData = ref<any[]>([])

// ========== 工具函数 ==========
const statusText = (status: string) => {
  const map: Record<string, string> = {
    connecting: '连接中...',
    connected: '已连接',
    disconnected: '已断开',
    error: '连接错误',
  }
  return map[status] || '未连接'
}

const closeWindow = () => {
  window.close()
}

const setTerminalRef = (tabId: string, el: HTMLElement | null) => {
  if (el) {
    terminalRefs.set(tabId, el)
  } else {
    terminalRefs.delete(tabId)
  }
}

// ========== 主机列表加载 ==========
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

// ========== 节点点击 ==========
const handleNodeClick = (data: any) => {
  if (!data.isHost || !data.hostData) return

  // 如果已有该主机的选项卡，直接切换
  const existing = tabs.value.find(t => t.host.id === data.hostData.id)
  if (existing) {
    switchTab(existing.id)
    return
  }

  // 创建新选项卡
  createTab(data.hostData)
}

// ========== 选项卡管理 ==========
const createTab = async (host: any) => {
  const tabId = `tab-${++tabIdCounter}`

  const tab: TabContext = {
    id: tabId,
    host,
    terminal: null as any,
    fitAddon: null as any,
    ws: null,
    status: 'disconnected',
  }

  tabs.value.push(tab)
  activeTabId.value = tabId

  await nextTick()

  const containerEl = terminalRefs.get(tabId)
  if (!containerEl) return

  // 初始化终端
  const terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: "'Courier New', 'Menlo', 'DejaVu Sans Mono', monospace",
    theme: {
      background: '#1e1e1e',
      foreground: '#ffffff',
      cursor: '#ffffff',
      selectionBackground: '#264f78',
    },
    allowProposedApi: true,
  })

  const fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  containerEl.innerHTML = ''
  terminal.open(containerEl)

  tab.terminal = terminal
  tab.fitAddon = fitAddon

  // 终端输入 → WebSocket
  terminal.onData((data) => {
    if (tab.ws?.readyState === WebSocket.OPEN) {
      tab.ws.send(JSON.stringify({ type: 'data', data }))
    }
  })

  terminal.onResize(({ cols, rows }) => {
    if (tab.ws?.readyState === WebSocket.OPEN) {
      tab.ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })

  // 建立 WebSocket 连接
  connectTabWebSocket(tab)

  // 首次适配尺寸
  nextTick(() => fitAddon.fit())
}

const connectTabWebSocket = (tab: TabContext) => {
  const { host, terminal } = tab

  tab.status = 'connecting'
  terminal.clear()

  const token = localStorage.getItem('token')
  const wsUrl = `ws://localhost:8080/ws/ssh/${host.id}?token=${token}`

  const ws = new WebSocket(wsUrl)
  tab.ws = ws

  ws.onopen = () => {
    tab.status = 'connected'
    terminal.write('\r\n\x1b[32m[已连接到 ' + host.name + ' (' + host.ip + ')]\x1b[0m\r\n')
    nextTick(() => tab.fitAddon.fit())
  }

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'data') {
      terminal.write(data.data)
    } else if (data.type === 'error') {
      terminal.write('\r\n\x1b[31m[错误: ' + data.data + ']\x1b[0m\r\n')
    } else if (data.type === 'connected') {
      terminal.write('\r\n\x1b[32m[' + data.data + ']\x1b[0m\r\n')
    }
  }

  ws.onerror = () => {
    tab.status = 'error'
    terminal.write('\r\n\x1b[31m[WebSocket连接错误]\x1b[0m\r\n')
  }

  ws.onclose = () => {
    tab.status = 'disconnected'
    terminal.write('\r\n\x1b[33m[连接已关闭]\x1b[0m\r\n')
  }
}

const switchTab = (tabId: string) => {
  activeTabId.value = tabId
  // 切换后重新适配终端尺寸
  nextTick(() => {
    const tab = tabs.value.find(t => t.id === tabId)
    if (tab?.fitAddon) {
      tab.fitAddon.fit()
      tab.terminal.focus()
    }
  })
}

const closeTab = (tabId: string) => {
  const idx = tabs.value.findIndex(t => t.id === tabId)
  if (idx === -1) return

  const tab = tabs.value[idx]

  // 清理资源
  if (tab.ws) {
    tab.ws.close()
    tab.ws = null
  }
  tab.terminal.dispose()
  terminalRefs.delete(tabId)

  tabs.value.splice(idx, 1)

  // 如果关闭的是当前激活的选项卡，切换到相邻选项卡
  if (activeTabId.value === tabId) {
    if (tabs.value.length > 0) {
      const nextIdx = Math.min(idx, tabs.value.length - 1)
      switchTab(tabs.value[nextIdx].id)
    } else {
      activeTabId.value = ''
    }
  }
}

const reconnectTab = (tab: TabContext) => {
  if (tab.ws) {
    tab.ws.close()
    tab.ws = null
  }
  tab.status = 'connecting'
  connectTabWebSocket(tab)
}

// ========== 窗口 resize ==========
const handleResize = () => {
  const tab = activeTab()
  if (tab?.fitAddon) {
    tab.fitAddon.fit()
  }
}

// ========== 生命周期 ==========
onMounted(async () => {
  await loadHostGroups()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  // 清理所有选项卡
  for (const tab of tabs.value) {
    if (tab.ws) tab.ws.close()
    tab.terminal.dispose()
  }
  tabs.value = []
})
</script>

<style scoped>
.terminal-page {
  height: 100vh;
  width: 100vw;
  display: flex;
  background: #1e1e1e;
  overflow: hidden;
}

/* ========== 左侧主机列表 ========== */
.host-sidebar {
  width: 260px;
  background: #252526;
  border-right: 1px solid #404040;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #404040;
}

.sidebar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e0e0e0;
  font-size: 14px;
  font-weight: 500;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.sidebar-content :deep(.el-tree) {
  background: transparent;
}

.sidebar-content :deep(.el-tree-node__content) {
  height: 36px;
  padding: 0 12px;
}

.sidebar-content :deep(.el-tree-node__content:hover) {
  background: #2a2d2e;
}

.sidebar-content :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background: #094771;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
}

.tree-node .el-icon {
  color: #dcb67a;
  flex-shrink: 0;
}

.tree-node.is-host .el-icon {
  color: #4ec9b0;
}

.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #cccccc;
}

.tree-node.is-host .node-label {
  color: #d4d4d4;
}

.host-status {
  margin-left: auto;
  flex-shrink: 0;
}

/* ========== 右侧终端区域 ========== */
.terminal-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ========== 工具栏 ========== */
.terminal-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #2d2d2d;
  border-bottom: 1px solid #404040;
  flex-shrink: 0;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* ========== 选项卡栏 ========== */
.tab-bar {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #404040;
  flex-shrink: 0;
  overflow-x: auto;
  scrollbar-width: thin;
}

.tab-bar::-webkit-scrollbar {
  height: 3px;
}

.tab-bar::-webkit-scrollbar-thumb {
  background: #555;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  cursor: pointer;
  border-right: 1px solid #333;
  white-space: nowrap;
  font-size: 13px;
  color: #888;
  transition: all 0.15s ease;
  position: relative;
}

.tab-item:hover {
  background: #2d2d2d;
  color: #ccc;
}

.tab-item.active {
  background: #1e1e1e;
  color: #e0e0e0;
}

.tab-item.active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: #3b82f6;
}

.tab-icon {
  font-size: 14px;
  color: #4ec9b0;
}

.tab-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-ip {
  color: #666;
  font-size: 12px;
}

.tab-close {
  font-size: 12px;
  padding: 2px;
  border-radius: 4px;
  color: #666;
  transition: all 0.15s ease;
}

.tab-close:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #e0e0e0;
}

/* ========== 终端容器 ========== */
.terminals-wrapper {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.terminal-container {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 4px;
}

.terminal-container.hidden {
  display: none;
}

.terminal-container :deep(.xterm) {
  height: 100%;
}

.terminal-container :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

/* ========== 空状态 ========== */
.terminal-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
}

.terminal-placeholder .el-icon {
  margin-bottom: 16px;
  color: #555;
}

.terminal-placeholder p {
  margin: 0;
  font-size: 14px;
}
</style>
