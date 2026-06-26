<template>
  <div class="domains">
    <div class="header">
      <h2>域名管理</h2>
      <el-button type="primary" @click="showDialog = true">添加域名</el-button>
    </div>

    <el-table :data="domains" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="domain" label="域名" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="challenge" label="验证方式" width="100" />
      <el-table-column prop="created_at" label="创建时间" width="180" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" type="success" @click="applyCert(row)">申请证书</el-button>
          <el-button size="small" type="danger" @click="remove(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showDialog" title="添加域名" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="域名">
          <el-input v-model="form.domain" placeholder="example.com" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="admin@example.com" />
        </el-form-item>
        <el-form-item label="验证方式">
          <el-select v-model="form.challenge">
            <el-option label="HTTP" value="http" />
            <el-option label="DNS" value="dns" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getDomains, createDomain, deleteDomain, applyCertificate } from '../api'

const domains = ref<any[]>([])
const showDialog = ref(false)
const form = ref({ domain: '', email: '', challenge: 'http' })

async function load() {
  const { data } = await getDomains()
  domains.value = data
}

async function submit() {
  await createDomain(form.value)
  ElMessage.success('域名添加成功')
  showDialog.value = false
  form.value = { domain: '', email: '', challenge: 'http' }
  load()
}

async function remove(id: number) {
  await deleteDomain(id)
  ElMessage.success('删除成功')
  load()
}

async function applyCert(row: any) {
  try {
    await applyCertificate({ domain_id: row.id })
    ElMessage.success('证书申请已提交，请前往证书页面查看进度')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '申请失败')
  }
}

onMounted(load)
</script>

<style scoped>
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; }
</style>
