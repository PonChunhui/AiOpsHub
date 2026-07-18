<template>
  <div class="login-container">
    <div class="login-bg-layer">
      <div class="login-bg-gradient"></div>
      <div class="login-bg-grid"></div>
      <div class="login-bg-glow"></div>
      <div class="login-bg-particles">
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
        <div class="login-particle"></div>
      </div>
      <div class="login-bg-lines">
        <div class="login-line"></div>
        <div class="login-line"></div>
        <div class="login-line"></div>
        <div class="login-line"></div>
        <div class="login-line"></div>
      </div>
    </div>
    
    <el-card class="login-card">
      <div class="login-header">
        <div class="login-logo">
          <el-icon class="login-logo-icon"><Cloudy /></el-icon>
        </div>
        <h2 class="login-title">AiOpsHub</h2>
        <p class="login-subtitle">智能运维平台</p>
      </div>
      
      <div class="login-body">
        <el-form @submit.prevent="handleLogin">
          <el-form-item class="login-form-item">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              size="large"
              class="login-input"
              prefix-icon="User"
            />
          </el-form-item>
          
          <el-form-item class="login-form-item">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              size="large"
              class="login-input"
              prefix-icon="Lock"
              show-password
            />
          </el-form-item>
          
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              class="login-button"
              @click="handleLogin"
              :loading="loading"
            >
              登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import { Cloudy } from '@element-plus/icons-vue'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

const handleLogin = async () => {
  if (!form.username || !form.password) {
    ElMessage.warning('请输入用户名和密码')
    return
  }
  
  loading.value = true
  
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error: any) {
    const message = error.response?.data?.message || error.message || '登录失败'
    if (!message.includes('401')) {
      ElMessage.error(message)
    } else {
      ElMessage.error('用户名或密码错误')
    }
  } finally {
    loading.value = false
  }
}
</script>