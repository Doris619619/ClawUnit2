# 部署运维

## 前置条件

| 组件 | 版本要求 | 用途 |
|------|---------|------|
| Go | >= 1.26 | 编译运行 |
| PostgreSQL | >= 14 | 数据存储 |
| Kubernetes | >= 1.25 | Pod 编排 |
| UniAuth | 可达即可 | 鉴权、API Key 管理 |
| Open-Platform | 可达即可 | LLM 代理（Pod 内部调用） |

## 数据库初始化

### 1. 创建数据库

```bash
psql "postgresql://user:pass@host:port/postgres" -c "CREATE DATABASE clawunit;"
```

### 2. 执行迁移

```bash
psql "postgresql://user:pass@host:port/clawunit" -f manifest/deploy/migrations/001_init_schema.sql
```

迁移脚本创建 5 张表：

| 表 | 用途 |
|---|------|
| `instances` | 实例记录（状态、资源配置、API 模式、Pod 元数据） |
| `user_quotas` | 用户资源配额上限 |
| `skills` | 技能元数据（系统级 / 用户级） |
| `api_key_provisions` | 为实例创建的 UniAuth API Key 记录（仅 auto 模式） |
| `audit_events` | 审计日志 |

> 注意：不需要创建 users 表。用户身份完全由 UniAuth 管理，ClawUnit 使用 `owner_upn`（TEXT）关联。

### instances 表新增字段

相比早期版本，instances 表新增了以下字段：

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `api_mode` | TEXT | `'manual'` | API 配置模式：`auto`（Open-Platform 自动分配）或 `manual`（用户手动填写） |
| `provider` | TEXT | `'openrouter'` | 模型提供商标识 |
| `api_key` | TEXT | `''` | manual 模式下用户提供的 API Key |
| `base_url` | TEXT | `''` | manual 模式下用户提供的 API Base URL |
| `gateway_token` | TEXT | `''` | 实例 Gateway 认证 token，创建时随机生成 |
| `storage_class` | TEXT | `'standard'` | K8s StorageClass 名称 |
| `mount_path` | TEXT | `'/home/user/data'` | 用户数据卷挂载路径 |
| `gpu_enabled` | BOOLEAN | `FALSE` | 是否启用 GPU |

### 资源字段类型变更

资源配置字段已从 INT 改为 TEXT，使用 K8s 原生资源格式：

| 字段 | 旧类型 | 新类型 | 示例值 |
|------|--------|--------|--------|
| `cpu_cores` | INT（核数） | TEXT（K8s 格式） | `"2"`, `"200m"`, `"500m"` |
| `memory_gb` | INT（GB 数） | TEXT（K8s 格式） | `"4Gi"`, `"512Mi"`, `"2Gi"` |
| `disk_gb` | INT（GB 数） | TEXT（K8s 格式） | `"500Mi"`, `"10Gi"` |

这些值直接传给 K8s `resource.MustParse()`，支持 K8s 的全部资源表示法。

## 配置参考

配置文件位于 `manifest/config/config.yaml`，使用 GoFrame 标准格式。

### server — HTTP 服务

```yaml
server:
  httpPort: 8282              # 监听端口
  openapiPath: "/api.json"    # OpenAPI spec 路径
  swaggerPath: "/swagger"     # Swagger UI 路径
```

### database — 数据库连接

```yaml
database:
  default:
    link: "pgsql:user:pass@tcp(host:port)/clawunit"
    timezone: "Asia/Shanghai"
    createdAt: "created_at"   # 自动填充创建时间字段名
    updatedAt: "updated_at"   # 自动填充更新时间字段名
    debug: true               # 打印 SQL（生产环境建议关闭）
```

`link` 格式：`pgsql:用户名:密码@tcp(主机:端口)/数据库名`

如果密码含特殊字符（如 `@`），GoFrame 的 link DSN 格式以最后一个 `@tcp(` 分割，所以密码中的 `@` 无需转义。

### client — 外部服务

