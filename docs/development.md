# 开发指南

## 项目结构

```
ClawUnit/
├── api/                              # 请求/响应结构体定义
│   ├── instances/v1/instances.go     #   实例 CRUD + 生命周期
│   ├── skills/v1/skills.go           #   技能管理
│   ├── transfer/v1/transfer.go       #   配置导入导出
│   └── admin/v1/admin.go             #   管理员接口
├── internal/
│   ├── cmd/cmd.go                    # 入口：路由注册、中间件链、启动 SyncService
│   ├── controller/                   # HTTP 处理器
│   │   ├── instances/instances.go    #   实例控制器
│   │   ├── skills/skills.go          #   技能控制器
│   │   ├── transfer/transfer.go      #   迁移控制器
│   │   └── admin/admin.go            #   管理员控制器
│   ├── dao/                          # GoFrame 自动生成的 DAO
│   ├── model/entity/                 # 自动生成的实体
│   ├── model/do/                     # 自动生成的 DO
│   ├── service/
│   │   ├── k8s/                      # Kubernetes 资源操作
│   │   │   ├── client.go             #   K8s 客户端初始化 + 命名工具
│   │   │   ├── pod.go                #   Pod CRUD（init container + gateway container）
│   │   │   ├── pvc.go                #   PVC 创建
│   │   │   ├── namespace.go          #   用户 Namespace 管理
│   │   │   ├── network_policy.go     #   出口网络策略
│   │   │   ├── configmap.go          #   ConfigMap 管理（openclaw.json 配置）
│   │   │   ├── portforward.go        #   自动 port-forward（out-of-cluster 开发模式）
│   │   │   └── cleanup.go            #   资源回滚清理
│   │   ├── lifecycle/lifecycle.go    # 实例生命周期编排（含回滚逻辑）
│   │   ├── sync/sync.go             # K8s → DB 状态同步（后台 goroutine）
│   │   ├── uniauth/client.go        # UniAuth HTTP 客户端（权限、API Key）
│   │   ├── proxy/wschat.go          # WebSocket 聊天桥接（双向透传到 Pod）
│   │   └── audit/audit.go           # 审计日志写入
│   └── middlewares/
│       ├── uni_res.go                # UniResMiddleware 统一响应格式
│       └── auth.go                   # InjectIdentity + RequireAdmin
├── manifest/
│   ├── config/config.yaml            # 运行配置
│   └── deploy/migrations/            # 数据库迁移 SQL
├── docs/                             # 文档
├── main.go                           # 程序入口
├── go.mod
├── .golangci.yml                     # 严格 lint 配置
└── CLAUDE.md                         # AI 辅助开发指引
```

## 代码规范

### 分层约束

```
api/         → 只定义 Req/Res 结构体
controller/  → HTTP 入口，可直接访问 DAO
service/     → 业务逻辑，禁止 import controller 或 api
middlewares/ → 请求预处理
```

- Controller 可直接调用 DAO，不必为每个操作都经过 service 层
- 仅当逻辑在 >= 2 处复用时才提取到 service
- 层级依赖违反是 lint 错误（depguard）

### 错误处理

```go
// 验证失败、资源不存在
gerror.NewCodef(gcode.CodeNotFound, "instance %d not found", id)

// 运行时错误（DB、K8s、HTTP 调用）
gerror.Wrapf(err, "failed to create pod for instance %d", id)
```

- 不要吞掉错误，至少 `g.Log().Errorf(ctx, ...)` 后再返回
- 使用 GoFrame 预定义的 `gcode.*` 常量，注意 `gcode.CodeForbidden` 不存在，用 `gcode.CodeNotAuthorized`

### 注释

- 使用中文
- 解释意图（为什么），不解释代码字面意思
- 显而易见的赋值不需要注释

### 安全

- 禁止日志输出 API Key 或 Token
- 在系统边界（中间件、控制器）做校验，service 内部不重复校验

## API 端点参考

所有 API 基础路径为 `/api/`。前端通过 `X-Service: clawunit` 请求头路由到 ClawUnit 服务。

### 实例管理 `/api/instances/v1`

中间件链：UniResMiddleware → InjectIdentity

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| GET | `/list` | 用户实例列表 | `status`(可选), `page`, `pageSize` |
| GET | `/detail` | 实例详情 | `id` |
| POST | `/create` | 创建实例 | `name`(必填), `description`, `image`, `storageClass`, `gpuCount`, `apiMode`, `modelId`, `apiKey`, `baseUrl` |
| POST | `/update` | 更新实例元数据 | `id`(必填), `name`, `description` |
| POST | `/delete` | 删除实例及 K8s 资源 | `id` |
| POST | `/start` | 启动已停止或异常的实例 | `id`（接受 stopped 和 error 状态） |
| POST | `/stop` | 停止运行中的实例 | `id` |
| POST | `/restart` | 重启实例 | `id` |
| GET | `/status` | 实例实时状态 | `id` |
| GET | `/quota` | 用户配额和用量 | 无 |

