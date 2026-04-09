# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ClawUnit 是一个基于 GoFrame v2 的 Go 服务，为 CUHKSZ 提供 OpenClaw 实例的 Kubernetes 生命周期管理。每个用户可创建独立的 OpenClaw Gateway 实例（K8s Pod），通过自带 API Key 或 Open Platform 自动分配连接 LLM 服务。

## Architecture

**请求流:**
```
HTTP → UniResMiddleware → InjectIdentity (X-User-ID) → Controller → DAO / K8s Service
```

**路由组 (internal/cmd/cmd.go):**
- `/api/instances/v1` — 实例 CRUD + 生命周期（create/start/stop/restart/delete）+ 配置热更新
- `/api/skills/v1` — 技能管理（系统级只读 + 用户级读写）
- `/api/transfer/v1` — OpenClaw 配置导入导出
- `/api/admin/v1` — 管理员接口（全局实例列表、配额管理、同步控制）
- `/api/chat/v1` — 聊天代理：HTTP SSE（`/completions`）+ WebSocket（`/ws`，无中间件）
- `/api/channels/v1` — 渠道插件管理（install/uninstall/login SSE/status/restart-gateway）

**关键目录:**
- `api/` — 请求/响应结构体定义（含 gvalid 标签）
- `internal/controller/` — HTTP 处理器
- `internal/service/k8s/` — Kubernetes 资源管理（Pod/PVC/Service/Namespace/NetworkPolicy/ConfigMap/PortForward/Exec）
- `internal/service/lifecycle/` — 实例生命周期编排（含回滚）
- `internal/service/sync/` — K8s 状态同步到数据库（Pod phase → 实例 status + Pod IP）
- `internal/service/proxy/` — 聊天代理：HTTP proxy + WS 双向桥接（out-of-cluster 自动 port-forward）
- `internal/service/uniauth/` — UniAuth 客户端（权限校验、API Key 管理）
- `internal/middlewares/` — 统一响应、身份注入、管理员鉴权

**K8s 资源模型:**
- 一个用户 = 一个 namespace (`{base}-user-{upn_hash}`)
- 每个实例 = Pod + ClusterIP Service + NetworkPolicy + ConfigMap
- Pod 结构: init container (busybox 首次复制配置) + gateway container
- gateway 容器以 root 启动（装 Chromium 依赖后 runuser 切 node 用户）
- ConfigMap 包含 `openclaw.json`（gateway 配置 + LLM provider 配置 + controlUi 认证配置）
- 存储: 用户数据 PVC 直接挂载 `/home/node/.openclaw`（整目录持久化）+ 系统技能 PVC (ReadOnly) + Playwright PVC (ReadOnly) + /dev/shm (Memory EmptyDir)

**实例生命周期:**
```
Create: 配额检查 → 生成 gateway token → 插入 DB → 构建环境变量 →
        生成 OpenClaw 配置 → 确保 namespace → 创建 ConfigMap →
        确保 PVC → 创建 NetworkPolicy → 创建 Pod → 创建 Service
        （任意步骤失败自动回滚）

Start:  从 stopped/error 状态 → 恢复环境变量 → 创建 Pod → 创建 Service
        （ConfigMap/NetworkPolicy/PVC 已存在，不重建）

Stop:   删除 Pod → 删除 Service → 状态改为 stopped
        （保留 ConfigMap/NetworkPolicy/PVC/DB 记录）

Delete: 关闭 port-forward → 删除 Pod（等待终止 30s）→ 删除 Service →
        删除 NetworkPolicy → 删除 ConfigMap → 回收 API Key → 删除 DB 记录
        （用户 PVC 不删除）

Restart: Stop → Start
```

