<template>
  <div class="certs-page">
    <!-- 证书列表 -->
    <div class="section">
      <div class="section-header">
        <h3>证书列表</h3>
        <el-button type="primary" @click="$router.push('/certs/apply')">申请证书</el-button>
      </div>
      <el-table :data="certs" style="width:100%" v-loading="certsLoading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column label="域名" min-width="160">
          <template #default="{ row }">
            <span class="domain-cell">
              {{ row.domain }}
              <el-icon class="copy-icon" @click="copyText(row.domain)"><CopyDocument /></el-icon>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="签发时间" width="170">
          <template #default="{ row }">{{ row.status === 'issued' ? formatTime(row.issued_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="到期时间" width="170">
          <template #default="{ row }">{{ row.status === 'issued' ? formatTime(row.expires_at) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="error_msg" label="错误信息" min-width="140">
          <template #default="{ row }">
            <span v-if="row.error_msg" style="color:#f56c6c;font-size:13px">{{ row.error_msg }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="$router.push('/certs/detail?id=' + row.id)">详情</el-button>
            <el-button v-if="row.status === 'issued'" size="small" type="success" @click="$router.push('/certs/download?cert_id=' + row.id)">下载</el-button>
            <el-popconfirm title="确定删除此证书？" @confirm="deleteCert(row.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top:16px;text-align:right" v-if="certTotal > 0">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="certTotal"
          layout="total, prev, pager, next"
          @current-change="loadCerts"
        />
      </div>
      <p v-if="certs.length === 0 && !certsLoading" style="color:#999;font-size:13px;margin-top:12px">暂无证书</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { ElMessage } from "element-plus"
import { CopyDocument } from "@element-plus/icons-vue"
import { getCertificates, deleteCertificate } from "../api"

const certs = ref<any[]>([])
const certsLoading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const certTotal = ref(0)

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => ElMessage.success("已复制"))
}

function formatTime(t: string) {
  if (!t) return ""
  const d = new Date(t)
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function statusTagType(s: string) {
  if (s === "issued") return "success"
  if (s === "issuing") return "warning"
  if (s === "verifying") return "info"
  return "danger"
}

async function loadCerts() {
  certsLoading.value = true
  try {
    const { data } = await getCertificates({
      page: page.value,
      page_size: pageSize.value,
    })
    certs.value = data.list || []
    certTotal.value = data.total || 0
  } catch {
    certs.value = []
  } finally {
    certsLoading.value = false
  }
}

async function deleteCert(id: number) {
  try {
    await deleteCertificate(id)
    ElMessage.success("证书已删除")
    loadCerts()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "删除失败")
  }
}

onMounted(() => {
  loadCerts()
})
</script>

<style scoped>
.certs-page { width: 100%; }
.section {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 16px;
}
.section h3 { margin: 0; font-size: 16px; font-weight: 600; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.section-header h3 { margin: 0; }
.domain-cell { display: inline-flex; align-items: center; gap: 6px; }
.copy-icon { cursor: pointer; color: #999; font-size: 14px; }
.copy-icon:hover { color: #409eff; }
</style>
