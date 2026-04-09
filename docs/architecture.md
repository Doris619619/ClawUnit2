# 架构设计

## 整体架构

ClawUnit 是一个中间层服务，衔接前端 UI 与 Kubernetes 集群中的 OpenClaw 实例。它不直接提供 AI 能力，而是：
1. 管理 OpenClaw Pod 的生命周期
2. 代理用户聊天请求到对应 Pod
3. 通过 UniAuth 实施鉴权和配额控制
4. 通过 Open-Platform 为实例分配 LLM API Key

## 请求流

### 聊天请求（WebSocket 核心路径）

```
用户在 UI 输入消息
  → UI 建立 WebSocket 连接 ws://{host}/api/chat/v1/ws?instanceId=123
  → Vite Proxy / Gateway 转发到 ClawUnit :8282
  → proxy.WsChat（无中间件，handler 内部认证）：
      1. 从 X-User-ID header 或 userId query 获取身份
      2. 从 DB 查实例，验证归属 + 状态为 running
      3. buildPodWsURL：
         - out-of-cluster → 自动 port-forward 到本地随机端口
         - in-cluster → 直连 Pod IP
      4. dialPod：连接 Pod WS，发送 connect frame（control-ui 身份），
         跳过 connect.challenge，等待 hello-ok
      5. 建立双向桥接：前端 ↔ Pod 透传帧
      6. 前端→Pod 方向 RPC 白名单过滤（只允许 chat/sessions/approval 方法）
  → 前端发送 chat.send → Pod 执行 Agent → 流式 agent 事件回传
  → 前端按 session 过滤事件，渲染文本/工具调用/审批横幅
```

### 实例管理请求

```
UI 发送 POST /api/instances/v1/create
  → Gateway → ClawUnit
  → UniResMiddleware（统一响应格式）
  → InjectIdentity（鉴权）
  → Controller → lifecycle.Create：
      1. 配额检查（user_quotas 表）
      2. 实例名唯一性检查
      3. 生成实例专属 gateway_token（随机 32 字节 hex）
      4. 插入 instances 记录（status=creating, api_mode, gateway_token 等）
      5. 构建环境变量：
         - auto 模式 → 调 UniAuth 创建 API Key，注入 CUSTOM_API_KEY + CUSTOM_BASE_URL
         - manual 模式 → 直接使用用户提供的 apiKey/baseUrl
         - 所有模式 → 注入 OPENCLAW_GATEWAY_TOKEN
      6. 生成 openclaw.json 配置（含 gateway、模型、provider 信息）
      7. 确保用户 Namespace 存在
      8. 创建 ConfigMap（存放 openclaw.json）
      9. 确保用户 PVC 存在
      10. 创建 NetworkPolicy
      11. 创建 Pod（init container 拷贝配置 + gateway container 运行 OpenClaw）
      12. 更新 DB 中 pod_name/pod_namespace
      ※ 任意步骤失败 → 回滚已创建的资源（含 ConfigMap 清理）
  → SyncService 每 5 秒轮询，检测 Pod Ready 后更新 status=running
```

### 管理员请求

```
UI 发送 GET /api/admin/v1/instances/list
  → Gateway → ClawUnit
  → UniResMiddleware → InjectIdentity → RequireAdmin
  → 额外调 UniAuth 验证 (upn, "clawunit", "admin") 权限
  → Controller 执行查询
```

## 每个用户的隔离模型

每个用户拥有独立的 Kubernetes 资源空间：

