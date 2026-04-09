// Package proxy 把前端 HTTP/WebSocket 请求转发到 OpenClaw Pod。
//
// 公开函数列表：
//
//	WsChat        WebSocket 聊天双向桥接（前端 ws ↔ Pod ws）
//	LatestMedia   读取 Pod PVC 上的最新媒体文件（浏览器截图等）
//
// 两个函数都是 ghttp.Request handler，由 internal/cmd/cmd.go 注册。
//
// # 为什么需要代理
//
// 前端浏览器没法直接访问 K8s 集群内的 Pod，所以 ClawUnit 充当反向代理：
//
//	浏览器 ──ws──> ClawUnit ──ws──> OpenClaw Pod (LAN binding, port 18789)
//	浏览器 ──http─> ClawUnit ──exec──> OpenClaw Pod (read PVC files)
//
// 同时这一层也是安全边界：
//
//   - WS 桥做 RPC 方法白名单过滤，阻止用户调用管理类方法（config.set 等）
//   - 媒体代理做路径校验、文件名校验、图片 magic byte 校验，防止任意文件读取
//
// # WsChat 的 RPC 白名单
//
// 前端发往 Pod 的 RPC 只允许这些方法（见 allowedMethods）：
//
//	chat.send / chat.abort / chat.history
//	sessions.list / sessions.delete / sessions.patch
//	exec.approval.resolve / plugin.approval.resolve
//
// 其他方法（包括 config.set/apply、skills.install、connect 等管理操作）
// 都被拦截，避免恶意前端通过 WS 越权配置 OpenClaw。心跳和事件类帧
// （type != "req"）放行。
//
// # WsChat 连接握手
//
// ClawUnit 连接 OpenClaw Pod 时要发送 connect 帧伪装成 control-ui 客户端：
//
//	{
//	  "type":"req", "method":"connect",
//	  "params": {
//	    "client": {"id":"openclaw-control-ui","mode":"webchat"},
//	    "caps": ["tool-events"],
//	    "auth": {"token":"<gateway_token>"}
//	  }
//	}
//
// 加上 Origin: http://127.0.0.1 的 header（OpenClaw Gateway 校验 origin），
// gateway 配置里 dangerouslyDisableDeviceAuth=true 跳过设备认证。
//
// 握手成功后双向透传所有帧，直到任一端断开连接。
//
// # LatestMedia 的安全检查
//
// 媒体读取走 K8s exec（cat 命令），存在任意文件读取风险，必须做完
// 整路径校验。流程：
//
//  1. 校验 mediaType 在白名单内（browser/generated/camera/canvas）
//  2. 校验实例归属（owner_upn 必须匹配）
//  3. ls -t 列出 mediaType 目录下最新文件
//  4. 校验文件名格式：UUID + 图片扩展名（regex）
//  5. readlink -f 解析符号链接，校验真实路径仍在 mediaBase 下（防穿越）
//  6. cat 读取内容
//  7. 校验图片 magic bytes（PNG/JPEG/WEBP）
//  8. 写审计日志，返回内容
//
// 任何一步失败都返回相应 HTTP 状态码（403/404/422），并写审计 warning。
//
// # 双向透传的并发模型
//
// WsChat 用 sync.WaitGroup 启动两个 goroutine（Pod→前端 / 前端→Pod），
// 任一方向断开就关闭整个连接。注意 done channel 是用来让 frontend→Pod
// 那边能够及时退出的 —— pod conn 关闭后 ReadMessage 会立刻返回 error，
// 但 frontConn.ReadMessage 是阻塞的，需要 done 信号唤醒它检查状态。
package proxy
