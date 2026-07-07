<template>
  <div class="mcp-tool-selector">
    <el-collapse v-model="activeServers" @change="handleServerChange">
      <el-collapse-item
        v-for="server in servers"
        :key="server.id"
        :name="server.id"
        :title="server.name"
      >
        <template #title>
          <div class="server-title">
            <el-icon><Connection /></el-icon>
            <span>{{ server.name }}</span>
            <el-tag :type="server.status === 'active' ? 'success' : 'info'" size="small">
              {{ server.status }}
            </el-tag>
            <el-tag v-if="serverTools[server.id]" type="primary" size="small">
              {{ serverTools[server.id].length }} 工具
            </el-tag>
          </div>
        </template>

        <div v-loading="toolsLoading[server.id]" class="tools-container">
          <div v-if="serverTools[server.id]" class="tools-list">
            <div
              v-for="tool in serverTools[server.id]"
              :key="tool.name"
              class="tool-item"
              :class="{ selected: selectedTools.includes(tool.name) }"
              @click="toggleTool(tool.name)"
            >
              <div class="tool-header">
                <el-checkbox :model-value="selectedTools.includes(tool.name)" />
                <span class="tool-name">{{ tool.name }}</span>
              </div>
              <div class="tool-description">{{ tool.description }}</div>
              <div v-if="tool.inputSchema?.properties" class="tool-params">
                <el-tag size="small" type="info">
                  {{ Object.keys(tool.inputSchema.properties).length }} 参数
                </el-tag>
                <el-button
                  size="small"
                  text
                  @click.stop="showToolParams(tool)"
                >
                  查看参数
                </el-button>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无工具" />
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- 工具参数详情对话框 -->
    <el-dialog
      v-model="paramsDialogVisible"
      :title="currentTool?.name + ' 参数'"
      width="600px"
    >
      <div v-if="currentTool?.inputSchema" class="params-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="参数类型">
            {{ currentTool.inputSchema.type || 'object' }}
          </el-descriptions-item>
          <el-descriptions-item
            v-for="(prop, key) in currentTool.inputSchema.properties"
            :key="key"
            :label="key"
          >
            <div>
              <el-tag size="small">{{ prop.type || 'any' }}</el-tag>
              <span v-if="prop.description" class="param-desc">{{ prop.description }}</span>
              <el-tag v-if="currentTool.inputSchema.required?.includes(key)" type="danger" size="small">
                必填
              </el-tag>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 选中工具统计 -->
    <div v-if="selectedTools.length > 0" class="selected-summary">
      <el-alert type="success" :closable="false">
        已选择 {{ selectedTools.length }} 个工具，AI 将自动使用这些工具处理请求
      </el-alert>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Connection } from '@element-plus/icons-vue'
import { mcpApi } from '@/api/index'

interface Props {
  sessionId?: string
}

const props = defineProps<Props>()
const emit = defineEmits(['update:selectedTools'])

const servers = ref<any[]>([])
const serverTools = ref<Record<string, any[]>>({})
const toolsLoading = ref<Record<string, boolean>>({})
const activeServers = ref<string[]>([])
const selectedTools = ref<string[]>([])
const paramsDialogVisible = ref(false)
const currentTool = ref<any>(null)

const loadServers = async () => {
  try {
    const res = await mcpApi.listServers()
    if (res.code === 200) {
      servers.value = res.servers || []
      const activeServerIds = servers.value
        .filter(s => s.status === 'active')
        .map(s => s.id)
      activeServers.value = activeServerIds.slice(0, 2)
    }
  } catch (error: any) {
    ElMessage.error('加载 MCP Server 失败: ' + error.message)
  }
}

const loadServerTools = async (serverId: string) => {
  if (serverTools.value[serverId]) return

  toolsLoading.value[serverId] = true
  try {
    const res = await mcpApi.getServerTools(serverId)
    if (res.code === 200) {
      serverTools.value[serverId] = res.tools || []
    }
  } catch (error: any) {
    ElMessage.error('加载工具失败: ' + error.message)
  } finally {
    toolsLoading.value[serverId] = false
  }
}

const handleServerChange = (serverIds: string[]) => {
  serverIds.forEach(id => loadServerTools(id))
}

const toggleTool = (toolName: string) => {
  const index = selectedTools.value.indexOf(toolName)
  if (index > -1) {
    selectedTools.value.splice(index, 1)
  } else {
    selectedTools.value.push(toolName)
  }
  emit('update:selectedTools', selectedTools.value)
}

const showToolParams = (tool: any) => {
  currentTool.value = tool
  paramsDialogVisible.value = true
}

watch(selectedTools, (newVal) => {
  emit('update:selectedTools', newVal)
})

onMounted(() => {
  loadServers()
})
</script>

<style scoped>
.mcp-tool-selector {
  padding: 10px;
  max-height: 400px;
}

.server-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 500;
}

.server-title .el-icon {
  color: var(--el-color-primary);
}

.tools-container {
  padding: 10px;
}

.tools-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tool-item {
  padding: 12px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.tool-item:hover {
  border-color: var(--el-color-primary);
  background-color: var(--el-fill-color-light);
}

.tool-item.selected {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.tool-name {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.tool-description {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.tool-params {
  display: flex;
  align-items: center;
  gap: 10px;
}

.param-desc {
  margin-left: 10px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.selected-summary {
  margin-top: 15px;
}

.params-detail {
  padding: 10px;
}
</style>