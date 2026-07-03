<template>
  <div class="add-domain-page">
    <div class="header">
      <h2>添加域名</h2>
      <el-button text @click="$router.push('/domains')">返回列表</el-button>
    </div>

    <el-form :model="form" label-width="80px" style="max-width:500px" @submit.prevent="submitDomain">
      <el-form-item label="域名">
        <el-input v-model="form.domain" placeholder="example.com" />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="form.email" :placeholder="authState.email || 'admin@example.com'" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" size="large" :disabled="!form.domain" :loading="loading" native-type="submit">
          提交
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { useAuth } from "../stores/auth"
import { ElMessage } from "element-plus"
import { createDomain } from "../api"

const router = useRouter()
const { state: authState } = useAuth()
const loading = ref(false)
const form = ref({ domain: "", email: "" })

async function submitDomain() {
  if (!form.value.domain) return
  loading.value = true
  const email = form.value.email || authState.email || ""
  try {
    const { data } = await createDomain({ domain: form.value.domain, email, challenge: "dns" })
    ElMessage.success("域名添加成功")
    router.push(`/domains/detail?id=${data.id}`)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "添加失败")
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; font-size: 20px; font-weight: 600; }
</style>
