# kitakami_hibiki_ssl_manager

SSL 证书全生命周期管理 Web 应用 —— 通过浏览器可视化页面完成证书的申请、自动续签与多平台部署。

## 功能

- **可视化证书管理** — 浏览器界面一站式管理所有域名证书，申请、续签、部署全流程可视化
- **ACME 自动申请** — 对接 Let's Encrypt 等 CA，通过 HTTP 或 DNS 验证自动签发证书
- **智能续签** — 定时检测证书有效期，到期前自动续签，支持邮件/Webhook 通知
- **一键部署** — 将证书部署到 Nginx、Apache、CDN、云平台等多种目标
- **多域名支持** — 批量管理多个域名，支持泛域名证书
- **操作审计** — 记录所有证书操作日志，便于追溯

## 技术栈

- **后端** — Go
- **前端** — Vue 3 + TypeScript + Element Plus
- **数据库** — SQLite / PostgreSQL

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/kitakami_hibiki/kitakami_hibiki_ssl_manager.git
cd kitakami_hibiki_ssl_manager

# 启动服务
go run ./cmd/server

# 浏览器访问
# http://localhost:8080
```

## 界面截图

> TODO

## 使用流程

1. **添加域名** — 填写域名、选择验证方式（HTTP / DNS）
2. **申请证书** — 系统自动完成 ACME 验证并签发证书
3. **配置部署** — 选择部署目标（Nginx、CDN 等），一键推送证书
4. **自动续签** — 系统后台定时检查，到期前自动续签并重新部署

## 目录结构

```
.
├── cmd/server/            # 服务入口
├── internal/
│   ├── handler/           # HTTP 接口
│   ├── acme/              # ACME 协议
│   ├── deploy/            # 部署模块
│   ├── scheduler/         # 定时续签
│   └── store/             # 数据持久化
├── web/                   # 前端代码
├── config.yaml            # 配置文件
├── data/                  # 运行时生成 — 数据库
├── certs/                 # 运行时生成 — 证书文件
└── Makefile
```

所有运行时数据（数据库、证书）均存储在可执行文件所在的工作目录下。

## 平台差异

仅 Nginx 部署涉及平台差异：

| 项目 | Linux | Windows |
|---|---|---|
| Nginx 配置路径 | `/etc/nginx/conf.d/` | `C:\nginx\conf\conf.d\` |
| Nginx 重载命令 | `systemctl reload nginx` | `nginx -s reload` |

## 许可证

[Apache License 2.0](LICENSE)
