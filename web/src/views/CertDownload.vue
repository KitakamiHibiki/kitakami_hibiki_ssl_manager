<template>
  <div class="cert-download-page">
    <div class="header">
      <h2>证书下载</h2>
      <el-button text @click="$router.push(`/domains/detail?id=${domainId}`)">返回域名详情</el-button>
    </div>

    <div v-if="status === 'issuing'" class="section loading-section">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p style="margin-top:16px;color:#409eff;font-size:16px">证书签发中，请稍候...</p>
      <p style="color:#999;font-size:13px">正在通过 Let's Encrypt 签发证书，通常需要 10-30 秒</p>
    </div>

    <div v-else-if="status === 'issued' && cert" class="section">
      <h3 style="color:#67c23a;margin-bottom:16px">证书已签发</h3>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="域名">{{ domain }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag type="success">已签发</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="签发时间">{{ formatTime(cert.issued_at) }}</el-descriptions-item>
        <el-descriptions-item label="到期时间">{{ formatTime(cert.expires_at) }}</el-descriptions-item>
      </el-descriptions>

      <div style="margin-top:20px">
        <h4>下载证书文件</h4>
        <div style="display:flex;gap:12px;margin-top:12px">
          <el-button type="primary" @click="downloadFile('fullchain')" :loading="downloading === 'fullchain'">
            下载 fullchain.pem
          </el-button>
          <el-button type="primary" @click="downloadFile('privkey')" :loading="downloading === 'privkey'">
            下载 privkey.pem
          </el-button>
        </div>
        <p style="color:#e6a23c;font-size:13px;margin-top:12px">
          请妥善保管私钥文件，不要泄露给他人。
        </p>
      </div>
    </div>

    <div v-else-if="status === 'error'" class="section">
      <h3 style="color:#f56c6c;margin-bottom:16px">证书申请失败</h3>
      <p style="color:#f56c6c">{{ errorMsg }}</p>
      <el-button style="margin-top:16px" @click="$router.push(`/domains/detail?id=${domainId}`)">返回域名详情</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue"
import { useRoute } from "vue-router"
import { ElMessage } from "element-plus"
import { Loading } from "@element-plus/icons-vue"
import { getCertStatus, getCertificate } from "../api"
import api from "../api"

const route = useRoute()
const domainId = Number(route.query.domain_id) || 0
const certId = Number(route.query.cert_id) || 0

const status = ref("issuing")
const errorMsg = ref("")
const domain = ref("")
const cert = ref<any>(null)
const downloading = ref("")

let pollTimer: number | null = null

function formatTime(t: string) {
  if (!t) return "-"
  const d = new Date(t)
  if (isNaN(d.getTime())) return "-"
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

async function downloadFile(type: string) {
  downloading.value = type
  try {
    const resp = await api.get("/certs/download", {
      params: { id: certId, type },
      responseType: "blob",
    })
    const url = URL.createObjectURL(new Blob([resp.data]))
    const a = document.createElement("a")
    a.href = url
    a.download = `${domain.value}.${type}.pem`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success("下载成功")
  } catch (e: any) {
    ElMessage.error("下载失败")
  } finally {
    downloading.value = ""
  }
}

async function poll() {
  try {
    const { data } = await getCertStatus(domain.value)
    status.value = data.status
    if (data.status === "issued") {
      stopPoll()
      const { data: certData } = await getCertificate(certId)
      cert.value = certData
    } else if (data.status === "error") {
      errorMsg.value = data.error_msg
      stopPoll()
    }
  } catch {}
}

function stopPoll() {
  if (pollTimer !== null) { clearInterval(pollTimer); pollTimer = null }
}

onMounted(async () => {
  if (!domainId || !certId) {
    status.value = "error"
    errorMsg.value = "参数无效"
    return
  }
  try {
    const { data: certData } = await getCertificate(certId)
    domain.value = certData.domain
    status.value = certData.status
    if (certData.status === "issued") {
      cert.value = certData
    } else if (certData.status === "error") {
      errorMsg.value = certData.error_msg
    } else {
      pollTimer = window.setInterval(poll, 1000)
    }
  } catch {
    status.value = "error"
    errorMsg.value = "加载证书信息失败"
  }
})

onUnmounted(stopPoll)
</script>

<style scoped>
.cert-download-page { width: 100%; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.section {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 16px;
}
.loading-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
}
</style>
