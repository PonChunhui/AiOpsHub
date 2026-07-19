<template>
  <!-- 全屏页面：无侧边栏和头部（终端、文件管理） -->
  <router-view v-if="isFullScreenPage" />
  
  <!-- 其他页面：带侧边栏和头部 -->
  <el-container class="layout-container" v-else-if="route.path !== '/login' && route.path !== '/register'">
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
        
        <el-menu-item index="/agents-manage" class="sidebar-menu-item">
          <el-icon><Monitor /></el-icon>
          <span>Agent管理</span>
        </el-menu-item>
        
        <el-menu-item index="/host-manage" class="sidebar-menu-item">
          <el-icon><Platform /></el-icon>
          <span>主机管理</span>
        </el-menu-item>
        
        <el-menu-item index="/alerts-manage" class="sidebar-menu-item">
          <el-icon><Bell /></el-icon>
          <span>告警管理</span>
        </el-menu-item>
        
        <el-menu-item index="/mcp-manage" class="sidebar-menu-item">
          <el-icon><Tools /></el-icon>
          <span>MCP管理</span>
        </el-menu-item>
        
        <el-menu-item index="/tools-manage" class="sidebar-menu-item">
          <el-icon><Setting /></el-icon>
          <span>工具管理</span>
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
      <el-header v-if="!noHeaderPage" class="header-container">
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
      
      <el-main class="page-container" :style="noHeaderPage ? 'height: 100vh !important' : ''">
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
  Monitor,
  Bell,
  User,
  FolderOpened,
  Tools,
  Setting,
  Cloudy,
  SwitchButton,
  ChatDotRound,
  Platform
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isFullScreenPage = computed(() => 
  route.path.startsWith('/host-terminal/') || 
  route.path.startsWith('/host-file-manage')
)
const noHeaderPage = computed(() => 
  route.path === '/ai-assistant'
)
const activeMenu = computed(() => route.path)
const username = computed(() => authStore.username || '未登录')

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
/* ===== Layout-only styles (visual styles in sidebar.css / header.css / common.css) ===== */

.layout-container {
  height: 100vh;
}

/* Sidebar layout overrides — eliminate Element Plus defaults */
.sidebar-container {
  padding: 0 !important;
  margin: 0 !important;
}

/* Main content area layout */
.page-container {
  overflow-y: auto !important;
  overflow-x: hidden !important;
  height: calc(100vh - var(--header-height)) !important;
}

.el-main {
  overflow-y: auto !important;
  overflow-x: hidden !important;
}

:deep(.el-aside) {
  padding: 0 !important;
  margin: 0 !important;
}
</style>

<style>
/* ===== Global Overrides ===== */

/* Disable Vue Devtools floating icon */
#__vue-devtools-container__,
.vue-devtools {
  display: none !important;
  visibility: hidden !important;
}

/* Element Plus layout resets — eliminate default spacing */
.el-aside {
  padding: 0 !important;
  margin: 0 !important;
}

.el-menu {
  padding-left: 0 !important;
  border-right: none !important;
}

.el-menu-item {
  padding: 0 20px !important;
}
</style>