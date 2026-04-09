// Package configmap 管理 OpenClaw 实例的 ConfigMap 资源。
//
// 公开函数列表：
//
//	OpenClawConfig   生成 openclaw.json 配置文本
//	Create           创建/更新实例 ConfigMap
//	PatchPlugin      在已有 ConfigMap 的 openclaw.json 中加/删插件配置
//	Delete           删除实例 ConfigMap（不存在不报错）
//
// # ConfigMap 与 PVC 的关系
//
// ConfigMap 只是"初始配置模板"，真正运行时的 openclaw.json 在 PVC 上：
//
//  1. 实例第一次启动时，init container 从 ConfigMap 读取 openclaw.json
//     拷贝到 PVC 的 /home/node/.openclaw/openclaw.json
//  2. OpenClaw 运行后所有的运行时改动都直接写在 PVC 上的那份 json，
//     ConfigMap 保持不变
//  3. UpdateConfig API 重新生成 ConfigMap 后，需要重启 Pod（init
//     container 会再次复制覆盖 PVC），否则改动不生效
//
// 例外：PatchPlugin 是直接更新 ConfigMap，但 OpenClaw 通过 inotify 监听
// 配置文件变化（Gateway 重启时重新加载），所以装完插件需要调
// channels.RestartGateway 让进程重新读 PVC 上的 openclaw.json。
//
// # OpenClawConfig 的生成逻辑
//
// 不用模板引擎是历史原因：早期 openclaw.json 结构简单，几个 Sprintf 就
// 够；后来加了搜索引擎、agents 自定义、多 LLM provider，现在已经有
// 不少嵌套结构。如果再扩展建议换成 text/template 或者 struct 序列化，
// 但目前的 fmt.Sprintf 风格保持一致更易读。
//
// 生成的 JSON 关键字段：
//
//	gateway.bind = "lan"                       绑 Pod 内 LAN，不暴露
//	gateway.auth.token = ${OPENCLAW_GATEWAY_TOKEN}  环境变量替换，每实例一份
//	gateway.controlUi.dangerouslyDisableDeviceAuth = true  绕过设备认证（ws 桥需要）
//	models.providers.sglang.api = "openai-completions"     用 sglang 而不是 custom
//	browser.ssrfPolicy.dangerouslyAllowPrivateNetwork = true  浏览器需连本地 CDP
//
// # 为什么必须用 sglang provider 而不是 custom
//
// 后端 LLM（GLM-5/Qwen 等）由 sglang 部署，返回 reasoning_content 字段。
// OpenClaw 的 custom provider 不解析这个字段，会得到空回复。改成 sglang
// provider + reasoning:true 就能正常解析。这条踩过坑的经验保存在
// memory/project_sglang_provider.md。
package configmap
