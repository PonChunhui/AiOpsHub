<template>
  <el-card>
    <template #header>
      <div>
        <span>Agent管理</span>
        <el-button @click="loadAgents" size="small" style="margin-left: 10px">
          重新加载
        </el-button>
        <el-tag v-if="loading" type="info" style="margin-left: 10px">加载中...</el-tag>
        <el-tag v-else type="success" style="margin-left: 10px">
          共 {{ agents.length }} 个Agent
        </el-tag>
      </div>
    </template>
    
    <el-alert v-if="error" type="error" :closable="false">
      错误: {{ error }}
    </el-alert>
    
    <el-alert v-if="agents.length === 0 && !loading && !error" type="info" :closable="false">
      暂无数据。后端API数据: 
      <a href="http://localhost:8080/api/v1/agents" target="_blank">点击查看</a>
    </el-alert>
    
    <el-table v-loading="loading" :data="agents" style="width: 100%">
      <el-table-column prop="id" label="ID" width="250" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="type" label="类型" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'">
            {{ row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" />
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { agentApi } from '@/api'
import { ElMessage } from 'element-plus'

const agents = ref<any[]>([])
const loading = ref(false)
const error = ref('')

onMounted(() => {
  console.log('=== Agents页面已挂载 ===')
  loadAgents()
})

const loadAgents = async () => {
  loading.value = true
  error.value = ''
  
  try {
    console.log('开始调用API...')
    const res = await agentApi.list()
    console.log('API返回:', res)
    
    if (res && res.data) {
      agents.value = res.data
      console.log('设置agents数据:', agents.value)
      ElMessage.success(`加载成功，共 ${agents.value.length} 个Agent`)
    } else {
      console.log('API返回格式异常:', res)
      error.value = '数据格式异常'
      ElMessage.warning('数据格式异常')
    }
  } catch (err: any) {
    console.error('API调用失败:', err)
    error.value = err.message || '未知错误'
    ElMessage.error('加载失败: ' + error.value)
  } finally {
    loading.value = false
  }
}
</script>