#### 创建实例 — 请求示例

**manual 模式（默认）：**

```json
POST /api/instances/v1/create
{
    "name": "my-openclaw",
    "description": "个人编码助手",
    "apiMode": "manual",
    "modelId": "qwen/qwen3.6-plus-preview:free",
    "apiKey": "sk-or-xxx",
    "baseUrl": "https://openrouter.ai/api/v1"
}
```

**auto 模式（通过 Open-Platform 分配 key）：**

```json
POST /api/instances/v1/create
{
    "name": "my-openclaw",
    "description": "组织内部实例",
    "apiMode": "auto",
    "modelId": "gpt-4o"
}
```

> **注意：** 资源配额（CPU、内存、磁盘）现在由服务端配置统一控制，前端不再传递。`apiMode` 默认为 `manual`。

#### 创建实例 — 响应示例

```json
{
    "code": 0,
    "message": "",
    "data": {
        "id": 42
    }
}
```

#### 实例列表 — 响应示例

```json
{
    "code": 0,
    "message": "",
    "data": {
        "list": [
            {
                "id": 42,
                "name": "my-openclaw",
                "status": "running",
                "cpuCores": "2",
                "memoryGb": "4Gi",
                "diskGb": "500Mi",
                "gpuCount": 0,
                "gpuEnabled": false,
                "image": "ghcr.io/openclaw/openclaw:latest",
                "storageClass": "longhorn",
                "containerPort": 18789,
                "ownerUpn": "alice@cuhk.edu.cn",
                "createdAt": "2026-03-30 10:00:00"
            }
        ],
        "total": 1
    }
}
```

> **注意：** `cpuCores`、`memoryGb`、`diskGb` 现在是 TEXT 类型，返回 K8s 资源格式字符串（如 `"2"`, `"4Gi"`, `"500Mi"`）。

### 聊天代理 `/api/chat/v1`

不走中间件，handler 内部认证。

| 方法 | 路径 | 说明 | 参数 |
|------|------|------|------|
| GET | `/ws` | WebSocket 聊天桥接 | `instanceId`(query), `userId`(query/header) |

#### WebSocket 聊天端点

`GET /api/chat/v1/ws` 不走中间件，handler 内部认证（从 `X-User-ID` header 或 `userId` query param 获取身份）。连接建立后双向透传 OpenClaw WS 帧，前端发送的 RPC 方法经白名单过滤。

**WS RPC 白名单：**
- `chat.send` / `chat.abort` / `chat.history` — 聊天操作
- `sessions.list` / `sessions.delete` / `sessions.patch` — 会话管理
- `exec.approval.resolve` / `plugin.approval.resolve` — 审批操作

管理类 RPC（`config.set`、`skills.install` 等）被拦截。

**Session 管理：** 前端每个 WS 连接通过 `X-Openclaw-Session-Key` header 生成独立 session，工具事件按 sessionKey 隔离。通过 `sessions.list` 可列出会话，`sessions.delete` 可删除，`sessions.patch` 可更新会话属性（如重命名）。

### 技能管理 `/api/skills/v1`

中间件链：UniResMiddleware → InjectIdentity

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/system/list` | 系统预置技能列表 |
| GET | `/user/list` | 用户自定义技能列表 |
| POST | `/user/upload` | 上传技能（multipart/form-data） |
| POST | `/user/delete` | 删除用户技能 |

### 配置迁移 `/api/transfer/v1`

中间件链：UniResMiddleware → InjectIdentity

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/export` | 导出实例配置（tar.gz 二进制流） |
| POST | `/import` | 导入配置（multipart/form-data） |

### 管理员 `/api/admin/v1`

中间件链：UniResMiddleware → InjectIdentity → RequireAdmin

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/instances/list` | 全局实例列表（可按用户/状态过滤） |
| POST | `/quotas/update` | 设置用户配额 |
| GET | `/sync/status` | K8s 同步服务状态 |
| POST | `/sync/force` | 强制触发一次同步 |

## 统一响应格式

除聊天代理外，所有 API 响应经 UniResMiddleware 包装为统一格式：

```json
{
    "code": 0,
    "message": "",
    "data": { ... }
}
```

错误时 `code` 非 0，`message` 携带错误描述。常见 code：

| code | 含义 |
|------|------|
| 0 | 成功 |
| 51 | Not Found |
| 52 | Not Authorized（鉴权失败或权限不足） |
| 50 | 内部错误 |

## 新增功能流程

以"新增一个 API 端点"为例：

### 1. 定义请求/响应结构体

在 `api/{module}/v1/{module}.go` 中添加：

```go
type MyNewReq struct {
    g.Meta `path:"/my-new" method:"post" tags:"Module" summary:"新功能" dc:"描述"`

    ParamA string `json:"paramA" v:"required" dc:"参数A"`
}

