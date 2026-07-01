<template>
  <div class="users">
    <h2>用户管理</h2>
    <el-table :data="users" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column label="角色" width="120">
        <template #default="{ row }">
          <el-select
            :model-value="row.role"
            size="small"
            @change="(val: string) => changeRole(row.id, val)"
            :disabled="row.id === currentUserId"
          >
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button
            size="small" type="danger"
            @click="remove(row.id)"
            :disabled="row.id === currentUserId"
          >删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUsers, updateUser, deleteUser } from '../api'

interface User {
  id: number
  username: string
  role: string
  created_at: string
}

const users = ref<User[]>([])
const currentUserId = ref(0)

async function load() {
  const { data } = await getUsers()
  users.value = data
}

async function changeRole(id: number, role: string) {
  await updateUser(id, role)
  ElMessage.success('角色已更新')
  load()
}

async function remove(id: number) {
  await ElMessageBox.confirm('确定删除该用户？', '确认', { type: 'warning' })
  await deleteUser(id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<style scoped>
h2 { margin-bottom: 20px; }
</style>
