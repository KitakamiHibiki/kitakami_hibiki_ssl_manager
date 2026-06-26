<template>
  <div class="deploy">
    <h2>部署管理</h2>
    <p>选择证书部署到目标服务器。</p>

    <el-form label-width="100px" style="max-width: 500px">
      <el-form-item label="选择证书">
        <el-select v-model="certId" placeholder="请选择证书">
          <el-option v-for="c in issuedCerts" :key="c.id" :label="c.domain" :value="c.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="部署目标">
        <el-select v-model="target">
          <el-option label="Nginx" value="nginx" />
          <el-option label="本地目录" value="local" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="deploy">部署</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCertificates, deployCertificate } from '../api'

const certs = ref<any[]>([])
const certId = ref<number | null>(null)
const target = ref('nginx')

const issuedCerts = computed(() => certs.value.filter((c: any) => c.status === 'issued'))

async function load() {
  const { data } = await getCertificates()
  certs.value = data
}

async function deploy() {
  if (!certId.value) {
    ElMessage.warning('请选择证书')
    return
  }
  await deployCertificate(certId.value, target.value)
  ElMessage.success('部署成功')
}

onMounted(load)
</script>

<style scoped>
h2 { margin-bottom: 10px; }
p { margin-bottom: 20px; color: #666; }
</style>
