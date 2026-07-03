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
      </div>

      <div v-if="error" class="section" style="color:#f56c6c">
        <h3>错误</h3>
        <p>{{ error }}</p>
      </div>

      <div v-else class="section">
        <h3>DNS 验证</h3>
        <div v-if="!challengeValue" style="color:#999">
          正在获取 DNS 挑战值...
        </div>
        <template v-else>
          <p style="color:#666;margin-bottom:12px">请在 DNS 管理中添加以下 TXT 记录：</p>
          <div class="challenge-box">
            <div class="challenge-row">
              <span class="challenge-label">主机记录：</span>
              <code>_acme-challenge.{{ domain }}</code>
              <el-button size="small" @click="copyText('_acme-challenge.' + domain)">复制</el-button>
            </div>
            <div class="challenge-row">
              <span class="challenge-label">记录值：</span>
              <code style="word-break:break-all">{{ challengeValue }}</code>
              <el-button size="small" @click="copyText(challengeValue)">复制</el-button>
            </div>
          </div>
          <p style="color:#999;font-size:13px;margin:12px 0">
            添加记录后点击下方按钮进行 DNS 认证。
          </p>
          <el-button type="primary" @click="dnsVerify" :loading="verifying" :disabled="verified">
            {{ verified ? 'DNS 认证通过' : 'DNS 认证' }}
          </el-button>
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue"
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { getDomainDetail, applyCertificate, getChallengeValue, getCertStatus, verifyDNS } from "../api"

const route = useRoute()
const router = useRouter()
const id = Number(route.query.id) || 0

const loading = ref(true)
const error = ref("")
const domain = ref("")
const challengeValue = ref("")
const verifying = ref(false)
const verified = ref(false)

let challengeTimer: number | null = null
let statusTimer: number | null = null
let retryCount = 0

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => ElMessage.success("已复制"))
}

async function load() {
  if (!id) { error.value = "无效的域名 ID"; loading.value = false; return }
  try {
    const { data } = await getDomainDetail(id)
    domain.value = data.domain.domain
    loading.value = false
    await startApply()
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
  try {
    await applyCertificate(id)
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "申请失败"
    return
  }
  startChallengePoll()
  startStatusPoll()
}

async function dnsVerify() {
  verifying.value = true
  try {
    const { data } = await verifyDNS(domain.value)
    verified.value = true
    router.replace(`/domains/detail/cert-download?domain_id=${id}&cert_id=${data.cert_id}`)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "DNS 认证失败")
  } finally {
    verifying.value = false
  }
}

function stopAllPolls() {
  if (challengeTimer !== null) { clearInterval(challengeTimer); challengeTimer = null }
  if (statusTimer !== null) { clearInterval(statusTimer); statusTimer = null }
}

function startChallengePoll() {
  retryCount = 0
  challengeTimer = window.setInterval(async () => {
    try {
      const { data } = await getChallengeValue(domain.value)
      if (data.challenge_value) {
        challengeValue.value = data.challenge_value
        if (challengeTimer !== null) { clearInterval(challengeTimer); challengeTimer = null }
      }
    } catch (e: any) {
      retryCount++
      const code = e?.code
      if (code === 404) {
        if (retryCount >= 15) {
          stopAllPolls()
          error.value = "获取挑战值超时，请检查 ACME 服务是否可达"
        }
        return
      }
      stopAllPolls()
      error.value = e?.message || "获取挑战值失败"
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
