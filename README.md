# SSL Manager

SSL 证书全生命周期管理 Web 应用。通过浏览器可视化页面完成域名证书的申请、自动续签、多平台部署与用户管理。

## 功能

- **域名管理** — 添加、删除域名，配置部署节点与自动部署/续签开关
- **证书管理** — 独立证书管理模块，列表查看、详情、下载、部署、删除
- **ACME 申请** — 对接 Let's Encrypt CA，DNS-01 手动验证签发证书
- **多域名/通配符** — 支持 SAN 多域名和通配符证书（如 `*.example.com`）
- **智能续签** — 定时检测证书有效期，到期前自动续签
- **SSH 部署** — 通过 SSH/SCP 将证书部署到远程 Nginx 服务器，支持密码和密钥认证
- **部署日志** — 每次部署记录详细日志，包含上传路径与 nginx reload 输出
- **多平台支持** — Windows / Linux
- **用户与角色** — 注册用户和管理员角色，API 权限隔离
- **系统配置** — ACME 目录、JWT 密钥等配置持久化到数据库，通过 API 管理
- **数据库迁移** — 内置迁移引擎，数据结构变更可追溯

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25 + Gin |
| ACME 客户端 | go-acme/lego v4 |
| 数据库 | SQLite（GORM） |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 认证 | JWT（golang-jwt v5） |
| SSH | golang.org/x/crypto |

## 目录结构

```
.
├── cmd/server/                 # 服务入口
│   └── main.go                 #  启动流程：配置 → DB → 迁移 → 路由
├── internal/
│   ├── acme/                   # ACME 协议（lego 封装）
│   │   ├── app.go              #  申请状态跟踪
│   │   ├── client.go           #  证书申请
│   │   └── manual.go           #  DNS-01 手动验证处理器
│   ├── auth/                   # JWT 签发与解析
│   │   └── jwt.go
│   ├── config/                 # 配置加载
│   │   └── config.go
│   ├── deploy/                 # 证书部署
│   │   ├── deployer.go         #  部署调度器
│   │   └── ssh.go              #  SSH 连接、文件上传、nginx reload
│   ├── handler/                # HTTP 接口
│   │   ├── auth.go             #  注册 / 登录 / 当前用户
│   │   ├── cert.go             #  证书申请 / DNS 验证 / 状态 / 下载
│   │   ├── certificate.go      #  证书管理 CRUD（独立模块）
│   │   ├── deploy.go           #  部署触发 / 部署日志 / 平台信息
│   │   ├── domain.go           #  域名 CRUD
│   │   ├── node.go             #  部署节点管理
│   │   ├── system_config.go    #  系统配置与迁移历史
│   │   └── user.go             #  用户管理（admin）
│   ├── middleware/             # Gin 中间件
│   │   └── auth.go             #  JWT 鉴权 + 管理员守卫
│   ├── platform/               # 平台检测
│   │   ├── platform.go
│   │   ├── linux.go
│   │   └── windows.go
│   ├── response/               # 统一响应格式
│   │   └── response.go
│   └── store/                  # 数据持久化
│       ├── db.go               #  GORM 初始化 + 模型定义
│       ├── migrate.go          #  数据库迁移引擎
│       ├── cert.go             #  证书表操作
│       ├── deploy.go           #  部署日志表操作
│       ├── domain.go           #  域名表操作
│       ├── node.go             #  部署节点表操作
│       ├── system_config.go    #  系统配置表操作
│       └── user.go             #  用户表操作 + 种子管理员
├── web/                        # 前端项目
│   └── src/
│       ├── api/                #  Axios 接口封装
│       ├── router/             #  Vue Router
│       ├── stores/             #  Pinia 状态管理（认证）
│       └── views/              #  页面组件
│           ├── Dashboard.vue   #  仪表盘
│           ├── Login.vue       #  登录
│           ├── Domains.vue     #  域名列表
│           ├── AddDomain.vue   #  添加域名
│           ├── DomainDetail.vue#  域名详情
│           ├── Certs.vue       #  证书列表
│           ├── CertDetail.vue  #  证书详情
│           ├── CertApply.vue   #  证书申请
│           ├── CertDownload.vue#  证书下载
│           ├── Deploy.vue      #  部署节点管理
│           └── Users.vue       #  用户管理
├── config.yaml                 # 启动必需配置
├── Makefile                    # 构建/运行命令
├── go.mod / go.sum             # Go 依赖
└── README.md
```

## 快速开始

### 前置条件

- Go 1.21+
- Node.js 22+（仅前端开发时需要）

### 启动（后端 + 内嵌前端）

```bash
# 克隆并进入项目目录
git clone https://github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager.git
cd kitakami_hibiki_ssl_manager

# 编译前端（首次或更新前端后）
cd web && npm ci && npm run build && cd ..

# 启动服务（默认 8080 端口）
go run ./cmd/server
```

浏览器访问 http://localhost:8080

### Makefile 快捷命令