```
用户 alice@cuhk.edu.cn
  │
  ├── Namespace: clawunit-user-a1b2c3d4    ← UPN 的 SHA256 前 8 位
  │     │
  │     ├── PVC: clawunit-user-a1b2c3d4     ← 用户数据卷（跨实例共享）
  │     │     └── /home/user/data/          ← 挂载到每个 Pod
  │     │
  │     ├── ConfigMap: clawunit-1-config          ← openclaw.json 配置
  │     │     └── openclaw.json: gateway 配置 + 模型 provider 配置
  │     │
  │     ├── Pod: clawunit-1-myinstance              ← 实例 1
  │     │     ├── Init Container: init-config（busybox）
  │     │     │     └── 拷贝 ConfigMap → /home/node/.openclaw/openclaw.json
  │     │     ├── Container: gateway（OpenClaw 镜像）
  │     │     │     └── Command: sh -c "npx playwright install-deps chromium; exec runuser -u node -- node /app/dist/index.js gateway run"
  │     │     ├── Volume: openclaw-home (EmptyDir)   ← OpenClaw 工作目录
  │     │     ├── Volume: openclaw-config (ConfigMap) ← 配置文件
  │     │     ├── Volume: system-skills (ReadOnly)   ← 系统技能 PVC
  │     │     ├── Volume: playwright-browsers (ReadOnly) ← Playwright 浏览器 PVC
  │     │     ├── Volume: user-data (ReadWrite)      ← 用户数据 PVC
  │     │     ├── Volume: tmp-volume (EmptyDir)      ← /tmp
  │     │     ├── Port 18789: Gateway API（探针 + 聊天代理）
  │     │     └── Env:
  │     │           HOME=/home/node
  │     │           OPENCLAW_CONFIG_DIR=/home/node/.openclaw
  │     │           PLAYWRIGHT_BROWSERS_PATH=/home/node/.cache/ms-playwright
  │     │           OPENCLAW_GATEWAY_TOKEN=<hex>  ← 实例专属 gateway 认证 token
  │     │           CUSTOM_API_KEY=sk-xxx         ← auto 模式由 UniAuth 创建 / manual 模式用户提供
  │     │           CUSTOM_BASE_URL=http://...    ← auto 模式指向 Open-Platform / manual 模式用户提供
  │     │           INSTANCE_ID=1
  │     │           OWNER_UPN=alice@cuhk.edu.cn
  │     │
  │     └── NetworkPolicy: clawunit-1-myinstance-netpol
  │           └── 限制出口：只允许 DNS + 系统命名空间（80, 443, 8032）+ 外网 80/443
  │
  └── DB Records:
        ├── instances: id=1, status=running, pod_ip=10.x.x.x
        ├── user_quotas: max_instances=3, max_cpu=8, ...
        └── api_key_provisions: api_key_hash=sha256(sk-xxx)
```

**关键设计决策：**

| 决策 | 说明 |
|------|------|
| 不建 users 表 | 用 `owner_upn` (TEXT) 关联 UniAuth 身份，避免数据同步 |
| 每用户一个 PVC | 用户的多个实例共享同一个数据卷，停止/重启不丢数据 |
| UPN Hash 命名 | Namespace/PVC 用 UPN 的 SHA256[:8] 生成确定性短名，避免特殊字符 |
| NetworkPolicy 限制出口 | Pod 只能访问 DNS、系统命名空间和外网 80/443，防止滥用 |
| API Key 不存明文（auto 模式） | 只存 hash，启动时重新创建（吊销旧的 + 新建） |
| 双 API 模式 | auto 通过 Open-Platform 分配 key；manual 用户自带 key，无 UniAuth 依赖 |
| 每实例 ConfigMap | openclaw.json 经 ConfigMap 挂载，init container 拷贝到工作目录 |
| 每实例 gateway_token | 创建时随机生成，存入 DB，聊天代理用作 Bearer token 鉴权 |

## K8s 资源模型

### 资源命名规则

| 资源 | 命名格式 | 示例 |
|------|---------|------|
| Namespace | `{base}-user-{upn_hash8}` | `clawunit-user-a1b2c3d4` |
| Pod | `clawunit-{instanceID}-{name}` | `clawunit-1-myinstance` |
| NetworkPolicy | `clawunit-{instanceID}-{name}-netpol` | `clawunit-1-myinstance-netpol` |
| ConfigMap | `clawunit-{instanceID}-config` | `clawunit-1-config` |
| 用户 PVC | `clawunit-user-{upn_hash8}` | `clawunit-user-a1b2c3d4` |

