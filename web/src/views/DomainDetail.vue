<template>
  <div class="detail-page">
    <div class="header">
      <h2>域名详情</h2>
      <el-button text @click="$router.push('/domains')">返回列表</el-button>
    </div>

    <div v-if="loading" style="color:#999">加载中...</div>
    <div v-else-if="error" style="color:#f56c6c">{{ error }}</div>
    <template v-else-if="detail">
      <div class="section">
        <h3>基本信息</h3>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="域名">
            <strong>{{ detail.domain.domain }}</strong>
          </el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ detail.domain.email }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(detail.domain.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(detail.domain.updated_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="section">
        <h3>操作</h3>
        <div style="display:flex;align-items:center;gap:24px;flex-wrap:wrap">
          <el-button type="success" @click="$router.push(`/domains/detail/cert-apply?id=${id}`)">
            申请证书
          </el-button>
          <el-button type="danger" @click="remove">删除域名</el-button>
          <div style="display:flex;align-items:center;gap:8px">
            <span style="font-size:14px">自动部署</span>
            <el-switch v-model="deployForm.enabled" @change="saveDeploy" :loading="saving" />
          </div>
          <div style="display:flex;align-items:center;gap:8px">
            <span style="font-size:14px">自动续签</span>
            <el-switch v-model="deployForm.auto_renew" @change="saveDeploy" :loading="saving" />
          </div>
        </div>
      </div>

      <div class="section">
        <div class="section-header">
          <h3>部署配置</h3>
          <el-button v-if="!editing" size="small" @click="editing = true">修改</el-button>
          <el-button v-else type="primary" size="small" @click="saveDeploy" :loading="saving">保存</el-button>
        </div>
        <el-form label-width="100px">
          <el-form-item label="目标节点">
            <el-select v-model="deployForm.node_id" placeholder="选择节点" :disabled="!editing" style="width:240px">
              <el-option v-for="n in nodes" :key="n.id" :label="n.name" :value="n.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="服务器类型">
            <el-select v-model="deployForm.type" :disabled="!editing" style="width:240px">
              <el-option label="Nginx" value="nginx" />
              <el-option label="Apache" value="apache" disabled />
              <el-option label="其他" value="other" disabled />
            </el-select>
          </el-form-item>
          <el-form-item label="证书文件名">
            <el-input v-model="deployForm.cert_name" placeholder="fullchain.pem" :disabled="!editing" style="width:240px" />
          </el-form-item>
          <el-form-item label="证书部署路径">
            <el-input v-model="deployForm.cert_path" placeholder="/etc/nginx/certs" :disabled="!editing" style="width:240px" />
          </el-form-item>
          <el-form-item label="私钥文件名">
            <el-input v-model="deployForm.key_name" placeholder="privkey.key" :disabled="!editing" style="width:240px" />
          </el-form-item>
          <el-form-item label="私钥部署路径">
            <el-input v-model="deployForm.key_path" placeholder="/etc/nginx/certs" :disabled="!editing" style="width:240px" />
          </el-form-item>
        </el-form>
      </div>

      <div class="section">
        <h3>证书列表</h3>
        <el-table :data="certs" style="width:100%">
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="签发时间" width="180">
            <template #default="{ row }">{{ row.status === 'issued' ? formatTime(row.issued_at) : '-' }}</template>
          </el-table-column>
          <el-table-column label="到期时间" width="180">
            <template #default="{ row }">{{ row.status === 'issued' ? formatTime(row.expires_at) : '-' }}</template>
          </el-table-column>
          <el-table-column prop="error_msg" label="错误信息" min-width="200">
            <template #default="{ row }">
              <span v-if="row.error_msg" style="color:#f56c6c;font-size:13px">{{ row.error_msg }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.status === 'issued'" type="primary" size="small" @click="$router.push(`/domains/detail/cert-download?domain_id=${id}&cert_id=${row.id}`)">
                下载
              </el-button>
              <el-button v-if="row.status === 'issued'" type="success" size="small" @click="doDeploy(row.id)" :loading="deployingId === row.id">
                部署
              </el-button>
            </template>
          </el-table-column>
        </el-table>
        <div style="margin-top:16px;text-align:right" v-if="certTotal > 0">
          <el-pagination
            v-model:current-page="certPage"
            :page-size="certPageSize"
            :total="certTotal"
            layout="total, prev, pager, next"
            @current-change="loadCerts"
          />
        </div>
        <p v-if="certs.length === 0 && !loading" style="color:#999">暂无证书</p>
      </div>

      <div class="section">
        <h3>部署记录</h3>
        <el-table :data="deployLogs" size="small" style="width:100%">
          <el-table-column prop="id" label="ID" width="50" />
          <el-table-column prop="cert_id" label="证书ID" width="70" />
          <el-table-column prop="node_name" label="目标节点" width="120" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : row.status === 'pending' ? 'info' : 'danger'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="error_msg" label="详情" min-width="200">
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
import { ref, reactive, onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { getDomainDetail, deleteDomain, updateDomain, getNodes, deployCert, getDeployLogs } from "../api"
import api from "../api"

const route = useRoute()
const router = useRouter()
const id = Number(route.query.id) || 0

const loading = ref(true)
const error = ref("")
const detail = ref<{ domain: any } | null>(null)
const nodes = ref<any[]>([])
const certs = ref<any[]>([])
const saving = ref(false)
const certPage = ref(1)
const certPageSize = ref(5)
const certTotal = ref(0)
const deployingId = ref(0)
const editing = ref(false)
const deployLogs = ref<any[]>([])
const deployLogPage = ref(1)
const deployLogPageSize = ref(5)
const deployLogTotal = ref(0)

const deployForm = reactive({
  enabled: false,
  node_id: 0,
  type: "nginx",
  cert_name: "fullchain.pem",
  cert_path: "/etc/nginx/certs",
  key_name: "privkey.key",
  key_path: "/etc/nginx/certs",
  auto_renew: false,
})

function formatTime(t: string) {
  if (!t) return "-"
  const d = new Date(t)
  if (isNaN(d.getTime())) return "-"
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function statusTagType(s: string) {
  if (s === "issued") return "success"
  if (s === "issuing") return "warning"
  if (s === "verifying") return "info"
  return "danger"
}

async function loadDetail() {
  if (!id) { error.value = "无效的域名 ID"; loading.value = false; return }
  try {
    const domainRes = await getDomainDetail(id)
    const d = domainRes.data.domain
    detail.value = { domain: d }
    deployForm.enabled = d.deploy_enabled
    deployForm.node_id = d.deploy_node_id
    deployForm.type = d.deploy_type || "nginx"
    deployForm.cert_name = d.cert_name || "fullchain.pem"
    deployForm.cert_path = d.cert_path || "/etc/nginx/certs"
    deployForm.key_name = d.key_name || "privkey.key"
    deployForm.key_path = d.key_path || "/etc/nginx/certs"
    deployForm.auto_renew = d.auto_renew || false
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "加载失败"
  } finally {
    loading.value = false
  }
}

async function loadCerts() {
  try {
    const { data } = await api.get("/certs", { params: { page: certPage.value, page_size: certPageSize.value, domain_id: id } })
    certs.value = data.list || []
    certTotal.value = data.total || 0
  } catch {}
}

async function loadNodes() {
  try {
    const { data } = await getNodes()
    nodes.value = data || []
  } catch {}
}

onMounted(() => { loadDetail(); loadCerts(); loadNodes(); loadDeployLogs() })

async function saveDeploy() {
  saving.value = true
  try {
    await updateDomain(id, {
      deploy_enabled: deployForm.enabled,
      deploy_node_id: deployForm.node_id,
      deploy_type: deployForm.type,
      cert_name: deployForm.cert_name,
      cert_path: deployForm.cert_path,
      key_name: deployForm.key_name,
      key_path: deployForm.key_path,
      auto_renew: deployForm.auto_renew,
    })
    ElMessage.success("保存成功")
    editing.value = false
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "保存失败")
  } finally {
    saving.value = false
  }
}

async function remove() {
  await deleteDomain(id)
  ElMessage.success("域名已删除")
  router.push("/domains")
}

async function doDeploy(certId: number) {
  deployingId.value = certId
  try {
    await deployCert(certId)
    ElMessage.success("部署已启动")
    deployLogPage.value = 1
    setTimeout(() => loadDeployLogs(), 2000)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "部署失败")
  } finally {
    deployingId.value = 0
  }
}

async function loadDeployLogs() {
  if (!id) return
  try {
    const { data } = await getDeployLogs({ domain_id: id, page: deployLogPage.value, page_size: deployLogPageSize.value })
    deployLogs.value = data.list || []
    deployLogTotal.value = data.total || 0
  } catch {
    deployLogs.value = []
  }
}
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
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.section-header h3 { margin: 0; font-size: 16px; font-weight: 600; }
:deep(.el-table__body tr) { height: 48px; }
</style>
