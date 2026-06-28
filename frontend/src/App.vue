<template>
  <el-container class="layout-container" v-if="route.path !== '/login' && route.path !== '/register'">
    <el-aside width="200px" class="sidebar-container">
      <div class="sidebar-logo">AiOpsHub</div>
      
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/" class="sidebar-menu-item">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        
        <el-menu-item index="/ai-assistant" class="sidebar-menu-item">
          <el-icon><ChatDotRound /></el-icon>
          <span>AI助手</span>
        </el-menu-item>
        
        <el-menu-item index="/test" class="sidebar-menu-item">
          <el-icon><DocumentChecked /></el-icon>
          <span>API测试</span>
        </el-menu-item>
        
        <el-menu-item index="/agents-manage" class="sidebar-menu-item">
          <el-icon><Monitor /></el-icon>
          <span>Agent管理</span>
        </el-menu-item>
        
        <el-menu-item index="/alerts-manage" class="sidebar-menu-item">
          <el-icon><Bell /></el-icon>
          <span>告警管理</span>
        </el-menu-item>
        
        <el-menu-item index="/system-monitor-real" class="sidebar-menu-item">
          <el-icon><DataLine /></el-icon>
          <span>系统监控</span>
        </el-menu-item>
        
        <el-menu-item index="/monitor-dashboard" class="sidebar-menu-item">
          <el-icon><TrendCharts /></el-icon>
          <span>监控仪表板</span>
        </el-menu-item>
        
        <el-menu-item index="/log-query" class="sidebar-menu-item">
          <el-icon><Document /></el-icon>
          <span>日志查询</span>
        </el-menu-item>
        
        <el-menu-item index="/remediation" class="sidebar-menu-item">
          <el-icon><Tools /></el-icon>
          <span>自动修复</span>
        </el-menu-item>
        
        <el-menu-item index="/kubernetes" class="sidebar-menu-item">
          <el-icon><Cloudy /></el-icon>
          <span>Kubernetes</span>
        </el-menu-item>
        
        <el-menu-item index="/knowledge-base" class="sidebar-menu-item">
          <el-icon><FolderOpened /></el-icon>
          <span>知识库</span>
        </el-menu-item>
        
        <el-menu-item index="/users-manage" class="sidebar-menu-item">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header-container">
        <div class="header-title">
          <el-icon class="header-title-icon"><Cloudy /></el-icon>
          智能运维平台
        </div>
        
        <div class="header-actions">
          <div class="header-user">
            <div class="header-user-avatar">
              <el-icon><User /></el-icon>
            </div>
            <span class="header-user-name">{{ username }}</span>
          </div>
          
          <el-button @click="handleLogout" size="small">
            <el-icon><SwitchButton /></el-icon>
            退出
          </el-button>
        </div>
      </el-header>
      
      <el-main class="page-container">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
  
  <router-view v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { 
  HomeFilled,
  DocumentChecked,
  Monitor,
  Bell,
  Cpu,
  User,
  DataLine,
  FolderOpened,
  TrendCharts,
  Document,
  Tools,
  Cloudy,
  SwitchButton,
  ChatDotRound
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => route.path)
const username = computed(() => authStore.username || '未登录')

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
</style>