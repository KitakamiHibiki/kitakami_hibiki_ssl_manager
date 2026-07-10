<template>
  <div class="cert-apply-page">
    <div class="header">
      <h2>证书申请</h2>
      <el-button text @click="$router.push(`/domains/detail?id=${id}`)">返回域名详情</el-button>
    </div>

    <div v-if="loading" style="color:#999">加载中...</div>
    <template v-else>
      <div class="section">
        <p>域名：<strong>{{ domain }}</strong></p>

        <div v-if="!applied" style="margin-top:16px">
          <h4>附加域名（SAN）</h4>
          <p style="color:#999;font-size:13px;margin-bottom:12px">
            可选添加通配域名或其它子域名，如 *.example.com、www.example.com
          </p>
          <div v-for="(_, i) in extraDomains" :key="i" style="display:flex;gap:8px;margin-bottom:8px">
            <el-input v-model="extraDomains[i]" placeholder="*.example.com" style="width:260px" />
            <el-button type="danger" size="small" @click="extraDomains.splice(i,1)">删除</el-button>
          </div>
          <el-button size="small" @click="extraDomains.push('')">+ 添加域名</el-button>
          <div style="margin-top:16px">
            <el-button type="primary" @click="startApply" :loading="applying">开始申请</el-button>
          </div>
        </div>
      </div>

      <div v-if="error" class="section" style="color:#f56c6c">
        <h3>错误</h3>
        <p>{{ error }}</p>
      </div>

      <div v-else-if="applied" class="section">
        <h3>DNS 验证</h3>
        <p style="color:#666;margin-bottom:12px">请为以下所有域名添加 DNS TXT 记录：</p>

        <div v-for="d in allDomains" :key="d" class="challenge-box" style="margin-bottom:16px">
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
            <strong style="font-size:14px">{{ d }}</strong>
            <el-tag v-if="verifiedDomains[d]" type="success" size="small">已验证</el-tag>
            <el-tag v-else-if="verifyingDomains[d]" type="warning" size="small">验证中...</el-tag>
            <el-tag v-else type="info" size="small">待验证</el-tag>
          </div>

          <div v-if="challengeValues[d]">
            <div class="challenge-row">
              <span class="challenge-label">主机记录：</span>
              <code>_acme-challenge.{{ d }}</code>
              <el-button size="small" @click="copyText('_acme-challenge.' + d)">复制</el-button>
            </div>
            <div class="challenge-row">
              <span class="challenge-label">记录值：</span>
              <code style="word-break:break-all">{{ challengeValues[d] }}</code>
              <el-button size="small" @click="copyText(challengeValues[d])">复制</el-button>
            </div>
          </div>
          <div v-else style="color:#999;font-size:13px">正在获取挑战值...</div>

          <el-button
            v-if="challengeValues[d] && !verifiedDomains[d]"
            size="small"
            type="primary"
            style="margin-top:8px"
            @click="verifyOne(d)"
            :loading="verifyingDomains[d]"
            :disabled="!!verifiedDomains[d]"
          >
            验证 {{ d }}
          </el-button>
        </div>

        <div v-if="allVerified" style="margin-top:16px">
          <el-button type="success" disabled>全部验证通过，正在签发...</el-button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { getDomainDetail, getChallengeValue, getCertStatus } from "../api"
import api from "../api"

const route = useRoute()
const router = useRouter()
const id = Number(route.query.id) || 0

const loading = ref(true)
const error = ref("")
const domain = ref("")
const extraDomains = ref<string[]>([])
const allDomains = ref<string[]>([])
const applied = ref(false)
const applying = ref(false)
const challengeValues = reactive<Record<string, string>>({})
const verifiedDomains = reactive<Record<string, boolean>>({})
const verifyingDomains = reactive<Record<string, boolean>>({})
const allVerified = ref(false)

let challengeTimer: number | null = null
let statusTimer: number | null = null
let challengeRetry: Record<string, number> = {}

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => ElMessage.success("已复制"))
}

async function load() {
  if (!id) { error.value = "无效的域名 ID"; loading.value = false; return }
  try {
    const { data } = await getDomainDetail(id)
    domain.value = data.domain.domain
    loading.value = false
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "加载失败"
    loading.value = false
  }
}

onMounted(load)
onUnmounted(() => {
  if (challengeTimer !== null) { clearInterval(challengeTimer); challengeTimer = null }
  if (statusTimer !== null) { clearInterval(statusTimer); statusTimer = null }
})

async function startApply() {
  applying.value = true
  const extras = extraDomains.value.filter(d => d.trim() !== "")
  try {
    const { data } = await api.post("/certs/apply", {
      domain_id: id,
      extra_domains: extras,
    })
    allDomains.value = data.domains || [domain.value]
    applied.value = true
    startChallengePoll()
    startStatusPoll()
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "申请失败"
  } finally {
    applying.value = false
  }
}

async function verifyOne(d: string) {
  verifyingDomains[d] = true
  try {
    const { data } = await api.post("/certs/verify-dns", { domain: d })
    verifiedDomains[d] = true
    if (data.all_verified) {
      allVerified.value = true
      const certId = data.cert_id
      stopAllPolls()
      setTimeout(() => {
        router.replace(`/domains/detail/cert-download?domain_id=${id}&cert_id=${certId}`)
      }, 500)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "DNS 验证失败")
  } finally {
    verifyingDomains[d] = false
  }
}

function stopAllPolls() {
  if (challengeTimer !== null) { clearInterval(challengeTimer); challengeTimer = null }
  if (statusTimer !== null) { clearInterval(statusTimer); statusTimer = null }
}

function startChallengePoll() {
  challengeRetry = {}
  challengeTimer = window.setInterval(async () => {
    for (const d of allDomains.value) {
      if (challengeValues[d] || verifiedDomains[d]) continue
      challengeRetry[d] = (challengeRetry[d] || 0) + 1
      try {
        const { data } = await getChallengeValue(d)
        if (data.challenge_value) {
          challengeValues[d] = data.challenge_value
        }
      } catch (e: any) {
        if (e?.code === 404 && challengeRetry[d] >= 15) {
          stopAllPolls()
          error.value = d + " 挑战值获取超时"
          return
        }
      }
    }
  }, 2000)
}

function startStatusPoll() {
  statusTimer = window.setInterval(async () => {
    try {
      const { data } = await getCertStatus(domain.value)
      if (data.status === "error") {
        stopAllPolls()
        error.value = data.error_msg || "证书申请失败"
      }
    } catch {}
  }, 3000)
}
</script>

<style scoped>
.cert-apply-page { width: 100%; }
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
.challenge-box {
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 16px;
}
.challenge-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-family: monospace;
  font-size: 13px;
}
.challenge-row:last-child { margin-bottom: 0; }
.challenge-label {
  font-weight: 600;
  font-family: sans-serif;
  white-space: nowrap;
}
</style>
