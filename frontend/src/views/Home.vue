<template>
  <el-card class="page-card">
    <template #header>
      <div class="page-card-header">
        <h2 class="page-card-title">系统概览</h2>
      </div>
    </template>
    
    <el-row :gutter="24">
      <el-col :span="12">
        <div class="stats-card">
          <div class="stats-value">{{ stats.agents }}</div>
          <div class="stats-label">Agent数量</div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="stats-card">
          <div class="stats-value">{{ stats.alerts }}</div>
          <div class="stats-label">告警数量</div>
        </div>
      </el-col>
    </el-row>
  </el-card>
  
  <el-card class="page-card">
    <template #header>
      <div class="page-card-header">
        <h3 class="page-card-title">快速操作</h3>
      </div>
    </template>
    
    <div class="button-group">
      <el-button type="primary" size="large" @click="$router.push('/agents-manage')">
        <el-icon><Monitor /></el-icon>
        Agent管理
      </el-button>
      <el-button type="success" size="large" @click="$router.push('/alerts-manage')">
        <el-icon><Bell /></el-icon>
        查看告警
      </el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { agentApi, alertApi } from '@/api'

const stats = ref({
  agents: 0,
  alerts: 0
})

onMounted(async () => {
  try {
    const [agents, alerts] = await Promise.all([
      agentApi.list(),
      alertApi.list()
    ])
    
    stats.value.agents = agents.data?.length || 0
    stats.value.alerts = alerts.data?.length || 0
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
})
</script>