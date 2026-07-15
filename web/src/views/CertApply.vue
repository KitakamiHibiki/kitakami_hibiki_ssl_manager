<template>
  <div class="cert-apply-page">
    <div class="header">
      <h2>证书申请</h2>
      <el-button text @click="$router.push('/certs')">返回证书列表</el-button>
    </div>

    <!-- Step 1: 填写域名 -->
    <div v-if="!applied" class="section">
      <div class="cert-type-bar">
        <el-radio-group v-model="certType" size="large">
          <el-radio-button value="single">单域名证书</el-radio-button>
          <el-radio-button value="wildcard">泛域名证书</el-radio-button>
          <el-radio-button value="multi">多域名证书</el-radio-button>
        </el-radio-group>
      </div>

      <el-form label-width="80px" style="margin-top:20px">
        <el-form-item label="域名" required>
          <el-input
            v-if="certType !== 'multi'"
            v-model="domainInput"
            :placeholder="singlePlaceholder"
            style="width:360px"
          />
          <el-input
            v-else
            v-model="domainInput"
            type="textarea"
            :rows="5"
            :placeholder="multiPlaceholder"
          />
        </el-form-item>
      </el-form>

      <div v-if="certType === 'wildcard'" class="hint">
        泛域名证书将保护 <code>*.example.com</code> 及 <code>example.com</code> 本身，输入时请包含 <code>*.</code> 前缀。
      </div>
      <div v-else-if="certType === 'multi'" class="hint">
        每行填写一个域名，所有域名必须属于同一个根域名（如 example.com）。支持通配符。
      </div>
      <div v-else class="hint">
        输入单个完整域名，如 <code>example.com</code> 或 <code>www.example.com</code>。
      </div>

      <div style="margin-top:20px">
        <el-button type="primary" size="large" @click="goNext" :disabled="!domainInput.trim()">下一步</el-button>
      </div>
    </div>

    <div v-if="error" class="section" style="color:#f56c6c">
      <h3>错误</h3>
      <p>{{ error }}</p>
    </div>

    <!-- Step 2: 选择认证方式 + 验证 -->
    <div v-else-if="applied && !allVerified" class="section">
      <h3>认证方式</h3>
      <div class="cert-type-bar">
        <el-radio-group v-model="challengeType" size="large" @change="onChallengeTypeChange">
          <el-radio-button value="http">HTTP 认证</el-radio-button>
          <el-radio-button value="cname">免DNS认证</el-radio-button>
        </el-radio-group>
      </div>

      <!-- HTTP 认证说明 -->
      <div v-if="challengeType === 'http'" class="challenge-hint">
        <p><strong>代理验证模式</strong></p>
        <p>请在服务端（CDN、Nginx、OpenResty、HTTPd 等）中将 HTTP 验证请求路径代理至本服务器，以实现证书申请及后续更新的自动化：</p>
        <div class="proxy-info">
          <div class="challenge-row">
            <span class="challenge-label">代理路径：</span>
            <code>/.well-known/acme-challenge/</code>
          </div>
          <div class="challenge-row">
            <span class="challenge-label">目标地址：</span>
            <code>{{ proxyBaseUrl }}/.well-known/acme-challenge/</code>
          </div>
        </div>
        <p style="font-weight:600;margin-top:12px">Nginx 配置示例：</p>
        <pre class="nginx-config">location ^~ /.well-known/acme-challenge/ {
    proxy_pass {{ proxyBaseUrl }}/.well-known/acme-challenge/;
}</pre>
        <p style="color:#e6a23c;font-size:13px;margin-top:8px">
          ⚠️ 请不要删除以上服务端代理配置，否则会导致相关证书生成失败或自动更新失败。
        </p>
      </div>

      <!-- 免DNS认证说明 -->
      <div v-else class="challenge-hint">
        <p><strong>操作方式：</strong>前往您的 DNS 服务商，为以下域名添加 CNAME 记录，将验证委托给本系统处理。</p>
        <p style="color:#909399;font-size:13px;margin-top:4px">添加后点击"验证"按钮，系统会自动完成 DNS-01 挑战。</p>
      </div>

      <!-- 域名校验列表 -->
      <div v-for="d in allDomains" :key="d" class="challenge-box" style="margin-bottom:16px">
        <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
          <strong style="font-size:14px">{{ d }}</strong>
          <el-tag v-if="verifiedDomains[d]" type="success" size="small">已验证</el-tag>
          <el-tag v-else-if="verifyingDomains[d]" type="warning" size="small">验证中...</el-tag>
          <el-tag v-else type="info" size="small">待验证</el-tag>
        </div>

        <!-- HTTP 模式 -->
        <template v-if="challengeType === 'http'">
          <div class="challenge-row">
            <span class="challenge-label">验证 URL：</span>
            <code style="word-break:break-all">http://{{ d }}/.well-known/acme-challenge/&lt;token&gt;</code>
          </div>
        </template>
        <!-- 免DNS认证模式 -->
        <template v-else>
          <div class="challenge-row">
            <span class="challenge-label">记录类型：</span>
            <code>CNAME</code>
          </div>
          <div class="challenge-row">
            <span class="challenge-label">主机记录：</span>
            <code>_acme-challenge.{{ d }}</code>
            <el-button size="small" @click="copyText('_acme-challenge.' + d)">复制</el-button>
          </div>
          <div class="challenge-row">
            <span class="challenge-label">记录值：</span>
            <code>{{ cnameTarget() }}</code>
            <el-button size="small" @click="copyText(cnameTarget())">复制</el-button>
          </div>
        </template>

        <el-button
          v-if="!verifiedDomains[d]"
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
    </div>

    <!-- Step 3: 签发完成 -->
    <div v-if="allVerified" class="section">
      <h3 style="color:#67c23a">全部验证通过</h3>
      <p style="color:#666;margin-top:8px">证书正在签发中，请稍候...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue"
import { useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import api from "../api"

const router = useRouter()

const certType = ref("single")
const challengeType = ref("cname")
const domainInput = ref("")
const error = ref("")
const allDomains = ref<string[]>([])
const domainHash = ref("")
const applied = ref(false)
const applying = ref(false)
const verifiedDomains = reactive<Record<string, boolean>>({})
const verifyingDomains = reactive<Record<string, boolean>>({})
const allVerified = ref(false)

const singlePlaceholder = computed(() => certType.value === "wildcard" ? "*.example.com" : "example.com")
const multiPlaceholder = `example.com
www.example.com
api.example.com
*.example.com`

const proxyBaseUrl = ref(window.location.origin)
const proxyHost = computed(() => new URL(proxyBaseUrl.value).hostname)

onMounted(async () => {
  try {
    const { data } = await api.get('/platform')
    if (data.proxy_url) {
      proxyBaseUrl.value = data.proxy_url
    }
  } catch {}
})

function copyText(text: string) {
  navigator.clipboard.writeText(text).then(() => ElMessage.success("已复制"))
}

function cnameTarget(): string {
  return domainHash.value + '.challenge.' + proxyHost.value
}

function parseDomains(): { domain: string; extras: string[] } {
  const raw = domainInput.value.trim()
  if (certType.value === "multi") {
    const lines = raw.split("\n").map(s => s.trim()).filter(s => s !== "")
    if (lines.length === 0) return { domain: "", extras: [] }
    return { domain: lines[0], extras: lines.slice(1) }
  }
  return { domain: raw, extras: [] }
}

function goNext() {
  const { domain, extras } = parseDomains()
  if (!domain) return
  let finalExtras = [...extras]
  if (certType.value === "wildcard") {
    const root = domain.replace(/^\*\./, "")
    if (!finalExtras.includes(root) && root !== domain) {
      finalExtras.push(root)
    }
  }
  startApply(domain, finalExtras)
}

async function startApply(domain: string, extras: string[]) {
  applying.value = true
  try {
    const { data } = await api.post("/certs/apply", {
      domain,
      extra_domains: extras,
      challenge_type: challengeType.value,
    })
    allDomains.value = data.domains || [domain]
    domainHash.value = data.domain_hash || ''
    applied.value = true
  } catch (e: any) {
    error.value = e?.response?.data?.msg || "申请失败"
  } finally {
    applying.value = false
  }
}

function onChallengeTypeChange() {
  // Clear existing challenges when switching mode (would need re-apply)
}


async function verifyOne(d: string) {
  verifyingDomains[d] = true
  try {
    const { data } = await api.post("/certs/verify-dns", { domain: d })
    verifiedDomains[d] = true
    if (data.all_verified) {
      allVerified.value = true
      setTimeout(() => {
        router.replace('/certs')
      }, 1000)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "验证失败")
  } finally {
    verifyingDomains[d] = false
  }
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
.cert-type-bar { text-align: left; }
.hint {
  color: #909399;
  font-size: 13px;
  margin-top: 8px;
  line-height: 1.6;
}
.hint code {
  background: #f5f7fa;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 12px;
}
.challenge-hint {
  color: #666;
  font-size: 14px;
  margin: 16px 0;
  line-height: 1.7;
}
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
.proxy-info {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 12px 16px;
  margin-top: 10px;
}
.nginx-config {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 12px 16px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.6;
  margin: 8px 0 0 0;
  overflow-x: auto;
}
</style>