所有名称经过 `sanitizeK8sName` 处理：小写、替换非法字符为 `-`、截断到 63 字符。

### OpenClaw Pod 结构

每个 Pod 包含两个阶段的容器：

**Init Container: `init-config`**
- 镜像：`busybox:1.37`
- 职责：将 ConfigMap 中的 `openclaw.json` 拷贝到 `/home/node/.openclaw/openclaw.json`，并创建 workspace 目录
- 资源限制：CPU 50m-100m, 内存 32Mi-64Mi
- 以 uid=1000 运行

**Main Container: `gateway`**
- 镜像：用户指定或默认 OpenClaw 镜像
- 启动命令：`sh -c "npx playwright install-deps chromium > /dev/null 2>&1; exec runuser -u node -- node /app/dist/index.js gateway run"`
- 启动时先以 root 安装 Chromium 系统依赖（apt），再 `runuser` 降权到 uid=1000 运行 Gateway
- 运行 OpenClaw Gateway，通过 WebSocket（端口 18789）和 HTTP `/v1/chat/completions` 提供 API
- 以 uid=1000 运行，非 root，不允许特权提升

### ConfigMap 机制（openclaw.json）

每个实例创建一个 ConfigMap（`clawunit-{instanceID}-config`），内容为 `openclaw.json`，配置了：

- **gateway**：绑定 LAN，端口 18789，token 认证模式（token 由环境变量 `${OPENCLAW_GATEWAY_TOKEN}` 注入），`controlUi.dangerouslyDisableDeviceAuth` 和 `allowInsecureAuth` 启用（WS 桥接需要）
- **agents.defaults.model**：使用 `custom/{modelID}` 作为默认模型
- **tools**：profile（默认 full）、exec security 为 full、fs.workspaceOnly 为 false
- **browser**：executablePath 指向 Playwright Chromium、headless 模式、noSandbox（容器非 root 用户）、ssrfPolicy 控制私有网络访问
- **plugins**：entries 配置（搜索引擎等扩展）
- **cron**：默认关闭（`enabled: false`）
- **models.providers.custom**：baseUrl 和 apiKey 通过环境变量 `${CUSTOM_BASE_URL}` / `${CUSTOM_API_KEY}` 注入

环境变量替换由 OpenClaw 自身在启动时完成（`${VAR}` 语法）。ConfigMap 在实例生命周期中仅创建一次（首次 create 时），启动/重启复用已有 ConfigMap。删除实例时清理。

### Pod 卷挂载

每个 Pod 挂载六个卷：

| 卷名 | 来源 | 挂载路径 | 读写模式 | 用途 |
|------|------|---------|---------|------|
| `openclaw-home` | EmptyDir | `/home/node/.openclaw` | ReadWrite | OpenClaw 运行时目录（init container 写入配置） |
| `openclaw-config` | ConfigMap | `/config` | ReadOnly | 挂载 openclaw.json 供 init container 读取 |
| `system-skills` | PVC | `/skills/system` | ReadOnly | 系统预置技能（管理员维护） |
| `playwright-browsers` | PVC | `/home/node/.cache/ms-playwright` | ReadOnly | Playwright Chromium 浏览器二进制（全局共享） |
| `user-data` | PVC | 配置的 mountPath | ReadWrite | 用户个人数据（跨实例共享） |
| `tmp-volume` | EmptyDir | `/tmp` | ReadWrite | 临时文件 |

### Pod 健康检查

使用 exec 探针（通过 node 发 HTTP 请求），而非 TCP 探针：

| 探针 | 端点 | 参数 |
|------|------|------|
| Startup | `GET /healthz` (exec) | 失败阈值 120，间隔 5s，超时 5s |
| Readiness | `GET /readyz` (exec) | 初始延迟 15s，间隔 10s，超时 5s |
| Liveness | `GET /healthz` (exec) | 初始延迟 60s，间隔 30s，超时 10s |