**OpenClaw 集成:**
- 镜像: `ghcr.io/openclaw/openclaw:2026.4.1`（全功能版，含浏览器/工具/技能）
- 端口: 18789（Gateway API + Control UI + WebSocket）
- 配置: init container 首次从 ConfigMap 复制到 PVC；后续 OpenClaw 运行时修改直接持久化
- LLM 通过 `sglang` provider 配置（`reasoning: true`），环境变量 `CUSTOM_API_KEY` + `CUSTOM_BASE_URL` 注入
- 后端模型（GLM-5/Qwen 等）由 sglang 部署，返回 `reasoning_content` 字段；必须使用 `sglang` provider 而非 `custom`，否则 OpenClaw 无法解析 reasoning 响应
- 每个实例有独立的 `OPENCLAW_GATEWAY_TOKEN`（存 DB，proxy 代理时使用）
- 健康检查: exec probe（`node -e 'require("http").get(...)'`），startup 120 次 × 5s
- Chromium 依赖: Pod 启动时通过 `npx playwright install-deps chromium` 安装系统库；浏览器二进制来自 Playwright PVC
- Browser: `ssrfPolicy.dangerouslyAllowPrivateNetwork` 必须为 `true`（browser service 需连本地 CDP）
- 渠道插件: 安装到 PVC 持久化的 `extensions/` 目录，安装后通过 `PatchConfigMapPlugin` 更新 ConfigMap 的 plugins.allow + entries

**WebSocket 聊天桥接 (internal/service/proxy/wschat.go):**
```
前端浏览器 WS ←→ ClawUnit WS Handler ←→ OpenClaw Pod WS
```
- connect frame: `client.id="openclaw-control-ui"`, `mode="webchat"`, `caps=["tool-events"]`
- Origin header: `http://127.0.0.1`（通过 gateway origin 校验）
- gateway 配置: `controlUi.dangerouslyDisableDeviceAuth=true`（绕过设备认证）
- RPC 白名单过滤（前端→Pod 方向）：仅允许 `chat.send/abort/history`, `sessions.list/delete/patch`, `exec/plugin.approval.resolve`
- 管理类方法（config.set/apply、skills.install 等）被拦截
- session 隔离: 前端每个 WS 连接生成独立 sessionKey，事件按 sessionKey 过滤
- exec approval: gateway 广播 `exec.approval.requested`，前端展示审批横幅，用户决策通过 `exec.approval.resolve` 回传
  - **注意**: 当前 OpenClaw 2026.4.2 的 loopback pairing bug 导致审批后 followup 失败，exec 暂用 `security: "full"`

**API 配置双模式:**
- `manual`（默认）: 用户提供 apiKey + baseUrl + modelId
- `auto`: 通过 UniAuth 创建 API Key，使用 Open Platform

**NetworkPolicy 出站规则:**
- DNS: kube-system 的 53 端口
- 系统服务: 系统 namespace 的 80/443/8032/8033
- 外网: 任意地址的 80/443（LLM API、apt 仓库、web_fetch）

**PVC 策略:**
| PVC | 范围 | 访问模式 | 挂载点 | 生命周期 |
|-----|------|----------|--------|----------|
| 系统技能 | 每用户 namespace | ReadOnly | `/skills/system` | 管理员预创建 |
| Playwright | 每用户 namespace | ReadOnly | `/home/node/.cache/ms-playwright` | 管理员预创建 |
| 用户数据 | 每实例一个 | ReadWriteOnce | `/home/node/.openclaw`（整目录） | 创建实例时自动创建，不随实例删除 |

用户数据 PVC 持久化所有 OpenClaw 运行数据：workspace、memory、agents、extensions（插件）、credentials（渠道凭证）、openclaw.json（运行时配置）。

**开发环境 out-of-cluster 支持:**
- proxy 和 WS 桥自动检测 out-of-cluster 模式，通过 client-go SPDY portforward 建立到 Pod 的连接
- 缓存 port-forward 连接，TCP 活性检测后复用
- in-cluster 部署时直连 Pod IP，无额外开销

## Code Style

Observed conventions from the codebase. Follow these when writing or reviewing code:

### Layer boundaries

- **Controller → DAO is allowed.** Don't extract to a service just to avoid the direct access.
- **Extract to service only when logic is reused in ≥ 2 places.** Controllers that only call DAO and return are the norm; pass-through wrappers add indirection for no benefit.
- Layer dependency violations are **lint errors** (depguard): service cannot import controller/api packages, controllers cannot import other controllers.

### Error handling

