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
        
        <el-menu-item index="/agents-manage" class="sidebar-menu-item">
          <el-icon><Monitor /></el-icon>
          <span>Agent管理</span>
        </el-menu-item>
        
        <el-menu-item index="/alerts-manage" class="sidebar-menu-item">
          <el-icon><Bell /></el-icon>
          <span>告警管理</span>
        </el-menu-item>
        
        <el-menu-item index="/mcp-manage" class="sidebar-menu-item">
          <el-icon><Tools /></el-icon>
          <span>MCP管理</span>
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
  Monitor,
  Bell,
  User,
  FolderOpened,
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
/* ===== 布局容器样式 ===== */
.layout-container {
  height: 100vh;
}

/* ===== 侧边栏容器样式 ===== */
.sidebar-container {
  background: linear-gradient(180deg, #1f2937 0%, #111827 100%);
  height: 100vh;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  /* 消除Element Plus默认的padding和margin */
  padding: 0 !important;
  margin: 0 !important;
}

/* ===== Logo区域样式 ===== */
.sidebar-logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  background: rgba(59, 130, 246, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  letter-spacing: 1px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  /* 消除可能的间距 */
  padding: 0;
  margin: 0;
}

/* ===== 侧边栏菜单样式 ===== */
.sidebar-menu {
  flex: 1;
  border-right: none;
  background: transparent;
  padding: 12px 0 !important; /* 上下保留间距，左右无间距 */
  /* 消除Element Plus默认的左侧padding */
  padding-left: 0 !important;
}

/* ===== 菜单项样式 ===== */
.sidebar-menu-item {
  height: 50px;
  line-height: 50px;
  margin: 4px 12px; /* 保留菜单项左右间距，避免贴边 */
  border-radius: 8px;
  color: #d1d5db;
  transition: all 0.3s ease;
  /* 消除Element Plus默认padding */
  padding: 0 20px !important;
}

/* 菜单项悬停效果 */
.sidebar-menu-item:hover {
  background: rgba(59, 130, 246, 0.15);
  color: #fff;
}

/* 菜单项选中效果 */
.sidebar-menu-item.is-active {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: #fff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

/* 菜单项图标样式 */
.sidebar-menu-item .el-icon {
  font-size: 18px;
  margin-right: 8px;
}

/* ===== 页面容器样式 ===== */
.page-container {
  overflow: hidden !important;
  height: calc(100vh - 60px) !important;
  padding: 20px !important;
  background: #f9fafb;
}

/* Element Plus 主容器样式覆盖 */
.el-main {
  overflow: hidden !important;
}

/* Element Plus Aside组件样式覆盖 - 消除默认间距 */
:deep(.el-aside) {
  padding: 0 !important;
  margin: 0 !important;
}
</style>

<style>
/* ===== 全局样式覆盖 ===== */
/* 禁用 Vue Devtools 浮动图标 */
#__vue-devtools-container__,
.vue-devtools {
  display: none !important;
  visibility: hidden !important;
}

/* ===== Body样式强制重置 ===== */
/* 消除浏览器默认的8px margin */
body {
  margin: 0 !important;
  padding: 0 !important;
}

/* ===== Element Plus 全局样式覆盖 ===== */
/* 消除侧边栏默认间距 */
.el-aside {
  padding: 0 !important;
  margin: 0 !important;
}

/* 消除菜单默认padding */
.el-menu {
  padding-left: 0 !important;
  border-right: none !important;
}

/* 消除菜单项默认padding */
.el-menu-item {
  padding: 0 20px !important;
}
</style>