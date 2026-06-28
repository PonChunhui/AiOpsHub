<template>
  <el-card class="page-card">
    <template #header>
      <div class="page-card-header">
        <span class="page-card-title">API测试页面</span>
      </div>
    </template>
    
    <div class="button-group">
      <el-button type="primary" @click="testAgentsAPI">
        <el-icon><Connection /></el-icon>
        测试Agents API
      </el-button>
    </div>
    
    <div class="divider"></div>
    
    <el-alert v-if="testResult" :type="testStatus" :closable="false">
      <pre class="code-block">{{ testResult }}</pre>
    </el-alert>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { agentApi } from '@/api'
import { ElMessage } from 'element-plus'

const testResult = ref('')
const testStatus = ref<'success' | 'error' | 'info'>('info')

const testAgentsAPI = async () => {
  try {
    ElMessage.info('正在测试Agents API...')
    const res = await agentApi.list()
    testResult.value = JSON.stringify(res, null, 2)
    testStatus.value = 'success'
    ElMessage.success('Agents API调用成功！')
  } catch (error: any) {
    testResult.value = '错误: ' + error.message + '\n' + JSON.stringify(error.response?.data, null, 2)
    testStatus.value = 'error'
    ElMessage.error('Agents API调用失败')
  }
}
</script>