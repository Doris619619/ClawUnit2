# 鉴权机制

## 概述

ClawUnit 本身不管理用户账号，完全依赖 UniAuth 进行身份验证和权限控制。整个鉴权链路如下：

```
用户浏览器
  → UI 登录（UniAuth 颁发 session）
  → 请求带上 cookie/token
  → Gateway 微服务（根据 X-Service 头路由）
      → 从 session 中提取用户身份
      → 设置 X-User-ID 请求头
  → ClawUnit 读取 X-User-ID
      → 调用 UniAuth 验证权限
      → 通过后设置 ctx("ownerUpn")
```

## 中间件详解

### InjectIdentity（所有 API 路由）

位置：`internal/middlewares/auth.go`

1. 从请求头提取 `X-User-ID`（由 Gateway 注入，值为用户的 UPN，如 `alice@cuhk.edu.cn`）
2. 调用 UniAuth `POST /auth/check`，验证 `(upn, "clawunit", "access")`
3. 验证通过 → 将 `ownerUpn` 存入请求上下文
4. 验证失败 → 返回 401

```
请求头示例：
X-User-ID: alice@cuhk.edu.cn
```

### RequireAdmin（管理员路由）

位置：`internal/middlewares/auth.go`

在 InjectIdentity 之后执行，额外验证 `(upn, "clawunit", "admin")`。

### UniAuth 权限模型

ClawUnit 在 UniAuth 的 Casbin RBAC 中注册两个权限：

| Subject (用户) | Object | Action | 用途 |
|---------------|--------|--------|------|
| alice@cuhk.edu.cn | clawunit | access | 普通用户访问 ClawUnit |
| admin@cuhk.edu.cn | clawunit | admin | 管理员操作（全局实例列表、配额管理） |

**在 UniAuth 中配置步骤：**
1. 在 UniAuth 管理后台添加 `clawunit` 作为 object
2. 为需要使用 ClawUnit 的用户添加 `(user, clawunit, access)` 策略
3. 为管理员添加 `(admin, clawunit, admin)` 策略

## 实例归属控制

每个 API 请求都通过 `ownerUpn` 确保用户只能操作自己的实例：

```go
// 所有用户端查询都带 owner_upn 过滤
g.DB().Model("instances").Where("owner_upn", ownerUpn).Where("id", req.Id)
```

管理员端点（`/api/admin/v1`）不加 `owner_upn` 过滤，可查看所有用户的实例。

## API 模式与鉴权依赖

ClawUnit 支持两种 API 配置模式，影响 API Key 的来源和 UniAuth 的依赖程度：

### auto 模式（依赖 UniAuth）

通过 UniAuth/Open-Platform 自动分配 API Key。所有 LLM 调用经过 Open-Platform 的计费和审计。适用于组织内统一管理的场景。

### manual 模式（无 UniAuth API Key 依赖）

用户在创建实例时自行提供 `apiKey` 和 `baseUrl`（如 OpenRouter、OpenAI 等）。ClawUnit 将这些值存入 DB，每次启动 Pod 时注入环境变量。该模式下：

- **不调用** UniAuth 的 API Key 管理接口
- 用户鉴权（`InjectIdentity`）仍然需要 UniAuth
- 用户自行承担 key 的有效性、计费和安全

### 实例 Gateway Token

每个实例（不论 API 模式）在创建时生成一个随机 `gateway_token`（64 字符 hex），用于 OpenClaw Gateway 的 token 认证：

```
OpenClaw Gateway 配置（openclaw.json）：
  auth.mode = "token"
  auth.token = "${OPENCLAW_GATEWAY_TOKEN}"   ← 由环境变量注入

ClawUnit 聊天代理转发时：
  Authorization: Bearer {gateway_token}      ← 从 DB 读取
```

这确保只有 ClawUnit 代理能访问 Pod 的 Gateway API，而非集群内任意服务。

### NetworkPolicy 出口控制

每个实例的 NetworkPolicy 控制 Pod 的出口流量：

| 目标 | 协议/端口 | 用途 |
|------|----------|------|
| kube-system 命名空间 | UDP/TCP 53 | DNS 解析 |
| 系统命名空间 | TCP 80, 443, 8032 | Open-Platform 等内部服务 |
| 任意地址 | TCP 80, 443 | 外网 HTTP/HTTPS（LLM API、apt 仓库、web_fetch、浏览器工具等） |

manual 模式下 Pod 需要访问外部 LLM API（如 `api.openai.com`），浏览器工具和 apt 安装也需要外网访问，因此 NetworkPolicy 放行了出口 80/443 端口。

## API Key 生命周期（auto 模式）

auto 模式下，ClawUnit 为每个 OpenClaw 实例在 UniAuth/Open-Platform 中创建专属 API Key，用于实例内部调用 LLM 服务。

