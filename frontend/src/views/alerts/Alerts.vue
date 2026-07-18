<template>
  <el-card>
    <template #header>
      <span>告警管理</span>
    </template>
    
    <el-table :data="alerts">
      <el-table-column prop="id" label="ID" width="250" />
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="severity" label="严重性">
        <template #default="{ row }">
          <el-tag :type="row.severity === 'high' ? 'danger' : 'warning'">
            {{ row.severity }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="source" label="来源" />
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { alertApi } from '@/api'

const alerts = ref<any[]>([])

onMounted(async () => {
  try {
    const res = await alertApi.list()
    alerts.value = res.data || []
  } catch (error) {
    console.error('Failed to load alerts:', error)
  }
})
</script>