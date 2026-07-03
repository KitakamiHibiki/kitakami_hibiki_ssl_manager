<template>
  <div class="nodes-page">
    <div class="header">
      <h2>部署节点</h2>
      <el-button type="primary" @click="openAdd">添加节点</el-button>
    </div>

    <el-table :data="nodes" style="width:100%">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" />
      <el-table-column label="类型" width="80">
        <template #default="{ row }">
          <el-tag>{{ row.node_type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="host" label="主机" />
      <el-table-column prop="port" label="端口" width="80" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button size="small" @click="testConn(row.id)" :loading="testingId === row.id">测试</el-button>
          <el-button size="small" @click="openEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <p v-if="nodes.length === 0 && !loading" style="color:#999;margin-top:20px">暂无节点，请添加</p>

    <el-dialog v-model="showDialog" :title="editingId ? '编辑节点' : '添加节点'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="my-server" />
        </el-form-item>
        <el-form-item label="节点类型">
          <el-select v-model="form.node_type" style="width:100%">
            <el-option label="SSH" value="ssh" />
          </el-select>
        </el-form-item>
        <el-form-item label="主机地址">
          <el-input v-model="form.host" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="认证方式">
          <el-radio-group v-model="form.auth_type">
            <el-radio value="password">密码</el-radio>
            <el-radio value="key">密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.auth_type === 'password'" label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item v-if="form.auth_type === 'key'" label="私钥">
          <el-input v-model="form.ssh_key" type="textarea" :rows="6" placeholder="粘贴 SSH 私钥内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="submit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { ElMessage } from "element-plus"
import { getNodes, createNode, deleteNode, updateNode, testNode } from "../api"

const nodes = ref<any[]>([])
const loading = ref(true)
const showDialog = ref(false)
const testingId = ref(0)
const editingId = ref(0)

const defaultForm = { name: "", node_type: "ssh", host: "", port: 22, username: "", auth_type: "password", password: "", ssh_key: "" }
const form = ref({ ...defaultForm })

function formatTime(t: string) {
  if (!t) return ""
  const d = new Date(t)
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function openAdd() {
  editingId.value = 0
  form.value = { ...defaultForm }
  showDialog.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  form.value = {
    name: row.name,
    node_type: row.node_type,
    host: row.host,
    port: row.port,
    username: row.username,
    auth_type: row.auth_type,
    password: "",
    ssh_key: "",
  }
  showDialog.value = true
}

async function load() {
  loading.value = true
  try {
    const { data } = await getNodes()
    nodes.value = data
  } catch {} finally {
    loading.value = false
  }
}

async function submit() {
  try {
    if (editingId.value) {
      await updateNode(editingId.value, form.value)
      ElMessage.success("节点修改成功")
    } else {
      await createNode(form.value)
      ElMessage.success("节点添加成功")
    }
    showDialog.value = false
    form.value = { ...defaultForm }
    editingId.value = 0
    load()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "操作失败")
  }
}

async function remove(id: number) {
  await deleteNode(id)
  ElMessage.success("节点已删除")
  load()
}

async function testConn(id: number) {
  testingId.value = id
  try {
    await testNode(id)
    ElMessage.success("连接成功")
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "连接失败")
  } finally {
    testingId.value = 0
  }
}

onMounted(load)
</script>

<style scoped>
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.header h2 { margin: 0; font-size: 20px; font-weight: 600; }
</style>
