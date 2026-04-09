# Browser 工具超时排查记录

## 问题现象

OpenClaw 实例的 browser 工具调用（如 `open url`）始终返回 `timed out`，错误在 ~4 秒内触发。
Agent 使用浏览器工具时 LLM 收到的错误为 `[tools] browser failed: timed out`。

## 排查时间线

### 阶段 1：怀疑 Playwright 安装问题

**假设**: Playwright 浏览器二进制文件缺失或系统依赖不全，Chrome 无法启动。

**尝试**:
- 手动 `npx playwright install --with-deps chromium`
- 确认 Chrome 二进制文件存在于 `/home/node/.cache/ms-playwright/chromium-1208/chrome-linux64/chrome`
- 测试手动启动 Chrome: `chrome --headless=new --no-sandbox --remote-debugging-port=18800` — **成功**
- CDP endpoint `http://127.0.0.1:18800/json/version` 返回正常

**结论**: Chrome 本身能正常启动，不是 Playwright 安装问题。

### 阶段 2：怀疑 Longhorn 存储 I/O 慢

**假设**: Playwright PVC (ReadOnlyMany) 挂载在 Longhorn 上，I/O 延迟导致 Chrome 启动超时。

**尝试**:
- 去掉 Playwright PVC，改为 Pod 启动时 `install --with-deps chromium` 直接装到本地
- 测试 Chrome 启动速度

**结论**: 改为本地安装后 Chrome 启动很快，但 browser 工具仍然超时。**不是存储问题**。

### 阶段 3：怀疑 Chrome zombie 进程

**假设**: Chrome 子进程变成 zombie，占用资源导致后续操作超时。

**尝试**:
- 安装 `tini` 作为 PID 1，确保正确回收子进程
- 启动命令改为 `exec tini -- runuser -u node -- node /app/dist/index.js gateway run`

**结论**: zombie 进程没了，但 browser 工具仍然超时。**zombie 是副作用，不是根因**。

### 阶段 4：怀疑 OpenClaw 启动 Chrome 太慢

**假设**: OpenClaw browser service 启动 Chrome 的流程太慢，触发了工具超时。

**尝试**:
- 改用 `attachOnly` 模式：Pod 启动时先手动启动 Chrome，OpenClaw 通过 `cdpUrl` 直连
- 启动命令加入 pre-start Chrome + 等待 CDP ready 的逻辑
- 配置 `"openclaw": { "cdpUrl": "http://127.0.0.1:18800", "attachOnly": true }`

**结论**: browser service status API 显示 `running: true, cdpReady: true`，但 **调用 tabs/open 仍然超时**。

### 阶段 5：深入分析 browser service 代码

手动调用 browser service API 排查：

```bash
# 查状态 — 成功
curl http://127.0.0.1:18791/?profile=openclaw
# → running: true, cdpReady: true

# 打开标签页 — 超时
curl -X POST http://127.0.0.1:18791/tabs/open -d '{"url":"https://www.baidu.com"}'
# → 超时
```

逐步分析调用链：
1. `tabs/open` → 内部调用 `fetchHttpJson` → 请求 `http://127.0.0.1:18800`（Chrome CDP）
2. 查看 `fetchHttpJson` 源码：默认 timeout 30s，不是 4s
3. 查看 `core-api` 中的 timeout 配置：`tabs/open` 的 `timeoutMs: 15e3`
4. **关键线索**: 超时时间固定 ~4 秒，与代码中任何显式 timeout 值都不匹配

### 阶段 6：发现 SSRF 策略拦截（根因）

在 `pw-ai` 模块的 `fetchHttpJson` 中发现 SSRF 检查逻辑：

```javascript
// isPrivateIp 检查 127.0.0.1 是否为私有地址
if (!options?.ssrfPolicy?.dangerouslyAllowPrivateNetwork && isPrivateIp(hostname)) {
    throw new Error("SSRF: private network request blocked");
}
```

OpenClaw 的 SSRF 策略默认拒绝对私有网络地址的请求。browser service 连接 Chrome CDP 时使用 `http://127.0.0.1:18800`，这是一个私有地址，**被 SSRF 策略静默拦截**。

错误被包装后变成了通用的 `timed out`，掩盖了真正的原因。

## 根因

**`ssrfPolicy.dangerouslyAllowPrivateNetwork` 默认为 `false`**，导致 browser service 对 `127.0.0.1`（Chrome CDP）的所有 HTTP 请求被 SSRF 策略拦截。

这解释了所有现象：
- Chrome 能正常启动（SSRF 不影响进程启动）
- CDP endpoint 能手动 curl 访问（curl 不经过 SSRF 检查）
- `attachOnly` + `cdpReady: true` 但 `tabs/open` 仍超时（连接 established 但后续请求被拦）
- 超时时间 ~4 秒（SSRF 拦截后的错误处理 + 重试耗时，不是真正的网络超时）

## 修复

在 OpenClaw 配置的 `browser` 部分加上：

```json
"ssrfPolicy": {
  "dangerouslyAllowPrivateNetwork": true
}
```

一行配置修复。所有之前的变通方案（attachOnly、pre-start Chrome、tini）都不需要。

## 最终配置

```json
"browser": {
  "enabled": true,
  "executablePath": "/home/node/.cache/ms-playwright/chromium-1208/chrome-linux64/chrome",
  "defaultProfile": "openclaw",
  "headless": true,
  "noSandbox": true,
  "profiles": {
    "openclaw": { "cdpPort": 18800 }
  },
  "ssrfPolicy": {
    "dangerouslyAllowPrivateNetwork": true
  }
}
```

启动命令只需安装 Playwright Chromium + 启动 gateway，OpenClaw 自行管理 Chrome 生命周期。

## 教训

1. **SSRF 拦截会伪装成超时** — 错误信息不透明，容易误导排查方向
2. **手动 curl 能通不代表应用层也能通** — 应用内部可能有额外的安全策略层
3. **从最简配置开始** — 应该先确认最基本的 `tabs/open` API 调用是否正常，再叠加复杂配置
4. **查看应用源码中的安全策略** — browser service 的 `fetchHttpJson` 包含了不明显的 SSRF 检查逻辑
