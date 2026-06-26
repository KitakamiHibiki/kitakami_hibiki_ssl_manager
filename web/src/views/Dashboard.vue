<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>域名总数</template>
          <div class="stat">{{ loading ? '-' : domains.length }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>已签发证书</template>
          <div class="stat">{{ loading ? '-' : issuedCount }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>即将过期</template>
          <div class="stat warn">{{ loading ? '-' : expiringCount }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>当前平台</template>
          <div class="stat">{{ platform || '未知' }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getDomains, getCertificates, getPlatform } from '../api'

const domains = ref<any[]>([])
const certificates = ref<any[]>([])
const platform = ref('')
const loading = ref(true)

const issuedCount = computed(() => certificates.value.filter((c: any) => c.status === 'issued').length)
const expiringCount = computed(() => {
  const now = new Date()
  const threshold = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000)
  return certificates.value.filter((c: any) => {
    if (c.status !== 'issued') return false
    const exp = new Date(c.expires_at)
    return exp < threshold
  }).length
})

onMounted(async () => {
  try {
    const [d, c] = await Promise.all([getDomains(), getCertificates()])
    domains.value = d.data
    certificates.value = c.data
  } catch (e) {
    console.error(e)
  }
  try {
    const p = await getPlatform()
    platform.value = `${p.data.os}/${p.data.arch}`
  } catch (e) {
    console.error('platform fetch failed:', e)
  }
  loading.value = false
})
</script>

<style scoped>
.dashboard h2 { margin-bottom: 20px; }
.stat { font-size: 32px; font-weight: bold; text-align: center; padding: 10px 0; }
.stat.warn { color: #e6a23c; }
</style>
