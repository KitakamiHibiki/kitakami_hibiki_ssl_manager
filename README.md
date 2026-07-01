# SSL Manager

SSL 证书全生命周期管理 Web 应用。通过浏览器可视化页面完成域名证书的申请、自动续签、多平台部署与用户管理。

## 功能

- **证书管理** — 一站式管理所有域名证书，查看签发状态和到期时间
- **ACME 自动申请** — 对接 Let"s Encrypt CA，HTTP 验证自动签发证书
- **智能续签** — 定时检测证书有效期，到期前自动续签
- **一键部署** — 将证书部署到 Nginx 或本地目录
- **多平台支持** — Windows / Linux 自动识别 Nginx 路径与重载命令
- **用户与角色** — 注册用户和管理员角色，API 权限隔离
- **系统配置** — ACME 目录、JWT 密钥、通知邮箱等配置持久化到数据库，通过 API 管理
- **数据库迁移** — 内置迁移引擎，数据结构变更可追溯、可回滚

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25 + Gin |
| ACME 客户端 | go-acme/lego v4 |
| 数据库 | SQLite（GORM） |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 认证 | JWT（golang-jwt v5） |
| 调度 | robfig/cron v3 |

## 目录结构

```
.
├── cmd/server/                 # 服务入口
│   └── main.go                 #  启动流程：配置 → DB → 迁移 → 路由
├── internal/
│   ├── acme/                   # ACME 协议（lego 封装）
│   │   ├── client.go           #  证书申请与续签
│   │   └── provider.go         #  HTTP-01 验证处理器
│   ├── auth/                   # JWT 签发与解析
│   │   └── jwt.go
│   ├── config/                 # 配置加载（yaml + DB 合并）
│   │   └── config.go
│   ├── deploy/                 # 证书部署
│   │   ├── deploy.go           #  部署器接口
│   │   ├── local.go            #  本地文件部署
│   │   └── nginx.go            #  Nginx 部署（含重载）
│   ├── handler/                # HTTP 接口
│   │   ├── auth.go             #  注册 / 登录 / 当前用户
│   │   ├── cert.go             #  证书申请 / 续签 / 列表 / 下载
│   │   ├── deploy.go           #  部署触发
│   │   ├── domain.go           #  域名 CRUD
│   │   ├── system_config.go    #  系统配置与迁移历史
│   │   └── user.go             #  用户管理（admin）
│   ├── middleware/             # Gin 中间件
│   │   └── auth.go             #  JWT 鉴权 + 管理员守卫
│   ├── platform/               # 平台检测
│   │   ├── platform.go         #  接口定义
│   │   ├── linux.go            #  Linux 实现
│   │   └── windows.go          #  Windows 实现
│   ├── scheduler/              # 定时续签调度
│   │   └── scheduler.go
│   └── store/                  # 数据持久化
│       ├── db.go               #  GORM 初始化 + AfterInit
│       ├── migrate.go          #  数据库迁移引擎
│       ├── domain.go           #  域名表操作
│       ├── cert.go             #  证书表操作
│       ├── user.go             #  用户表操作 + 种子管理员
│       └── system_config.go    #  系统配置表操作
├── web/                        # 前端项目
│   └── src/
│       ├── api/                #  Axios 接口封装
│       ├── router/             #  Vue Router（登录 / 仪表盘 / 域名 / 证书 / 部署 / 用户）
│       ├── stores/             #  Pinia 状态管理（认证）
│       ├── views/              #  页面组件
│       └── components/         #  可复用组件
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

storage:
  driver: sqlite
  dsn: ./data/certs.db
```

`storage.dsn` 和 `server.port` 必须在启动前确定。其余配置（ACME 目录、JWT 密钥、调度参数、通知等）存储在数据库 `system_configs` 表中，可通过管理界面或 API 修改。

### 运行时配置（DB）

| 字段 | 默认值 | 说明 |
|---|---|---|
acme_directory | https://acme-v02.api.letsencrypt.org/directory | ACME CA 目录 URL |
check_interval | 0 3 * * * | 续签 cron 表达式（UTC） |
renew_before_days | 30 | 到期前多少天触发续签 |
notify_email | "" | 通知邮箱 |
notify_webhook | "" | 通知 Webhook URL |
jwt_secret | change-me-in-production | JWT 签名密钥 |

> 修改 JWT secret 后，用旧 secret 签发的 token 将在下次重启后失效。修改 ACME directory 或调度参数同样需要重启。

## API 参考

所有 API 前缀 `/api`。认证接口除外，其余需 `Authorization: Bearer <token>`。

### 认证

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/auth/register | 注册新用户（邮箱 + 密码） |
| POST | /api/auth/login | 登录，返回 JWT |
| GET | /api/auth/me | 当前登录用户信息 |

### 域名

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/domains | 域名列表 |
| POST | /api/domains | 添加域名 |
| DELETE | /api/domains | 删除域名（?id=） |

### 证书

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/certs/apply | 申请新证书 |
| POST | /api/certs/renew | 续签指定证书 |
| GET | /api/certs | 证书列表 |
| GET | /api/certs/detail | 证书详情（?id=） |
| GET | /api/certs/download | 下载 fullchain.pem（?id=） |

### 部署

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/deploy | 部署证书到目标 |
| GET | /api/platform | 当前操作系统信息 |

### 用户管理（admin）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/users | 用户列表 |
| PUT | /api/users | 修改用户角色 |
| DELETE | /api/users | 删除用户 |

### 系统配置（admin）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/system/config | 获取运行时配置 |
| PUT | /api/system/config | 更新运行时配置 |
| GET | /api/system/migrations | 查看迁移记录 |

## 数据存储

### 文件

```
data/
  certs.db            # SQLite 数据库（用户、域名、证书、系统配置、迁移记录）
certs/
  <domain>/
    fullchain.pem     # 证书链
    privkey.pem       # 私钥（权限 0600）
    issuer.pem        # CA 中间证书（可选）
```

### 数据库表

| 表 | 说明 |
|---|---|
users | 用户（邮箱、用户名、bcrypt 密码、角色）
domains | 域名（关联用户、ACME 邮箱、验证方式）
certificates | 证书记录（关联域名、状态、签发/到期时间）
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
# 热重载启动（不编译前端）
go run ./cmd/server
```

### 前端

```bash
cd web
npm run dev     # Vite 开发服务器
npm run build   # 编译到 web/dist/
npm run preview # 预览编译结果
```

## 平台差异

Nginx 部署路径与重载命令：

| 项目 | Linux | Windows |
|---|---|---|
| 配置目录 | /etc/nginx/conf.d/ | C:\nginx\conf\conf.d\ |
| 重载命令 | systemctl reload nginx | nginx -s reload |

## 许可证

[Apache License 2.0](LICENSE)