type MyNewRes struct {
    Result string `json:"result" dc:"结果"`
}
```

GoFrame 根据 `g.Meta` 标签自动注册路由。

### 2. 实现控制器方法

在 `internal/controller/{module}/{module}.go` 中添加对应方法：

```go
func (c *Controller) MyNew(ctx context.Context, req *v1.MyNewReq) (res *v1.MyNewRes, err error) {
    ownerUpn := ctx.Value("ownerUpn").(string)
    // 业务逻辑...
    return &v1.MyNewRes{Result: "ok"}, nil
}
```

### 3. 如果涉及 K8s 操作

在 `internal/service/k8s/` 中添加资源操作方法，在 `lifecycle.go` 中编排调用顺序和回滚逻辑。

### 4. 如果涉及数据库变更

在 `manifest/deploy/migrations/` 中添加新的迁移文件（如 `002_add_xxx.sql`），然后重新生成 DAO：

```bash
gf gen dao
```

### 5. 检查

```bash
gofmt -w .
golangci-lint run --fix
golangci-lint run
go test ./...
```

## 关键实现细节

### configmap.go — OpenClaw 配置生成

`k8s.OpenClawConfig(opts)` 生成 `openclaw.json` 配置字符串，配置内容：

- **Gateway**：端口 18789，绑定 LAN，token 认证（`${OPENCLAW_GATEWAY_TOKEN}` 环境变量注入），`controlUi.dangerouslyDisableDeviceAuth` 和 `allowInsecureAuth` 启用
- **Agents**：默认模型为 `custom/{modelID}`，workspace 路径 `/home/node/.openclaw/workspace`，sandbox 关闭
- **Tools**：profile（默认 full），exec 安全级别 full，fs.workspaceOnly 为 false
- **Browser**：executablePath 指向 Playwright Chromium，headless 模式，noSandbox，ssrfPolicy 控制私有网络访问
- **Plugins**：entries 配置（搜索引擎等扩展通过此注入）
- **Cron**：默认关闭（`enabled: false`）
- **Models.providers.custom**：baseUrl 和 apiKey 通过 `${CUSTOM_BASE_URL}` / `${CUSTOM_API_KEY}` 环境变量注入

Pod 启动命令为 `sh -c "npx playwright install-deps chromium > /dev/null 2>&1; exec runuser -u node -- node /app/dist/index.js gateway run"`，先以 root 安装 Chromium 系统依赖再降权运行。Startup 探针失败阈值 120（间隔 5s，约 10 分钟超时），因安装依赖耗时。

`k8s.CreateConfigMap()` 将配置存入 K8s ConfigMap（命名 `clawunit-{instanceID}-config`），已存在时更新。`k8s.DeleteConfigMap()` 在实例删除时清理。

### portforward.go — 自动 Port-Forward

本地开发（out-of-cluster）时，聊天代理无法直连 Pod IP。`portforward.go` 提供自动 port-forward 机制：

- `GetLocalPort(podName, podNamespace, remotePort)` — 获取本地转发端口。in-cluster 返回 0，out-of-cluster 自动建立 SPDY port-forward
- `ClosePortForward(podName, podNamespace)` — 关闭并清理 port-forward 连接

内部维护 `pfCache`（`map[string]*portForwardEntry`），key 为 `namespace/podName`。每次获取时先 TCP 探活，端口不通则重建。

### 实例生命周期回滚

`lifecycle.Create` 中，每个步骤失败时清理已创建的资源。回滚顺序包含 ConfigMap：

```
步骤 11 失败（Pod）→ 删 NetworkPolicy → 删 ConfigMap → 吊销 API Key → 删 DB 记录
步骤 8 失败（ConfigMap）→ 吊销 API Key → 删 DB 记录
```

### SyncService 状态同步

后台 goroutine 每 5 秒：
1. 查询 DB 中 `status=creating` 或 `running` 的实例
2. 通过 K8s API 查询每个实例的 Pod 状态
3. 根据 Pod 状态更新 DB（Running+Ready → running，Failed → error 等）
4. Pod Ready 时同步更新 `pod_ip`

### 用户身份传递

```
Gateway 注入 X-User-ID header
  → InjectIdentity 中间件提取并调 UniAuth 验证
  → 存入 ctx("ownerUpn")
  → Controller 从 ctx 取 ownerUpn
  → 所有 DB 查询带 owner_upn 过滤
```

## Swagger / OpenAPI

启动后访问：

- OpenAPI JSON: `http://localhost:8282/api.json`
- Swagger UI: `http://localhost:8282/swagger`

GoFrame 根据 `api/` 目录中的 `g.Meta` 标签和 `dc` 注释自动生成 OpenAPI 文档。
