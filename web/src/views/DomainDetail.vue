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
        <el-button type="success" @click="$router.push(`/domains/detail/cert-apply?id=${id}`)">
          申请证书
        </el-button>
        <el-button type="danger" @click="remove">删除域名</el-button>
      </div>

      <div class="section">
        <h3>部署</h3>
        <el-form label-width="100px">
          <el-form-item label="启用部署">
            <el-switch v-model="deployForm.enabled" />
          </el-form-item>
          <el-form-item label="目标节点">
            <el-select v-model="deployForm.node_id" placeholder="选择节点" :disabled="!deployForm.enabled" style="width:240px">
              <el-option v-for="n in nodes" :key="n.id" :label="n.name" :value="n.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="部署路径">
            <el-input v-model="deployForm.path" placeholder="/etc/nginx/certs" :disabled="!deployForm.enabled" style="width:240px" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveDeploy" :loading="saving">保存</el-button>
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
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { getDomainDetail, deleteDomain, updateDomain, getNodes } from "../api"
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
const certPageSize = ref(10)
const certTotal = ref(0)

const deployForm = reactive({
  enabled: false,
  node_id: 0,
  path: "",
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
    deployForm.path = d.deploy_path
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "加载失败"
  } finally {
    loading.value = false
  }
}

async function loadCerts() {
  try {
    const { data } = await api.get("/certs", { params: { page: certPage.value, page_size: certPageSize.value } })
    certs.value = (data.list || []).filter((c: any) => c.domain_id === id)
    certTotal.value = data.total || 0
  } catch {}
}

async function loadNodes() {
  try {
    const { data } = await getNodes()
    nodes.value = data || []
  } catch {}
}

onMounted(() => { loadDetail(); loadCerts(); loadNodes() })

async function saveDeploy() {
  saving.value = true
  try {
    await updateDomain(id, {
      deploy_enabled: deployForm.enabled,
      deploy_node_id: deployForm.node_id,
      deploy_path: deployForm.path,
    })
    ElMessage.success("保存成功")
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