```bash
make build          # 编译后端到 bin/server
make build-frontend # 编译前端
make run            # 编译前端 + 启动服务
make dev            # go run 直接启动
make frontend       # 单独启动前端开发服务器
make clean          # 清理 bin/ web/dist/
```

### 默认管理员账户

首次启动自动创建：

| 用户名 | 密码 | 角色 |
|---|---|---|
admin | admin123 | admin

## 配置

### config.yaml（启动必需）

```yaml
server:
  port: 8080
acme:
  directory: https://acme-v02.api.letsencrypt.org/directory
storage:
  driver: sqlite
  dsn: ./data/certs.db
auth:
  jwt_secret: ""
  deploy_key: <hex-key>
```

`auth.jwt_secret` 留空则自动生成随机密钥。`auth.deploy_key` 用于外部 API 调用鉴权。

## API 参考

所有 API 前缀 `/api`。认证接口除外，其余需 `Authorization: Bearer <token>`。

### 认证

| 方法 | 路径 | 说明 |
|---|---|---|
POST | /api/auth/register | 注册新用户 |
POST | /api/auth/login | 登录，返回 JWT |
GET | /api/auth/me | 当前登录用户信息 |

### 域名

| 方法 | 路径 | 说明 |
|---|---|---|
GET | /api/domains | 域名列表 |
POST | /api/domains | 添加域名 |
DELETE | /api/domains | 删除域名（?id=） |
PUT | /api/domains | 更新域名部署配置 |
GET | /api/domains/detail | 域名详情（?id=） |

### 证书申请

| 方法 | 路径 | 说明 |
|---|---|---|
POST | /api/certs/apply | 申请新证书（支持 extra_domains） |
GET | /api/certs/challenge-value | 获取 DNS 挑战值（?domain=） |
POST | /api/certs/verify-dns | 验证 DNS TXT 记录 |
GET | /api/certs/status | 申请状态（?domain=） |
GET | /api/certs | 证书列表（分页，?domain_id=） |
GET | /api/certs/detail | 证书详情（?id=） |
GET | /api/certs/download | 下载证书文件（?id=&type=fullchain\|privkey） |

### 证书管理（独立模块）

| 方法 | 路径 | 说明 |
|---|---|---|
GET | /api/certificates | 证书列表（分页，?domain_id=） |
GET | /api/certificates/detail | 证书详情（?id=） |
DELETE | /api/certificates | 删除证书（?id=） |

### 部署

| 方法 | 路径 | 说明 |
|---|---|---|
POST | /api/certs/deploy | 部署证书（cert_id） |
GET | /api/certs/deploy-logs | 部署日志（?cert_id=\|domain_id=） |
GET | /api/platform | 当前操作系统信息 |

### 部署节点

| 方法 | 路径 | 说明 |
|---|---|---|
GET | /api/nodes | 节点列表 |
POST | /api/nodes | 添加节点 |
PUT | /api/nodes | 修改节点 |
DELETE | /api/nodes | 删除节点（?id=） |
GET | /api/nodes/test | 测试连接（?id=） |

### 用户管理（admin）

| 方法 | 路径 | 说明 |
|---|---|---|
GET | /api/users | 用户列表 |
PUT | /api/users | 修改用户角色 |
DELETE | /api/users | 删除用户 |

### 系统配置（admin）

| 方法 | 路径 | 说明 |
|---|---|---|
GET | /api/system/config | 获取运行时配置 |
PUT | /api/system/config | 更新运行时配置 |
GET | /api/system/migrations | 查看迁移记录 |

## 数据存储

### 文件

```
data/
  certs.db            # SQLite 数据库
certs/
  <domain>/
    fullchain.pem     # 证书链
    privkey.pem       # 私钥（权限 0600）
```

### 数据库表

| 表 | 说明 |
|---|---|
users | 用户（邮箱、用户名、bcrypt 密码、角色）
domains | 域名（关联用户、ACME 邮箱、部署配置、自动续签开关）
certificates | 证书记录（关联域名、SAN 域名列表、状态、签发/到期时间）
deploy_nodes | 部署节点（SSH 连接信息、认证方式）
deploy_logs | 部署日志（证书ID、节点、状态、详情、起止时间）
system_configs | 运行时配置（单行）
schema_migrations | 迁移记录（版本号、描述、执行时间）

## 数据库迁移

项目内置迁移引擎。迁移定义在 `internal/store/migrate.go` 的 `allMigrations()` 中：

```go
func allMigrations() []Migration {
    return []Migration{
        {Version: "2026-07-01-001", Description: "..."},
        {Version: "2026-07-01-002", Description: "..."},
        // 追加新迁移在这里
    }
}
```

新增迁移后，启动时自动执行未应用的步骤，每步在事务中运行。失败时事务回滚，服务拒绝启动。

## 开发

### 后端

```bash
go run ./cmd/server
```

### 前端

```bash
cd web
npm run dev     # Vite 开发服务器
npm run build   # 编译到 web/dist/
```

## 许可证

[Apache License 2.0](LICENSE)
