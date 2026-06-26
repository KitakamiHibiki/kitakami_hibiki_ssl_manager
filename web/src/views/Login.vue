<template>
  <div class="login-wrapper">
    <el-card class="login-card">
      <h2>{{ isLogin ? '登录' : '注册' }}</h2>
      <el-form @submit.prevent="submit">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
            {{ isLogin ? '登录' : '注册' }}
          </el-button>
        </el-form-item>
      </el-form>
      <p class="toggle">
        {{ isLogin ? '没有账号？' : '已有账号？' }}
        <a href="#" @click.prevent="isLogin = !isLogin">{{ isLogin ? '去注册' : '去登录' }}</a>
      </p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authLogin, authRegister } from '../api'
import { useAuth } from '../stores/auth'

const router = useRouter()
const { login } = useAuth()

const isLogin = ref(true)
const loading = ref(false)
const form = ref({ username: '', password: '' })

async function submit() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  loading.value = true
  try {
    const fn = isLogin.value ? authLogin : authRegister
    const { data } = await fn(form.value.username, form.value.password)
    login(data.token, data.username, data.role)
    ElMessage.success(isLogin.value ? '登录成功' : '注册成功')
    router.replace('/')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  display: flex; justify-content: center; align-items: center; height: 100vh; background: #f0f2f5;
}
.login-card {
  width: 400px;
}
.login-card h2 {
  text-align: center; margin-bottom: 24px;
}
.toggle {
  text-align: center; font-size: 14px; color: #999;
}
.toggle a { color: #409eff; }
</style>
