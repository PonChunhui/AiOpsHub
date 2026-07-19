<template>
  <div class="user-manage">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon><User /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-value">{{ users.length }}</div>
              <div class="stats-label">总用户数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon><Avatar /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-value">{{ adminCount }}</div>
              <div class="stats-label">管理员</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-value">{{ userCount }}</div>
              <div class="stats-label">普通用户</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 用户列表卡片 -->
    <el-card class="user-list-card">
      <template #header>
        <div class="card-header">
          <div class="card-header-left">
            <el-icon><User /></el-icon>
            <span>用户列表</span>
          </div>
          <div class="card-header-right">
            <el-input
              v-model="searchQuery"
              placeholder="搜索用户名或邮箱"
              clearable
              style="width: 250px; margin-right: 12px;"
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select v-model="roleFilter" placeholder="角色筛选" clearable style="width: 150px;" @change="handleSearch">
              <el-option label="全部角色" value="" />
              <el-option label="管理员" value="admin" />
              <el-option label="普通用户" value="user" />
            </el-select>
          </div>
        </div>
      </template>

      <el-table 
        v-loading="loading" 
        :data="filteredUsers" 
        stripe
        border
        style="width: 100%"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }"
      >
        <el-table-column prop="id" label="ID" width="180" show-overflow-tooltip />
        <el-table-column prop="username" label="用户名" width="160">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar :size="32" style="background: #409eff; margin-right: 8px;">
                {{ row.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <span class="username-text">{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" min-width="220" show-overflow-tooltip />
        <el-table-column prop="role" label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="getRoleColor(row.role)" effect="light" size="default">
              <el-icon style="margin-right: 4px;"><component :is="getRoleIcon(row.role)" /></el-icon>
              {{ getRoleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            <div class="time-info">
              <el-icon style="margin-right: 4px; color: #909399;"><Clock /></el-icon>
              {{ formatDate(row.created_at) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="editUser(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="danger" link @click="deleteUser(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
    
    <!-- 编辑用户对话框 -->
    <el-dialog v-model="dialogVisible" title="编辑用户" width="500px">
      <el-form :model="userForm" label-width="100px">
        <el-form-item label="用户名">
          <el-input v-model="userForm.username" disabled />
        </el-form-item>
        
        <el-form-item label="邮箱">
          <el-input v-model="userForm.email" />
        </el-form-item>
        
        <el-form-item label="角色">
          <el-select v-model="userForm.role" style="width: 100%">
            <el-option label="管理员" value="admin" />
            <el-option label="用户" value="user" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitUser">更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { userApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, Avatar, UserFilled, Search, Clock, Edit, Delete } from '@element-plus/icons-vue'

const users = ref<any[]>([])
const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)

// 搜索和筛选
const searchQuery = ref('')
const roleFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)

const editingId = ref('')
const userForm = reactive({
  username: '',
  email: '',
  role: 'user'
})

// 计算属性
const adminCount = computed(() => users.value.filter(u => u.role === 'admin').length)
const userCount = computed(() => users.value.filter(u => u.role === 'user').length)

const filteredUsers = computed(() => {
  let result = users.value
  
  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(u => 
      u.username?.toLowerCase().includes(query) || 
      u.email?.toLowerCase().includes(query)
    )
  }
  
  // 角色过滤
  if (roleFilter.value) {
    result = result.filter(u => u.role === roleFilter.value)
  }
  
  return result
})

const total = computed(() => filteredUsers.value.length)

// 角色相关
const getRoleColor = (role: string) => {
  switch (role) {
    case 'admin':
      return 'danger'
    case 'user':
      return 'primary'
    default:
      return 'info'
  }
}

const getRoleIcon = (role: string) => {
  switch (role) {
    case 'admin':
      return Avatar
    case 'user':
      return UserFilled
    default:
      return User
  }
}

const getRoleLabel = (role: string) => {
  switch (role) {
    case 'admin':
      return '管理员'
    case 'user':
      return '普通用户'
    default:
      return role
  }
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1
}

// 分页处理
const handleSizeChange = () => {
  currentPage.value = 1
}

const handleCurrentChange = () => {
  // 分页变化时重新加载（如果需要后端分页）
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await userApi.list()
    if (res && res.code === 200) {
      users.value = res.data || []
    }
  } catch (error: any) {
    ElMessage.error('加载失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const editUser = (user: any) => {
  editingId.value = user.id
  userForm.username = user.username
  userForm.email = user.email
  userForm.role = user.role
  dialogVisible.value = true
}

const submitUser = async () => {
  submitting.value = true
  try {
    const res = await userApi.update(editingId.value, {
      email: userForm.email,
      role: userForm.role
    })
    
    if (res && res.code === 200) {
      ElMessage.success('更新成功')
      dialogVisible.value = false
      loadUsers()
    } else {
      ElMessage.error(res.message || '更新失败')
    }
  } catch (error: any) {
    ElMessage.error('更新失败: ' + error.message)
  } finally {
    submitting.value = false
  }
}

const deleteUser = async (user: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await userApi.delete(user.id)
    if (res && res.code === 200) {
      ElMessage.success('删除成功')
      loadUsers()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-manage {
  padding: 0;
}

/* 统计卡片 */
.stats-row {
  margin-bottom: 20px;
}

.stats-card {
  transition: all 0.3s ease;
}

.stats-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.stats-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stats-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 28px;
  flex-shrink: 0;
}

.stats-info {
  flex: 1;
}

.stats-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
  margin-bottom: 4px;
}

.stats-label {
  font-size: 13px;
  color: #909399;
  font-weight: 500;
}

/* 用户列表卡片 */
.user-list-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.card-header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 表格样式 */
.user-info {
  display: flex;
  align-items: center;
}

.username-text {
  font-weight: 500;
  color: #303133;
}

.time-info {
  display: flex;
  align-items: center;
  color: #606266;
  font-size: 13px;
}

/* 表格操作按钮 */
:deep(.el-button) {
  font-weight: 500;
}

:deep(.el-button--primary.is-link) {
  color: #409eff;
}

:deep(.el-button--primary.is-link:hover) {
  color: #66b1ff;
}

:deep(.el-button--danger.is-link) {
  color: #f56c6c;
}

:deep(.el-button--danger.is-link:hover) {
  color: #f78989;
}

/* 表格样式优化 */
:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-table th.el-table__cell) {
  font-size: 13px;
  text-transform: none;
  letter-spacing: normal;
}

:deep(.el-table .el-table__row:hover > td) {
  background: #f5f7fa !important;
}

/* 分页样式 */
:deep(.el-pagination) {
  padding: 16px 0 0 0;
}

:deep(.el-pagination .el-pager li) {
  font-weight: 500;
}

/* 对话框样式 */
:deep(.el-dialog__header) {
  border-bottom: 1px solid #e4e7ed;
  margin-right: 0;
  padding: 16px 20px;
}

:deep(.el-dialog__body) {
  padding: 20px;
}

:deep(.el-dialog__footer) {
  border-top: 1px solid #e4e7ed;
  padding: 12px 20px;
}

/* 响应式 */
@media (max-width: 768px) {
  .stats-row {
    margin-bottom: 16px;
  }
  
  .stats-card {
    margin-bottom: 12px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .card-header-right {
    width: 100%;
    flex-direction: column;
  }
  
  .card-header-right .el-input,
  .card-header-right .el-select {
    width: 100% !important;
  }
}
</style>