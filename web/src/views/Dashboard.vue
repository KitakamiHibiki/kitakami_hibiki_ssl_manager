<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>证书总数</template>
          <div class="stat">{{ loading ? '-' : certCount }}</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>当前平台</template>
          <div class="stat">{{ platform || '未知' }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getCertificates, getPlatform } from '../api'

const certCount = ref(0)
const platform = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    const d = await getCertificates({ page_size: 1 })
    certCount.value = d.data.total || 0
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
