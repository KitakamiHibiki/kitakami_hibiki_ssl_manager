<template>
  <div class="domains">
    <div class="header">
      <h2>域名管理</h2>
      <el-button type="primary" @click="$router.push('/domains/add')">添加域名</el-button>
    </div>

    <el-table :data="domains" style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column label="域名">
        <template #default="{ row }">
          <span class="domain-cell">
            {{ row.domain }}
            <el-icon class="copy-icon" @click="copyDomain(row.domain)"><CopyDocument /></el-icon>
          </span>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="200">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="$router.push('/domains/detail?id=' + row.id)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { ElMessage } from "element-plus"
import { CopyDocument } from "@element-plus/icons-vue"
import { getDomains } from "../api"

const domains = ref<any[]>([])

function copyDomain(name: string) {
  navigator.clipboard.writeText(name).then(() => ElMessage.success("已复制域名"))
}

function formatTime(t: string) {
  if (!t) return ""
  const d = new Date(t)
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

  async function load() {
    const { data } = await getDomains()
    domains.value = data
  }

  onMounted(load)
</script>

<style scoped>
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; }
.domain-cell { display: inline-flex; align-items: center; gap: 6px; }
.copy-icon { cursor: pointer; color: #999; font-size: 14px; }
.copy-icon:hover { color: #409eff; }
</style>