## 状态同步机制（SyncService）

后台 goroutine，每 5 秒执行一次：

```
1. 查询 DB 中 status=creating 或 running 的实例
2. 对每个实例，通过 K8s API 查询 Pod 状态
3. 状态映射：
   - Pod Running + Ready → status=running（同时更新 pod_ip）
   - Pod Pending → 保持当前状态
   - Pod Succeeded → status=stopped
   - Pod Failed → status=error
   - Pod 不存在 + 原 running → status=stopped
   - Pod 不存在 + 原 creating → status=error
4. 状态变更时更新 DB
```

## 双 API 模式（auto / manual）

实例创建时可选择 API 配置模式，决定 LLM API Key 的来源：

### auto 模式

```
创建实例 → 调 UniAuth 创建 API Key → 注入环境变量：
  CUSTOM_API_KEY = UniAuth 返回的 rawApiKey
  CUSTOM_BASE_URL = Open-Platform 地址 + /open/v1

启动实例 → 吊销旧 key + 创建新 key → 注入新环境变量
```

- 依赖 UniAuth 和 Open-Platform
- API Key 经过 Open-Platform 的计费和审计
- 明文 key 不持久化，仅在 Pod 环境变量中存在

### manual 模式

```
创建实例 → 用户提供 apiKey + baseUrl → 存入 DB → 注入环境变量：
  CUSTOM_API_KEY = 用户提供的 key
  CUSTOM_BASE_URL = 用户提供的 base URL

启动实例 → 从 DB 读取 apiKey/baseUrl → 直接注入
```

- 不依赖 UniAuth 的 API Key 管理
- 用户自行管理 key 的有效性和计费
- key 存储在 instances 表的 `api_key` 字段中

### 选择逻辑

控制器根据 `req.ApiMode` 字段（默认 `manual`）传递给 lifecycle.Create，lifecycle 内部根据模式决定是否调用 UniAuth。

## 实例 Gateway Token

每个实例在创建时生成一个随机的 `gateway_token`（32 字节 hex 编码，64 字符），用于：

1. **Pod 内**：作为 `OPENCLAW_GATEWAY_TOKEN` 环境变量，OpenClaw Gateway 使用它验证传入请求
2. **聊天代理**：ClawUnit proxy 转发请求时，在 `Authorization: Bearer {gateway_token}` header 中携带

这确保只有 ClawUnit 代理能访问 Pod 的 Gateway API，防止其他 Pod 或服务直接调用。

## 自动 Port-Forward（开发模式）

out-of-cluster 模式下（本地开发），ClawUnit 无法直连 Pod IP。WebSocket 聊天桥接通过自动 port-forward 解决：

```
proxy.buildPodWsURL:
  1. 调用 k8s.GetLocalPort(podName, podNamespace, port)
  2. 如果 in-cluster（Config.Host 为空）→ 返回 0，调用方直连 Pod IP
  3. 如果 out-of-cluster：
     a. 检查缓存中是否有活跃的 port-forward（TCP 探活）
     b. 有 → 复用本地端口
     c. 无 → 分配随机本地端口，建立 SPDY port-forward
     d. 返回本地端口，调用方连接 127.0.0.1:{localPort}
```

port-forward 连接缓存在内存 map 中（key: `namespace/podName`），实例停止/删除时通过 `ClosePortForward` 清理。

## WebSocket 聊天桥接（WsChat）

`proxy.WsChat`（`wschat.go`）提供基于 WebSocket 的双向聊天通道：

```
前端 WS ↔ ClawUnit ↔ OpenClaw Pod WS
```

### 连接流程

