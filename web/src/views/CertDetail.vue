<template>
  <div class="detail-page">
    <div class="header">
      <h2>证书详情</h2>
      <el-button text @click="$router.push('/certs')">返回列表</el-button>
    </div>

    <div v-if="loading" style="color:#999">加载中...</div>
    <div v-else-if="error" style="color:#f56c6c">{{ error }}</div>
    <template v-else-if="cert">
      <div class="section">
        <h3>基本信息</h3>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="证书 ID">{{ cert.id }}</el-descriptions-item>
          <el-descriptions-item label="域名">
            <strong>{{ cert.domain }}</strong>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(cert.status)">{{ cert.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签发时间">
            {{ cert.status === 'issued' ? formatTime(cert.issued_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="到期时间">
            {{ cert.status === 'issued' ? formatTime(cert.expires_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(cert.created_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <div v-if="cert.status === 'issued'" class="section">
        <h3>操作</h3>
        <div style="display:flex;gap:12px;flex-wrap:wrap">
          <el-button type="primary" @click="downloadFile('fullchain','pem')" :loading="downloading === 'fullchain'">
            下载证书 (.pem)
          </el-button>
          <el-button type="primary" @click="downloadFile('privkey','key')" :loading="downloading === 'privkey'">
            下载私钥 (.key)
          </el-button>
          <el-button type="success" @click="doDeploy" :loading="deploying">
            部署证书
          </el-button>
          <el-popconfirm title="确定删除此证书？" @confirm="doDelete">
            <template #reference>
              <el-button type="danger">删除证书</el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>

      <div v-if="cert.error_msg" class="section">
        <h3 style="color:#f56c6c">错误信息</h3>
        <p style="color:#f56c6c">{{ cert.error_msg }}</p>
      </div>

      <div class="section">
        <h3>部署记录</h3>
        <el-table :data="deployLogs" size="small" style="width:100%">
          <el-table-column prop="id" label="ID" width="50" />
          <el-table-column prop="node_name" label="目标节点" width="120" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : row.status === 'pending' ? 'info' : 'danger'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="detail" label="详情" min-width="200">
            <template #default="{ row }">
              <div v-if="row.status === 'failed'" style="color:#f56c6c;font-size:13px">{{ row.error_msg }}</div>
              <div v-else-if="row.detail" style="font-size:13px;white-space:pre-line">{{ row.detail }}</div>
              <span v-else style="color:#999">-</span>
            </template>
          </el-table-column>
          <el-table-column label="开始时间" width="160">
            <template #default="{ row }">{{ formatTime(row.started_at) }}</template>
          </el-table-column>
          <el-table-column label="结束时间" width="160">
            <template #default="{ row }">{{ formatTime(row.finished_at) }}</template>
          </el-table-column>
        </el-table>
        <div style="margin-top:12px;text-align:right" v-if="deployLogTotal > 0">
          <el-pagination
            v-model:current-page="deployLogPage"
            :page-size="deployLogPageSize"
            :total="deployLogTotal"
            layout="total, prev, pager, next"
            @current-change="loadDeployLogs"
          />
        </div>
        <p v-if="deployLogs.length === 0" style="color:#999;font-size:13px;margin-top:12px">暂无部署记录</p>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { getCertificateDetail, deleteCertificate, deployCert, getDeployLogs } from "../api"
import api from "../api"

const route = useRoute()
const router = useRouter()
const id = Number(route.query.id) || 0

const loading = ref(true)
const error = ref("")
const cert = ref<any>(null)
const downloading = ref("")
const deploying = ref(false)
const deployLogs = ref<any[]>([])
const deployLogPage = ref(1)
const deployLogPageSize = ref(5)
const deployLogTotal = ref(0)

function formatTime(t: string) {
  if (!t) return "-"
  const d = new Date(t)
  if (isNaN(d.getTime())) return "-"
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function statusTagType(s: string) {
  if (s === "issued") return "success"
  if (s === "issuing") return "warning"
  if (s === "verifying") return "info"
  return "danger"
}

async function loadDetail() {
  if (!id) { error.value = "无效的证书 ID"; loading.value = false; return }
  try {
    const { data } = await getCertificateDetail(id)
    cert.value = data
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "加载失败"
  } finally {
    loading.value = false
  }
}

async function downloadFile(type: string, ext: string) {
  downloading.value = type
  try {
    const resp = await api.get("/certs/download", {
      params: { id, type },
      responseType: "blob",
    })
    const url = URL.createObjectURL(new Blob([resp.data]))
    const a = document.createElement("a")
    a.href = url
    a.download = `${cert.value.domain}.${ext}`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success("下载成功")
  } catch (e: any) {
    ElMessage.error("下载失败")
  } finally {
    downloading.value = ""
  }
}

async function doDeploy() {
  deploying.value = true
  try {
    await deployCert(id)
    ElMessage.success("部署已启动")
    deployLogPage.value = 1
    setTimeout(() => loadDeployLogs(), 2000)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "部署失败")
  } finally {
    deploying.value = false
  }
}

async function doDelete() {
  try {
    await deleteCertificate(id)
    ElMessage.success("证书已删除")
    router.push("/certs")
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "删除失败")
  }
}

async function loadDeployLogs() {
  if (!id) return
  try {
    const { data } = await getDeployLogs({ cert_id: id, page: deployLogPage.value, page_size: deployLogPageSize.value })
    deployLogs.value = data.list || []
    deployLogTotal.value = data.total || 0
  } catch {
    deployLogs.value = []
  }
}

onMounted(() => { loadDetail(); loadDeployLogs() })
</script>

<style scoped>
.detail-page { width: 100%; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.section {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 16px;
}
.section h3 { margin: 0 0 16px 0; font-size: 16px; font-weight: 600; }
</style>
