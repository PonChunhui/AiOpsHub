<template>
  <div class="home-container">
    <el-row :gutter="24" class="stats-row">
      <el-col :span="8">
        <div class="stats-card agents-card">
          <div class="stats-icon">
            <el-icon><Monitor /></el-icon>
          </div>
          <div class="stats-content">
            <div class="stats-value">{{ stats.agents }}</div>
            <div class="stats-label">Agent数量</div>
          </div>
          <div class="stats-decoration"></div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stats-card alerts-card">
          <div class="stats-icon">
            <el-icon><Bell /></el-icon>
          </div>
          <div class="stats-content">
            <div class="stats-value">{{ stats.alerts }}</div>
            <div class="stats-label">告警数量</div>
          </div>
          <div class="stats-decoration"></div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stats-card tokens-card">
          <div class="stats-icon">
            <el-icon><Coin /></el-icon>
          </div>
          <div class="stats-content">
            <div class="stats-value">{{ formatNumber(stats.totalTokens) }}</div>
            <div class="stats-label">Token消耗</div>
          </div>
          <div class="stats-decoration"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="detail-row">
      <el-col :span="12">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <el-icon class="card-icon"><Calendar /></el-icon>
              <h3>今日Token统计</h3>
            </div>
          </template>
          <div class="token-stats">
            <div class="token-item">
              <div class="token-label">输入Token</div>
              <div class="token-value">{{ formatNumber(stats.todayInputTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">输出Token</div>
              <div class="token-value">{{ formatNumber(stats.todayOutputTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">总消耗</div>
              <div class="token-value primary">{{ formatNumber(stats.todayTotalTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">估算费用</div>
              <div class="token-value success">¥{{ stats.todayCost.toFixed(2) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card class="detail-card">
          <template #header>
            <div class="card-header">
              <el-icon class="card-icon"><DataAnalysis /></el-icon>
              <h3>本月Token统计</h3>
            </div>
          </template>
          <div class="token-stats">
            <div class="token-item">
              <div class="token-label">输入Token</div>
              <div class="token-value">{{ formatNumber(stats.monthInputTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">输出Token</div>
              <div class="token-value">{{ formatNumber(stats.monthOutputTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">总消耗</div>
              <div class="token-value primary">{{ formatNumber(stats.monthTotalTokens) }}</div>
            </div>
            <div class="token-item">
              <div class="token-label">估算费用</div>
              <div class="token-value success">¥{{ stats.monthCost.toFixed(2) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="action-card">
      <template #header>
        <div class="card-header">
          <el-icon class="card-icon"><Operation /></el-icon>
          <h3>快速操作</h3>
        </div>
      </template>
      <div class="button-group">
        <div class="action-button agents-action" @click="$router.push('/agents-manage')">
          <div class="action-icon">
            <el-icon><Monitor /></el-icon>
          </div>
          <div class="action-content">
            <div class="action-title">Agent管理</div>
            <div class="action-desc">管理和配置智能Agent</div>
          </div>
        </div>
        <div class="action-button alerts-action" @click="$router.push('/alerts-manage')">
          <div class="action-icon">
            <el-icon><Bell /></el-icon>
          </div>
          <div class="action-content">
            <div class="action-title">查看告警</div>
            <div class="action-desc">处理系统告警信息</div>
          </div>
        </div>
        <div class="action-button assistant-action" @click="$router.push('/ai-assistant')">
          <div class="action-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
          <div class="action-content">
            <div class="action-title">AI助手</div>
            <div class="action-desc">智能运维对话助手</div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Monitor, Bell, ChatDotRound, Coin, Calendar, DataAnalysis, Operation } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '@/api'

const stats = ref({
  agents: 0,
  alerts: 0,
  totalTokens: 0,
  todayInputTokens: 0,
  todayOutputTokens: 0,
  todayTotalTokens: 0,
  todayCost: 0,
  monthInputTokens: 0,
  monthOutputTokens: 0,
  monthTotalTokens: 0,
  monthCost: 0
})

const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

onMounted(async () => {
  try {
    const [agentsRes, alertsRes, tokenRes] = await Promise.all([
      api.get('/agents'),
      api.get('/alerts'),
      api.get('/tokens/stats')
    ])
    
    console.log('API responses:', {
      agents: agentsRes,
      alerts: alertsRes,
      tokens: tokenRes
    })
    
    // agentsRes 已经是 response.data，不需要再访问 .data
    if (agentsRes?.code === 200) {
      stats.value.agents = agentsRes?.data?.agents?.length || agentsRes?.data?.total || 0
    }
    
    if (alertsRes?.code === 200) {
      stats.value.alerts = alertsRes?.data?.length || alertsRes?.data?.alerts?.length || 0
    }
    
    // tokenRes 直接就是 data，不需要检查 code
    if (tokenRes) {
      const tokenData = tokenRes.data || tokenRes
      console.log('Token data:', tokenData)
      stats.value.totalTokens = tokenData.total_tokens || 0
      stats.value.todayInputTokens = tokenData.today_input_tokens || 0
      stats.value.todayOutputTokens = tokenData.today_output_tokens || 0
      stats.value.todayTotalTokens = tokenData.today_total_tokens || 0
      stats.value.todayCost = tokenData.today_cost || 0
      stats.value.monthInputTokens = tokenData.month_input_tokens || 0
      stats.value.monthOutputTokens = tokenData.month_output_tokens || 0
      stats.value.monthTotalTokens = tokenData.month_total_tokens || 0
      stats.value.monthCost = tokenData.month_cost || 0
      console.log('Stats updated:', stats.value)
    }
} catch (error: any) {
  console.error('Failed to load stats:', error)
  
  // 401: 登录过期（API拦截器已提示）
  // 502/503/504: 服务器错误（API拦截器已提示）
  // 不显示重复提示
  if (error.response?.status !== 401 &&
      error.response?.status !== 502 &&
      error.response?.status !== 503 &&
      error.response?.status !== 504) {
    ElMessage.error('加载统计数据失败，请刷新页面重试')
  }
}
})
</script>

<style scoped>
.home-container {
  padding: 0;
  min-height: calc(100vh - 60px);
}

.stats-row {
  margin-bottom: 24px;
}

.stats-card {
  position: relative;
  padding: 32px 24px;
  border-radius: 16px;
  color: white;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 20px;
}

.stats-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.agents-card {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.alerts-card {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.tokens-card {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stats-icon {
  width: 64px;
  height: 64px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  backdrop-filter: blur(10px);
}

.stats-content {
  flex: 1;
  position: relative;
  z-index: 1;
}

.stats-value {
  font-size: 48px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 8px;
}

.stats-label {
  font-size: 16px;
  opacity: 0.9;
  font-weight: 500;
}

.stats-decoration {
  position: absolute;
  top: -50%;
  right: -20%;
  width: 200px;
  height: 200px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  z-index: 0;
}

.detail-row {
  margin-bottom: 24px;
}

.detail-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: none;
}

.detail-card :deep(.el-card__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-card :deep(.el-card__body) {
  padding: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.card-icon {
  font-size: 24px;
  color: #409eff;
}

.token-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.token-item {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.token-item:hover {
  background: #ecf5ff;
}

.token-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.token-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.token-value.primary {
  color: #409eff;
}

.token-value.success {
  color: #67c23a;
}

.action-card {
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: none;
}

.action-card :deep(.el-card__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.action-card :deep(.el-card__body) {
  padding: 24px;
}

.button-group {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.action-button {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  border-color: transparent;
}

.agents-action:hover {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}

.alerts-action:hover {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.assistant-action:hover {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.action-icon {
  width: 56px;
  height: 56px;
  background: #f5f7fa;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #409eff;
  transition: all 0.3s ease;
}

.action-button:hover .action-icon {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.action-content {
  flex: 1;
}

.action-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 4px;
}

.action-desc {
  font-size: 14px;
  opacity: 0.7;
}

@media (max-width: 1200px) {
  .button-group {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-row :deep(.el-col),
  .detail-row :deep(.el-col) {
    margin-bottom: 16px;
  }
  
  .button-group {
    grid-template-columns: 1fr;
  }
  
  .stats-value {
    font-size: 36px;
  }
}
</style>