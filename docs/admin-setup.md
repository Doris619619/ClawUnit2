# ClawUnit 管理员初始化指南

## Playwright 浏览器 PVC 配置

OpenClaw 使用 Playwright (Chromium) 提供浏览器自动化能力。浏览器二进制文件通过**全局共享 PVC** 预装，所有实例 Pod 以只读方式挂载。

PVC 名称通过 `k8s.playwrightPVC` 配置（默认 `clawunit-playwright-browsers`），是集群级全局资源，**只需创建一次**，所有用户命名空间的 Pod 共享使用。

### 1. 创建 PVC

在 ClawUnit 基础命名空间创建全局共享 PVC（Chromium 约需 600MB，PVC 建议 2Gi）：

```bash
NAMESPACE="clawunit"   # 或你的 k8s.namespace 配置值

kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clawunit-playwright-browsers   # 对应 k8s.playwrightPVC 配置值
  namespace: $NAMESPACE
  labels:
    app: clawunit
    component: playwright-browsers
spec:
  accessModes:
    - ReadOnlyMany
  storageClassName: longhorn
  resources:
    requests:
      storage: 2Gi
EOF
```

### 2. 运行安装 Job

```bash
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: install-playwright
  namespace: $NAMESPACE
spec:
  ttlSecondsAfterFinished: 600
  template:
    spec:
      securityContext:
        fsGroup: 1000
      containers:
        - name: install
          image: ghcr.io/openclaw/openclaw:latest
          command: ["node", "/app/node_modules/playwright-core/cli.js", "install", "chromium"]
          env:
            - name: PLAYWRIGHT_BROWSERS_PATH
              value: /browsers
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
          volumeMounts:
            - name: browsers
              mountPath: /browsers
          resources:
            requests:
              memory: 512Mi
              cpu: 500m
            limits:
              memory: 1Gi
              cpu: "1"
      restartPolicy: Never
      volumes:
        - name: browsers
          persistentVolumeClaim:
            claimName: clawunit-playwright-browsers   # 对应 k8s.playwrightPVC 配置值
  backoffLimit: 2
EOF
```

Job 会下载并安装 Chromium（约 300MB 下载，600MB 解压后），通常 2-5 分钟完成。

### 3. 验证安装结果

```bash
# 等待完成
kubectl -n $NAMESPACE wait --for=condition=complete job/install-playwright --timeout=300s

# 查看日志
kubectl -n $NAMESPACE logs -l job-name=install-playwright

# 确认 PVC 已绑定
kubectl -n $NAMESPACE get pvc clawunit-playwright-browsers
# 状态应为 Bound
```

### 4. 清理

Job 配置了 `ttlSecondsAfterFinished: 600`，完成 10 分钟后自动清理。手动清理：

```bash
kubectl -n $NAMESPACE delete job install-playwright
```

### 注意事项

- PVC accessMode 为 `ReadOnlyMany`，安装 Job 首次写入后所有 Pod 只读挂载
- Playwright PVC 是全局共享的，所有用户命名空间的 Pod 共用同一份浏览器二进制
- 如需更新浏览器版本（更换 OpenClaw 镜像后），删除旧 Job 重新执行步骤 2
- Chromium 在容器内以 `--no-sandbox` 模式运行（Pod 非 root 用户）

## 数据库初始化

```bash
# 创建数据库
psql "postgresql://user:pass@host:port/postgres" -c "CREATE DATABASE clawunit;"

# 执行初始化 SQL
psql "postgresql://user:pass@host:port/clawunit" -f manifest/deploy/migrations/001_init_schema.sql
```

## System Skills PVC

系统技能 PVC 也是全局共享资源（与 Playwright PVC 类似），通过 `k8s.systemSkillsPVC` 配置，所有用户的 Pod 以只读方式挂载到 `/skills/system`。**只需创建一次**：

```bash
NAMESPACE="clawunit"   # 或你的 k8s.namespace 配置值

kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: clawunit-system-skills   # 对应 k8s.systemSkillsPVC 配置值
  namespace: $NAMESPACE
  labels:
    app: clawunit
spec:
  accessModes:
    - ReadOnlyMany
  storageClassName: longhorn
  resources:
    requests:
      storage: 100Mi
EOF
```