- `gerror.NewCodef(gcode.CodeNotFound, "...")` — validation failures and resource-not-found.
- `gerror.Wrapf(err, "...")` — wrapping runtime errors from DAO or external calls.
- Never swallow errors silently; log with `g.Log().Errorf(ctx, ...)` before returning false or a zero value.
- Use GoFrame's predefined `gcode.*` constants. `gcode.CodeForbidden` does **not** exist — use `gcode.CodeNotAuthorized`.

### GoFrame ORM

- Context 只在 `.Ctx(ctx)` 设置一次，终端方法（`All`/`One`/`Count`/`Scan`/`Insert`/`Update`/`Delete`/`InsertAndGetId`）**不传 ctx 参数**。
- 使用 DO struct（`do.Instances{}`）进行 insert/update，nil 字段自动忽略。
- 示例: `dao.Instances.Ctx(ctx).Where("id", id).Data(do.Instances{Status: "running"}).Update()`

### DRY and helper extraction

- Extract a helper when the same logic appears in ≥ 2 places (dupl linter threshold: ~150 tokens).
- Do **not** extract prematurely — a helper used exactly once adds noise.
- Name helpers after what they compute, not where they're called from.

### Pointer to literal: use `new(literal)` (Go 1.26+)

Go 1.26 扩展了内建 `new` 函数，允许直接传入字面量或函数返回值，返回指向该值副本的指针。

**用法：**
```go
// ✅ 字面量
b := new(false)              // *bool 指向 false
i := new(42)                 // *int 指向 42
p := new(corev1.ProtocolTCP) // *corev1.Protocol

// ✅ 函数返回值
port := new(intstr.FromInt(53)) // *intstr.IntOrString
```

**禁止：** 不要对**已有变量**用 `new(x)`，会有指针 alias 的隐患（`new` 复制值，但容易让人误以为返回的是变量本身的地址）：
```go
var x int = 10
p := new(x)  // ❌ 看起来像 &x，实际是新的拷贝
```

**禁止造轮子：** 不要写 `boolPtr` / `int32Ptr` / `protocolPtr` 之类的 helper 函数，直接用 `new(...)`。

**仍然用 `&Struct{...}` 的场景：** 构造结构体指针（`&v1.CreateRes{Id: 1}`）保持 idiomatic Go 写法，不要改成 `new(v1.CreateRes{Id: 1})`。`new(literal)` 主要用于**基础类型/常量/函数返回值**。

### File organization

适用于 `internal/service/` 和 `internal/controller/` 下的所有包：

- **一个文件只放一个公共函数（首字母大写的函数）。** 文件名 = 公共函数的 snake_case 形式（`CreatePod` → `create_pod.go`，`PatchConfigMapPlugin` → `patch_configmap_plugin.go`）。
- **包级私有 helper（被多个公共函数复用）** 放在和包同名的 `<package>.go` 文件里。这个文件也存放包级类型定义、常量、变量。
- **只被某一个公共函数使用的私有 helper** 跟它的使用者放在同一个文件里。
- **不要新建 `helpers.go` / `utils.go` / `common.go` 之类的笼统文件。**

例：
```
internal/service/k8s/
  k8s.go                    # 包级 Client 类型、共享的 sanitizeK8sName / upnHash 等
  create_pod.go             # CreatePod + 仅它使用的 buildPodSpec 私有函数
  get_pod.go                # GetPod
  delete_pod.go             # DeletePod
  delete_pod_and_wait.go    # DeletePodAndWait
  create_config_map.go      # CreateConfigMap
  patch_config_map_plugin.go # PatchConfigMapPlugin
  ...
```

例外：`gf gen ctrl` 生成的 controller 文件名固定为 `<module>_v1_<method>.go`，不要改名。

### Function signatures

- Remove return values that are never used by any caller (`unparam` linter).
- Avoid redundant type conversions.
- Prefer returning the minimal set of values needed; combine multiple DB lookups into a single resolver that returns everything callers need.

### Comments

- Written in **Chinese**.
- Explain **why** or **what the intent is**, not what the code literally does.
- One-line comments for obvious assignments are noise; skip them.

### Security and logging

- Do **not** log API keys or tokens, even in error paths.
- Validate at system boundaries (middleware, controller); don't re-validate inside service helpers.

## Database