```yaml
client:
  uniauth:
    url: "http://localhost:8004"      # UniAuth 地址
  openPlatform:
    url: "http://localhost:8032"      # Open-Platform 地址
```

本地开发时通常使用 `localhost`，生产环境替换为集群内部服务地址（如 `http://uniauth-svc.system:8004`）。

### k8s — Kubernetes

```yaml
k8s:
  mode: "auto"                        # auto | incluster | outofcluster
  namespace: "clawunit"               # 基础命名空间前缀
  storageClass: "standard"            # PVC 使用的 StorageClass
  systemSkillsPVC: "clawunit-system-skills"  # 系统技能 PVC 名称
  playwrightPVC: "clawunit-playwright-browsers"  # Playwright 浏览器 PVC 名称
  kubeconfig: ""                      # out-of-cluster 模式下 kubeconfig 路径
```

| 参数 | 说明 |
|------|------|
| `mode: auto` | 自动检测：在 Pod 内用 InCluster，本地开发用 OutOfCluster |
| `mode: incluster` | 强制使用 ServiceAccount Token（部署在 K8s 内部时） |
| `mode: outofcluster` | 强制使用 kubeconfig 文件（本地开发时） |
| `namespace` | 用户 Namespace 前缀，实际名称为 `{namespace}-user-{upn_hash}` |
| `storageClass` | 用户 PVC 创建时使用的 StorageClass，需提前在集群中配置 |
| `systemSkillsPVC` | 系统预置技能的 PVC 名称，需提前手动创建并填充内容（全局共享 ReadOnlyMany） |
| `playwrightPVC` | Playwright 浏览器 PVC 名称，需提前创建并安装 Chromium（全局共享 ReadOnlyMany），默认 `clawunit-playwright-browsers` |
| `kubeconfig` | 留空则使用 `~/.kube/config`，仅 outofcluster 模式生效 |

### instance — OpenClaw 实例默认值

```yaml
instance:
  defaultImage: "ghcr.io/openclaw/openclaw:latest"   # OpenClaw 镜像
  defaultCPU: "2"             # K8s 格式 CPU（如 "200m", "1", "2"）
  defaultMemory: "4Gi"        # K8s 格式内存（如 "512Mi", "4Gi"）
  defaultDisk: "500Mi"        # K8s 格式磁盘（如 "500Mi", "10Gi"）
  containerPort: 18789        # Gateway API 端口（探针 + 聊天代理共用）
  gatewayPort: 18789          # OpenClaw Gateway API 端口（聊天代理使用）
  gatewayToken: ""            # 已废弃：现在每实例自动生成 gateway_token
  mountPath: "/home/user/data"  # 用户数据卷挂载路径
```

> **OpenClaw 镜像说明：** 推荐使用 `ghcr.io/openclaw/openclaw:latest`。Pod 中只运行 Gateway 模式（`node /app/dist/index.js gateway run`），不启动 Web UI。如需使用内部镜像仓库，配置为对应的 mirror 地址即可。

> **端口说明：** 当前 Pod 只暴露一个端口 18789（Gateway API），同时用于健康探针和聊天代理。`containerPort` 和 `gatewayPort` 配置为相同值。

### ConfigMap 与 Init Container

实例创建时自动生成 `openclaw.json` 配置文件并存入 ConfigMap。Pod 启动时通过 init container（`busybox:1.37`）将配置文件从 ConfigMap 卷拷贝到 OpenClaw 工作目录：

```
ConfigMap (openclaw-config) → /config/openclaw.json
  ↓ init-config container
/home/node/.openclaw/openclaw.json
  ↓ gateway container 启动时读取
OpenClaw Gateway 运行配置
```

配置中包含 gateway（含 controlUi 认证跳过）、agents、tools（profile/exec/fs）、browser（Playwright Chromium 路径/headless/noSandbox/ssrfPolicy）、plugins、cron（默认关闭）、models 等段落。`${OPENCLAW_GATEWAY_TOKEN}`、`${CUSTOM_API_KEY}`、`${CUSTOM_BASE_URL}` 由 OpenClaw 在启动时从环境变量替换。

