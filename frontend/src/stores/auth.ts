import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import { ElMessage } from 'element-plus'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const username = ref<string>(localStorage.getItem('username') || '')
  const email = ref<string>(localStorage.getItem('email') || '')
  const userId = ref<string>(localStorage.getItem('userId') || '')
  
  const isAuthenticated = computed(() => !!token.value)
  
  const login = async (user: string, pass: string) => {
    try {
      const res = await authApi.login(user, pass)
      
      if (res && res.code === 200) {
        token.value = res.data.token || 'mock-token'
        username.value = res.data.username
        userId.value = res.data.user_id
        
        localStorage.setItem('token', token.value)
        localStorage.setItem('username', username.value)
        localStorage.setItem('userId', userId.value)
        
        ElMessage.success('登录成功')
        return true
      } else {
        ElMessage.error(res.message || '登录失败')
        return false
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      ElMessage.error(error.message || '登录失败')
      return false
    }
  }
  
  const logout = async () => {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout API failed:', error)
    }
    
    token.value = ''
    username.value = ''
    userId.value = ''
    
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('userId')
    localStorage.removeItem('email')
    
    ElMessage.success('已退出登录')
  }
  
  const checkAuth = () => {
    if (!token.value) {
      return false
    }
    return true
  }
  
  return {
    token,
    username,
    email,
    userId,
    isAuthenticated,
    login,
    logout,
    checkAuth
  }
})