### 创建流程

```
实例创建时：
1. 调用 UniAuth POST /openPlatform/apikey
   - nickname: "clawunit-instance-{id}"
   - quotaPool: "clawunit"（可配置）
2. UniAuth 返回 { rawApiKey: "sk-xxx", apiKeyHash: "sha256..." }
3. rawApiKey 注入 Pod 环境变量：OPENAI_API_KEY=sk-xxx
4. apiKeyHash 存入 DB（不存明文）
5. 记录存入 api_key_provisions 表
```

### 启动流程（已停止的实例）

```
实例启动时（明文 key 已丢失）：
1. 查找现有未吊销的 key
2. 吊销旧 key（调 UniAuth DELETE /openPlatform/apikey）
3. 重新创建新 key
4. 新 rawApiKey 注入 Pod
```

### 删除流程

```
实例删除时：
1. 查找该实例所有未吊销的 key
2. 逐一吊销
3. 更新 api_key_provisions.revoked_at
```

### Pod 内环境变量

| 变量 | 值 | 来源 | 用途 |
|------|---|------|------|
| `HOME` | `/home/node` | 固定 | Node.js 进程的 home 目录 |
| `OPENCLAW_CONFIG_DIR` | `/home/node/.openclaw` | 固定 | OpenClaw 配置目录（由 init container 写入） |
| `OPENCLAW_GATEWAY_TOKEN` | `<hex>` | 创建时随机生成 | Gateway 认证 token |
| `CUSTOM_API_KEY` | `sk-xxx` | auto: UniAuth; manual: 用户提供 | OpenClaw 内调用 LLM 的认证凭据 |
| `CUSTOM_BASE_URL` | `http://...` | auto: Open-Platform; manual: 用户提供 | LLM API 的 Base URL |
| `INSTANCE_ID` | `123` | DB | 实例标识 |
| `OWNER_UPN` | `alice@cuhk.edu.cn` | 请求上下文 | 用户标识 |
| `PLAYWRIGHT_BROWSERS_PATH` | `/home/node/.cache/ms-playwright` | 固定 | Playwright 浏览器二进制搜索路径（指向共享 PVC 挂载点） |

auto 模式下 `CUSTOM_BASE_URL` 指向 Open-Platform（如 `http://localhost:8032/open/v1`），所有 LLM 调用经过 Open-Platform 的计费和审计。manual 模式下指向用户提供的外部 API 地址。

## 聊天代理鉴权

聊天请求（`/api/chat/v1/completions`）的鉴权：

1. **用户鉴权**：InjectIdentity 中间件验证 `X-User-ID`
2. **实例归属验证**：proxy 层查 DB 确认 `instance.owner_upn == ownerUpn`
3. **实例状态检查**：确认 `status == "running"` 且 `pod_ip` 非空
4. **Pod 通信**：使用 `instance.gatewayToken` 作为 Bearer token 与 Pod 通信（若配置了的话）

```
用户 → [X-User-ID 鉴权] → [实例归属检查] → [状态检查] → Pod:18789
```

## WebSocket 鉴权

WebSocket 聊天端点（`GET /api/chat/v1/ws`）不走中间件链，在 handler 内部完成认证：

1. **用户身份**：从 `X-User-ID` header 或 `userId` query param 获取 UPN
2. **实例归属**：查 DB 验证 `instance.owner_upn == upn` 且 `status == running`
3. **Pod 连接**：使用 `gateway_token` 作为 OpenClaw 认证 token
4. **WS 握手**：以 `client.id = "openclaw-control-ui"` 身份连接 Pod，请求 `operator.read/write/admin/approvals/pairing` 权限
5. **RPC 安全过滤**：前端发来的 RPC 方法经白名单过滤，只允许 `chat.*`、`sessions.*`、`exec.approval.resolve`、`plugin.approval.resolve`，管理类 RPC 被拦截

这种设计确保用户只能通过 WS 进行聊天和会话操作，无法绕过 ClawUnit 修改实例配置。

## Exec 审批鉴权

OpenClaw 的命令执行审批流程需要 `operator.approvals` scope：

1. Gateway 广播 `exec.approval.requested` 事件到 WS 连接
2. 前端展示审批横幅（允许/拒绝）
3. 用户决策通过 `exec.approval.resolve` RPC 回传（该方法在 WS 白名单中）

**当前状态：** 由于 OpenClaw 2026.4.2 版本的 loopback pairing bug，审批功能暂不可用。当前 exec 使用 `security: "full"` 自动批准所有执行。

## 前端 Feature Gate

UI 侧通过 UniAuth 的 Feature Gate 控制 ClawUnit 入口是否显示：

- Feature ID: `clawunit`
- 在 `ApplicationPermissionContext` 中注册
- 侧边栏的 "OpenClaw" 入口只对有权限的用户可见
