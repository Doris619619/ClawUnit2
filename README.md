# ClawUnit

ClawUnit 是 CUHKSZ AI 平台的 OpenClaw 实例管理服务。它负责在 Kubernetes 上创建、调度和管理 OpenClaw AI 助手实例，并提供聊天代理，让用户通过统一的前端界面与 OpenClaw 实例对话。

## 在平台中的位置

```
                  ┌───────────┐
                  │  用户浏览器  │
                  └─────┬─────┘
                        │
                  ┌─────▼─────┐
                  │  UI (前端)  │  React, /chat/clawunit 路径
                  └─────┬─────┘
                        │ X-Service: clawunit
                  ┌─────▼─────┐
                  │  Gateway   │  Go 微服务 / Vite Proxy
                  └─────┬─────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
   ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼──────┐
   │  UniAuth   │ │Open-Platform│ │  ClawUnit   │
   │  (鉴权)    │ │  (LLM代理) │ │ (本项目)     │
   │  :8004    │ │  :8032     │ │  :8282      │
   └───────────┘ └────────────┘ └─────┬──────┘
                                      │
                                ┌─────▼─────┐
                                │ Kubernetes │
                                │  Cluster   │
                                └─────┬─────┘
                                      │
                        ┌─────────────┼─────────────┐
                        │             │             │
                   ┌────▼───┐  ┌────▼───┐  ┌────▼───┐
                   │ Pod A  │  │ Pod B  │  │ Pod C  │
                   │OpenClaw│  │OpenClaw│  │OpenClaw│
                   │:18789  │  │:18789  │  │:18789  │
                   └────────┘  └────────┘  └────────┘
```

## 核心功能

- **实例生命周期管理**：创建、启动、停止、重启、删除 OpenClaw 实例（error 状态可直接重启）
- **WebSocket 实时聊天**：通过 WS 双向桥接连接 OpenClaw Pod，支持流式文本、工具调用事件、exec 审批
- **Session 管理**：每个连接独立 session，支持 `sessions.list`/`chat.history`/`sessions.delete`/`sessions.patch`
- **RPC 安全过滤**：WS 桥接层白名单过滤，仅允许聊天和 session 操作，阻止管理类 RPC（config.set 等）
- **Exec 审批（待 OpenClaw 修复）**：`exec.approval.requested` 事件推送到前端，用户可允许/拒绝命令执行
- **双模式 API 配置**：手动（用户自带 API Key + Base URL）或自动（通过 Open Platform 分配）
- **配置热更新**：通过 UpdateConfig API 重新生成 ConfigMap，agent/model/tools 配置热加载
- **资源配额**：按用户限制实例数量（资源配额使用服务端固定配置）
- **审计日志**：记录所有管理操作和聊天请求
- **K8s 资源编排**：自动管理 Namespace、Pod、PVC、Service、NetworkPolicy、ConfigMap
- **开发环境支持**：out-of-cluster 模式自动 port-forward 到 Pod

## 快速开始

```bash
# 1. 创建数据库（假设 PG 已就绪）
psql "postgresql://user:pass@host:port/postgres" -c "CREATE DATABASE clawunit;"
psql "postgresql://user:pass@host:port/clawunit" -f manifest/deploy/migrations/001_init_schema.sql

# 2. 编辑配置
vim manifest/config/config.yaml

# 3. 安装依赖
go mod tidy

# 4. 启动
go run main.go
```

## 配置说明

`manifest/config/config.yaml` 关键配置项：

```yaml
k8s:
  mode: "outofcluster"        # auto | incluster | outofcluster
  namespace: "gpt-clawunit"   # 基础命名空间
  storageClass: "longhorn"    # 默认 StorageClass

instance:
  defaultImage: "ghcr.io/openclaw/openclaw:latest"  # 或 :slim (gateway-only)
  defaultCPU: "2"             # K8s 资源格式
  defaultMemory: "4Gi"
  defaultDisk: "500Mi"
  containerPort: 18789        # OpenClaw Gateway 端口
```

## 文档

| 文档 | 内容 |
|------|------|
| [OpenClaw 集成路线图](docs/openclaw-integration-roadmap.md) | 已完成/待实现的 OpenClaw 功能清单 |
| [CLAUDE.md](CLAUDE.md) | 代码规范、架构说明、开发指引 |

## 技术栈

- **语言**: Go 1.26 + GoFrame v2
- **数据库**: PostgreSQL
- **容器编排**: Kubernetes (client-go, 含 port-forward 支持)
- **OpenClaw**: ghcr.io/openclaw/openclaw 镜像，端口 18789
- **OpenClaw 通信**: WebSocket 双向桥接（主）+ OpenAI 兼容 HTTP SSE（回退）
- **前端**: React (在 [UI](https://github.com/CUHKSZ-ITSO-Dev/UI) 仓库 `feat/clawunit-integration` 分支)