PostgreSQL，通过 GoFrame ORM 访问。关键表:
- `instances` — 实例记录（含 api_mode/api_key/base_url/gateway_token/cpu_cores(TEXT)/memory_gb(TEXT)/disk_gb(TEXT)）
- `user_quotas` — 用户配额（目前只跟踪 max_instances）
- `skills` — 技能元数据
- `api_key_provisions` — UniAuth API Key 记录
- `audit_events` — 审计日志

资源字段（cpu_cores/memory_gb/disk_gb）使用 TEXT 类型存储 K8s 格式值（如 `"1"`, `"2Gi"`, `"500Mi"`）。

## Code Generation (gf gen)

**两个生成命令是强制的，手动维护对应的文件等于和未来打架。**

### `gf gen dao` — 从数据库生成 ORM 代码

**触发时机：**
- 修改了 `manifest/deploy/migrations/*.sql` 数据库 schema
- 改了 `hack/config.yaml` 的 `typeMapping`

**生成的文件（不要手动改）：**
- `internal/dao/internal/*.go` — 列名常量
- `internal/model/entity/*.go` — typed struct（用于 `.Scan(&entity)` 读取）
- `internal/model/do/*.go` — `any` 字段 struct（用于 `.Data(do.X{})` 部分更新）

**类型映射在 `hack/config.yaml`：**
```yaml
typeMapping:
  int4: { type: int32 }   # PG INTEGER → Go int32（默认是 int）
  numeric: { type: decimal.Decimal, import: github.com/shopspring/decimal }
```

修改 schema 后必须跑：`gf gen dao`

### `gf gen ctrl` — 从 API request 结构生成 controller interface

**触发时机：**
- 在 `api/<module>/v1/*.go` 新增/修改 request 结构（带 `g.Meta` 路由 tag）

**生成的文件（不要手动改）：**
- `api/<module>/<module>.go` — `IModuleV1` interface 定义
- `internal/controller/<module>/<module>.go` — `New()` + `*ControllerV1` 类型
- `internal/controller/<module>/<module>_v1_<method>.go` — **每个 API 方法一个文件**

**写实现：** 在自动生成的 `<module>_v1_<method>.go` 文件里填方法体。**不要把所有方法塞进同一个文件**，会被下次 `gf gen ctrl` 冲掉或跟新生成的文件冲突。

**Helper / 非 interface 函数放哪里：** 放在同名 `<module>.go` 文件里（gen ctrl 生成的"only once"文件，工具不会重新生成它）。**不要新建 helpers.go 之类的独立文件。**

例如 `channels.go` 里可以放 `LoginSSE` 路由处理器、`getInstancePod` 等辅助函数和私有类型。

### 通用规则

- **永远不要**手动修改 `internal/dao/internal/`、`internal/model/entity/`、`internal/model/do/` 下的文件
- **永远不要**手动修改 `api/<module>/<module>.go` 的 interface 定义（在 v1 子目录改 request 结构后跑生成）
- 新增 API 后跑：`gf gen ctrl`
- 修改 schema 后跑：`gf gen dao`
- 生成完后必须 `go build ./...` + `golangci-lint run` 验证

## Linting

Strict golangci-lint config (`.golangci.yml`) with 60+ linters. Notable limits: `gocognit: 15`, `nestif: 15`, `maintidx: 10`. Layer dependency violations are lint errors.

Run `golangci-lint run --fix` first — it auto-fixes `wsl_v5`, `nlreturn`, `gofmt`, and `goimports` issues. Only diagnose the remaining failures manually.

## Common Commands

```bash
# Run the service
go run main.go

# Build for production
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o clawunit -ldflags="-w -s" main.go

# After code changes, always run:
gofmt -w .
go test ./...
golangci-lint run
```

## Configuration (manifest/config/config.yaml)

```yaml
instance:
  defaultImage: "ghcr.io/openclaw/openclaw:2026.4.1"
  defaultCPU: "4"           # K8s 资源格式
  defaultMemory: "8Gi"
  defaultDisk: "500Mi"
  containerPort: 18789
  gatewayPort: 18789
  mountPath: "/home/user/data"
```

## Git Conventions

- Commit message 用中文
- 不加 `Co-Authored-By` 行
