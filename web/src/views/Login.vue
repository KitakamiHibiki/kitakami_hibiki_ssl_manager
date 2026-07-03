<template>
  <div class="login-wrapper">
    <el-card class="login-card">
      <h2>{{ isLogin ? "登录" : "注册" }}</h2>
      <el-form>
        <el-form-item>
          <el-autocomplete
            v-model="form.email"
            :fetch-suggestions="queryEmailSuffix"
            placeholder="邮箱"
            style="width: 100%"
            @select="(item: any) => form.email = item.value"
          />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" style="width:100%" @click="submit">
            {{ isLogin ? "登录" : "注册" }}
          </el-button>
        </el-form-item>
      </el-form>
      <p class="toggle">
        {{ isLogin ? "没有账号？" : "已有账号？" }}
        <a href="#" @click.prevent="isLogin = !isLogin">{{ isLogin ? "去注册" : "去登录" }}</a>
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { ElMessage } from "element-plus"
import { authLogin, authRegister } from "../api"
import { useAuth } from "../stores/auth"

const router = useRouter()
const { login } = useAuth()

const isLogin = ref(true)
const loading = ref(false)
const form = ref({ email: "", password: "" })

const emailSuffixes = ["@gmail.com", "@outlook.com", "@hotmail.com", "@qq.com", "@163.com", "@126.com", "@foxmail.com", "@aliyun.com", "@yeah.net", "@sina.com", "@sohu.com", "@icloud.com"]

function queryEmailSuffix(query: string, cb: (list: { value: string }[]) => void) {
  const idx = query.indexOf("@")
  if (idx === -1 || idx === query.length - 1) return cb([])
  const typed = query.substring(idx + 1)
  const prefix = query.substring(0, idx + 1)
  const filtered = emailSuffixes
    .filter((s) => s.substring(1).startsWith(typed))
    .map((s) => ({ value: prefix + s.substring(1) }))
  cb(filtered)
}

async function submit() {
  if (!form.value.email || !form.value.password) {
    ElMessage.warning("请填写邮箱和密码")
    return
  }
  loading.value = true
  try {
    const fn = isLogin.value ? authLogin : authRegister
    const { data } = await fn(form.value.email, form.value.password)
    login(data.token, data.username, data.role, data.email)
    ElMessage.success(isLogin.value ? "登录成功" : "注册成功")
    router.replace("/")
  } catch (e: any) {
    if (e?.response?.status === 401) {
      ElMessage.error("账号或密码不正确")
    } else {
      ElMessage.error(e?.response?.data?.msg || "操作失败")
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #f0f2f5;
}
.login-card { width: 400px; }
.login-card h2 { text-align: center; margin-bottom: 24px; }
.toggle { text-align: center; font-size: 14px; color: #999; }
.toggle a { color: #409eff; }
</style>