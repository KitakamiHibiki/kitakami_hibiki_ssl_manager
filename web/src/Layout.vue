<template>
  <el-container class="layout">
    <el-aside width="200px">
      <div class="logo">KH SSL Manager</div>
      <el-menu router :default-active="$route.path" background-color="#304156" text-color="#bfcbd9" active-text-color="#409eff">
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/domains">
          <el-icon><Link /></el-icon>
          <span>域名管理</span>
        </el-menu-item>
        <el-menu-item index="/certs">
          <el-icon><Tickets /></el-icon>
          <span>证书管理</span>
        </el-menu-item>
        <el-menu-item index="/deploy">
          <el-icon><Upload /></el-icon>
          <span>部署节点</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <span>KH SSL Manager</span>
        <div class="user-area">
          <span class="username">{{ username }}</span>
          <el-button text @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Monitor, Link, Tickets, Upload, User } from '@element-plus/icons-vue'
import { useAuth } from './stores/auth'

const router = useRouter()
const { state, logout } = useAuth()
const username = computed(() => state.username)
const isAdmin = computed(() => state.role === 'admin')

function handleLogout() {
  logout()
  router.replace('/login')
}
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
.layout { height: 100vh; }
.el-aside { background-color: #304156; overflow: hidden; }
.logo { height: 60px; line-height: 60px; text-align: center; color: #fff; font-size: 18px; font-weight: bold; border-bottom: 1px solid #1d2b3a; }
.topbar { background: #fff; border-bottom: 1px solid #e6e6e6; display: flex; justify-content: space-between; align-items: center; font-size: 16px; }
.user-area { display: flex; align-items: center; gap: 12px; }
.username { color: #666; font-size: 14px; }
.el-main { background: #f0f2f5; padding: 20px; }
.el-menu { border-right: none; }
</style>
