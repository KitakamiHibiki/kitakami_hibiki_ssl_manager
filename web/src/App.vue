<template>
  <el-container class="layout">
    <el-aside width="200px">
      <div class="logo">SSL Manager</div>
      <el-menu router :default-active="$route.path" background-color="#304156" text-color="#bfcbd9" active-text-color="#409eff">
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/domains">
          <el-icon><Link /></el-icon>
          <span>域名管理</span>
        </el-menu-item>
        <el-menu-item index="/certificates">
          <el-icon><Document /></el-icon>
          <span>证书管理</span>
        </el-menu-item>
        <el-menu-item index="/deploy">
          <el-icon><Upload /></el-icon>
          <span>部署管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>kitakami_hibiki SSL Manager</el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { Monitor, Link, Document, Upload } from '@element-plus/icons-vue'
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
.layout { height: 100vh; }
.el-aside { background-color: #304156; overflow: hidden; }
.logo { height: 60px; line-height: 60px; text-align: center; color: #fff; font-size: 18px; font-weight: bold; border-bottom: 1px solid #1d2b3a; }
.el-header { background: #fff; border-bottom: 1px solid #e6e6e6; line-height: 60px; font-size: 16px; padding: 0 20px; }
.el-main { background: #f0f2f5; padding: 20px; }
.el-menu { border-right: none; }
</style>