## Kubernetes 集群准备

### RBAC 权限

ClawUnit 的 ServiceAccount 需要以下权限：

```yaml
# 对用户 namespace 的权限（clawunit-user-* 前缀）
- apiGroups: [""]
  resources: ["namespaces", "pods", "persistentvolumeclaims", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies"]
  verbs: ["get", "list", "create", "update", "delete"]
# port-forward 需要 pods/portforward 子资源权限（out-of-cluster 开发模式）
- apiGroups: [""]
  resources: ["pods/portforward"]
  verbs: ["create"]
```

### 系统技能 PVC

需要在 ClawUnit 可访问的命名空间中预先创建系统技能 PVC：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clawunit-system-skills
  namespace: clawunit     # 或 systemSkillsPVC 所在命名空间
spec:
  accessModes:
    - ReadOnlyMany
  resources:
    requests:
      storage: 5Gi
  storageClassName: standard
```

将系统技能文件放入该 PVC 后，所有 OpenClaw Pod 将以只读方式挂载到 `/skills/system`。

### Playwright 浏览器 PVC

管理员需预先创建 Playwright 浏览器 PVC 并安装 Chromium，所有 OpenClaw Pod 以只读方式挂载：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clawunit-playwright-browsers   # 或 k8s.playwrightPVC 配置值
  namespace: clawunit
spec:
  accessModes:
    - ReadOnlyMany
  resources:
    requests:
      storage: 2Gi
  storageClassName: standard
```

安装 Chromium 的方法参见 [管理员初始化指南](admin-setup.md)。PVC 挂载到 Pod 的 `/home/node/.cache/ms-playwright`，环境变量 `PLAYWRIGHT_BROWSERS_PATH` 指向此路径。

### StorageClass

确保 `k8s.storageClass` 配置的 StorageClass 在集群中存在：

```bash
kubectl get storageclass
```

用户 PVC 使用 `ReadWriteOnce` 模式。

### UniAuth 权限配置

在 UniAuth 管理后台完成以下配置：

1. 添加 `clawunit` 作为 object
2. 为普通用户添加策略 `(user_upn, clawunit, access)`
3. 为管理员添加策略 `(admin_upn, clawunit, admin)`
4. 在 Feature Gate 中注册 `clawunit`，控制 UI 侧入口可见性

## 启动

### 本地开发

```bash
# 确保 kubeconfig 可用
kubectl cluster-info

# 确保数据库可达
psql "postgresql://user:pass@host:port/clawunit" -c "SELECT 1;"

# 启动服务
go run main.go
```

服务启动后：
- HTTP API: `http://localhost:8282`
- Swagger UI: `http://localhost:8282/swagger`
- OpenAPI Spec: `http://localhost:8282/api.json`

### 生产编译

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o clawunit -ldflags="-w -s" main.go
```

### 容器化部署

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o clawunit -ldflags="-w -s" main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/clawunit /clawunit
COPY --from=builder /app/manifest /manifest
ENTRYPOINT ["/clawunit"]
```

部署为 Deployment 时使用 `incluster` 模式，确保 ServiceAccount 有上述 RBAC 权限。

## 健康检查

ClawUnit 本身可用以下方式检测：

```bash
# 服务是否启动
curl http://localhost:8282/api.json

# 数据库是否连通（查询实例列表会触发 DB 访问）
# 需要有效的 X-User-ID header
curl -H "X-User-ID: test@example.com" http://localhost:8282/api/instances/v1/list
```

## 日志

使用 GoFrame 标准日志：

```yaml
logger:
  level: "debug"    # debug | info | notice | warning | error
  stdout: true      # 输出到标准输出
  # path: "/var/log/clawunit"  # 可选，输出到文件
```

生产环境建议设置 `level: "info"` 并关闭 `database.default.debug`。