1. 前端连接 `GET /api/chat/v1/ws?instanceId=123&userId=alice@cuhk.edu.cn`
2. ClawUnit 验证用户身份和实例归属（WS 不走中间件，从 header 或 query param 获取身份）
3. 构建 Pod WS 地址（port-forward 或直连 Pod IP）
4. 连接 Pod WS，设置 `Origin: http://127.0.0.1`（OpenClaw Control UI 要求）
5. 发送 connect 握手帧：
   - `client.id = "openclaw-control-ui"`，`mode = "webchat"`
   - `role = "operator"`，请求 `operator.read/write/admin/approvals/pairing` 权限
   - `caps = ["tool-events"]`（接收工具执行事件）
   - `auth.token = gateway_token`（实例认证）
6. 等待 hello-ok 响应（跳过 `connect.challenge` 等中间事件），转发给前端
7. 启动双向透传 goroutine

### RPC 方法白名单

前端发来的 `type=req` 帧经过白名单过滤，只允许：

| 方法 | 用途 |
|------|------|
| `chat.send` | 发送聊天消息 |
| `chat.abort` | 中止正在执行的回复 |
| `chat.history` | 获取聊天历史 |
| `sessions.list` | 列出会话 |
| `sessions.delete` | 删除会话 |
| `sessions.patch` | 更新会话属性 |
| `exec.approval.resolve` | 审批命令执行请求 |
| `plugin.approval.resolve` | 审批插件操作请求 |

管理类方法（`config.set`、`config.apply`、`skills.install` 等）被拦截，防止用户通过 WS 绕过管控修改实例配置。非 `req` 类型帧（如心跳）直接放行。

### Session 隔离

前端每个 WS 连接通过 `X-Openclaw-Session-Key` header 生成独立 session。OpenClaw 的工具事件（tool-events）按 sessionKey 过滤，确保多窗口/多用户不互相干扰。

## Exec 审批机制

OpenClaw 支持命令执行审批流程：gateway 广播 `exec.approval.requested` 事件，前端展示审批横幅，用户决策通过 `exec.approval.resolve` RPC 回传。

**当前状态：** 由于 OpenClaw 2026.4.2 版本的 loopback pairing bug，审批通过后 followup 消息无法正确路由，导致审批功能暂不可用。当前 exec 使用 `security: "full"` 跳过审批。待 OpenClaw 修复后可切换为交互式审批模式。

## 数据模型

### instances 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL | 主键 |
| owner_upn | TEXT | 用户 UPN（来自 UniAuth） |
| name | TEXT | 实例名（同用户下唯一） |
| status | TEXT | creating / running / stopped / error / deleting |
| image | TEXT | Docker 镜像地址 |
| pod_ip | TEXT | Pod IP（running 时由 SyncService 更新） |
| api_mode | TEXT | API 配置模式：`auto`（Open-Platform 分配）或 `manual`（用户自带） |
| provider | TEXT | 模型提供商标识（默认 `openrouter`） |
| api_key | TEXT | manual 模式下用户提供的 API Key |
| base_url | TEXT | manual 模式下用户提供的 Base URL |
| gateway_token | TEXT | 实例 Gateway 认证 token（随机生成） |
| cpu_cores | TEXT | CPU 资源量，K8s 格式（如 `"2"`, `"200m"`） |
| memory_gb | TEXT | 内存资源量，K8s 格式（如 `"4Gi"`, `"512Mi"`） |
| disk_gb | TEXT | 磁盘资源量，K8s 格式（如 `"500Mi"`, `"10Gi"`） |
| gpu_count | INT | GPU 数量 |
| api_key_hash | TEXT | auto 模式下实例关联的 API Key 哈希 |
| container_port | INT | Gateway API 端口（默认 18789） |

### user_quotas 表

每用户一条记录，定义资源上限。无记录时使用默认配额（3 实例、8 CPU、16GB 内存、50GB 存储）。

### api_key_provisions 表

记录 ClawUnit 通过 UniAuth 为实例创建的 API Key。`revoked_at` 非空表示已吊销。

### audit_events 表

记录管理操作和聊天请求，包含 actor、action、resource 类型和 JSONB 详情。
