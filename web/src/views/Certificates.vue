<template>
  <div class="certs">
    <h2>证书管理</h2>

    <el-table :data="certs" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="domain" label="域名" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="issued_at" label="签发时间" width="180" />
      <el-table-column prop="expires_at" label="到期时间" width="180" />
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" type="warning" @click="renew(row.id)">续签</el-button>
          <el-button size="small" type="primary" @click="download(row.id)">下载</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCertificates, renewCertificate } from '../api'

const certs = ref<any[]>([])

function statusType(s: string) {
  return s === 'issued' ? 'success' : s === 'pending' ? 'warning' : s === 'error' ? 'danger' : 'info'
}

function statusLabel(s: string) {
  const map: Record<string, string> = { issued: '已签发', pending: '申请中', error: '失败', expired: '已过期' }
  return map[s] || s
}

async function load() {
  const { data } = await getCertificates()
  certs.value = data
}

async function renew(id: number) {
  await renewCertificate(id)
  ElMessage.success('续签已提交')
  setTimeout(load, 2000)
}

function download(id: number) {
  window.open(`/api/certs/${id}/download`)
}

onMounted(load)
</script>

<style scoped>
h2 { margin-bottom: 20px; }
</style>
