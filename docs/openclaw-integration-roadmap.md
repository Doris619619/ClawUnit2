# OpenClaw 集成路线图

基于 OpenClaw 文档的全面分析，以下是 ClawUnit 尚未适配的功能和集成建议。

## 已完成

- [x] 基础生命周期管理（create/start/stop/restart/delete）
- [x] 手动/自动 API 配置模式
- [x] ConfigMap 配置挂载 + init container
- [x] 自定义 LLM provider（apiKey + baseUrl + modelId）
- [x] 聊天代理（SSE 流式）通过 proxy 转发
- [x] out-of-cluster 自动 port-forward
- [x] 全功能版镜像（latest）支持
- [x] NetworkPolicy 外网 HTTPS 出口
- [x] 前端 MarkdownRenderer 渲染 AI 回复
- [x] WebSocket 聊天桥接（`wschat.go`：connect 帧 + RPC 白名单 + session 隔离 + Origin header + hello-ok 转发）
- [x] Playwright 浏览器 PVC（全局共享 ReadOnlyMany PVC，通过 `k8s.playwrightPVC` 配置，Pod 只读挂载到 `/home/node/.cache/ms-playwright`）
- [x] 配置热更新（搜索引擎/工具权限/系统提示词/modelId，通过 UpdateConfig API 重新生成 ConfigMap）
- [x] per-instance allowPrivateNetwork 选项
- [x] 三级 UI 页面（卡片列表→配置页→对话页）
- [x] 状态颜色区分 + 云面板风格配置页

## 高优先级（建议尽快实现）

### 1. 多模型支持
- **功能**: 一个实例配置多个 LLM provider，运行时切换
- **实现**: ConfigMap 里配多个 provider + x-openclaw-model header 选择
- **前端**: 聊天输入框加模型切换下拉

## 中优先级

### 2. Skills 技能系统
- **功能**: 上传 SKILL.md 文件，教 OpenClaw 特定任务
- **后端已有**: skills CRUD API
- **待实现**: 通过 K8s exec 将 SKILL.md 写入 Pod 的 workspace/skills/ 目录
- **前端**: skills 列表 + 上传/删除 UI

### 3. Web Search 配置
- **支持的搜索引擎**: Brave/Perplexity/Tavily/DuckDuckGo/Exa/SearXNG
- **实现**: 在 openclaw.json 配置搜索 provider + API key
- **前端**: 搜索引擎选择 + API key 配置

### 4. 会话管理
- **功能**: 保存/恢复聊天历史
- **OpenClaw 支持**: session transcript 自动保存在 state dir
- **实现**: 通过 K8s exec 读取 session JSONL 文件
- **前端**: 历史会话列表

### 5. 实例日志查看
- **功能**: 查看 OpenClaw gateway 运行日志
- **实现**: K8s logs API 或读取 /tmp/openclaw/*.log
- **前端**: 日志面板（可折叠）

### 6. Exec 审批 / 浏览器工具
- **功能**: 命令执行审批（允许/拒绝）和浏览器自动化工具
- **状态**: WS 白名单已包含 `exec.approval.resolve` 和 `plugin.approval.resolve`
- **阻塞**: OpenClaw 2026.4.2 版本存在 loopback pairing bug，审批通过后 followup 消息无法正确路由。浏览器工具同样受此 bug 影响
- **临时方案**: exec 使用 `security: "full"` 自动批准
- **待跟进**: OpenClaw 修复后切换为交互式审批模式

## 低优先级（后续迭代）

### 7. Cron 定时任务
- **功能**: 定时触发 agent 执行任务（如每日总结、定期检查）
- **实现**: 在 openclaw.json 启用 cron + 通过 API 管理 jobs

### 8. Hooks 事件触发
- **功能**: 在会话创建/重置/停止时自动执行操作
- **用途**: 审计日志、自动初始化、清理

### 9. 沙箱模式（K8s 适配）
- **当前限制**: Docker socket 在 K8s 里不安全
- **方案**: 使用 SSH sandbox 或 OpenShell 远程沙箱
- **用途**: 安全的学生代码执行

### 10. 通讯渠道集成
- **支持**: Discord/WhatsApp/Telegram/Slack/Email
- **用途**: 学生通过 IM 直接与 AI 助手对话
- **实现**: 在 openclaw.json 配置 channel + 认证

### 11. 多 Agent 路由
- **功能**: 一个实例内运行多个 agent，按上下文路由
- **用途**: 不同课程不同 agent 风格

### 12. Sub-Agent 编排
- **功能**: 主 agent 可以 spawn 子任务并行执行
- **用途**: 复杂研究任务、批量处理

### 13. 插件系统
- **功能**: 安装第三方 OpenClaw 插件扩展功能
- **用途**: 大学特定集成（Canvas/Blackboard LMS）

## 技术债

- [x] 实例删除时应等待 Pod 完全终止再删 DB 记录
- [ ] 创建实例时 modelId 应有默认值（如 "gpt-4o"）
- [x] auto 模式的 UniAuth API Key 创建路径需要修正（前端传入 quotaPool）
- [x] 前端聊天组件应支持流式渲染
- [x] 前端应支持 `stream: true` 的 SSE 事件流
