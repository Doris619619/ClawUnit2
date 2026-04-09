package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/service/k8s/portforward"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}
	reqCounter atomic.Int64

	// allowedMethods 用户可通过 WS 调用的 RPC 方法白名单
	// 管理类方法（config.set/apply、skills.install、sessions.delete 等）不允许
	allowedMethods = map[string]bool{
		"chat.send":               true,
		"chat.abort":              true,
		"chat.history":            true,
		"sessions.list":           true,
		"sessions.delete":         true,
		"sessions.patch":          true,
		"exec.approval.resolve":   true,
		"plugin.approval.resolve": true,
	}
)

// connectFrame 构建 OpenClaw connect 握手帧
func connectFrame(token string) []byte {
	id := fmt.Sprintf("c%d", reqCounter.Add(1))
	frame := map[string]any{
		"type":   "req",
		"id":     id,
		"method": "connect",
		"params": map[string]any{
			"minProtocol": 3,
			"maxProtocol": 3,
			"client": map[string]any{
				"id":       "openclaw-control-ui",
				"version":  "1.0.0",
				"platform": "web",
				"mode":     "webchat",
			},
			"role":        "operator",
			"scopes":      []string{"operator.read", "operator.write", "operator.admin", "operator.approvals", "operator.pairing"},
			"caps":        []string{"tool-events"},
			"commands":    []string{},
			"permissions": map[string]any{},
			"auth":        map[string]any{"token": token},
			"locale":      "en-US",
			"userAgent":   "clawunit.cuhksz/1.0.0",
		},
	}

	data, _ := json.Marshal(frame)

	return data
}

// WsChat 是浏览器到 OpenClaw Pod 的 WebSocket 双向桥接 handler。
//
// 前端通过 ws://<clawunit>/api/chat/v1/ws?userId=<upn>&instanceId=<id>
// 连接，ClawUnit 负责：
//
//  1. 校验 userId + instanceId 归属
//  2. 建立到 Pod 的 WebSocket 连接（in-cluster 直连 Pod IP，
//     out-of-cluster 走 portforward）
//  3. 完成 connect 握手伪装成 control-ui 客户端
//  4. 双向透传帧，前端→Pod 方向应用 RPC 白名单过滤
//
// 注意 WS handler 不走 UniResMiddleware（会破坏升级响应），也不走
// InjectIdentity（WebSocket 升级时不能可靠传 header），所以身份和实例
// 校验都在本函数内完成。
func WsChat(r *ghttp.Request) {
	ctx := r.Context()

	// WS 不走中间件，从 header 或 query param 获取用户身份
	ownerUpn := r.Header.Get("X-User-ID")
	if ownerUpn == "" {
		ownerUpn = r.GetQuery("userId").String()
	}

	if ownerUpn == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)

		return
	}

	instanceID := r.GetQuery("instanceId").Int64()
	if instanceID == 0 {
		r.Response.WriteStatus(http.StatusBadRequest)

		return
	}

	// 查询实例
	record, err := dao.Instances.Ctx(ctx).
		Where("id", instanceID).
		Where("owner_upn", ownerUpn).
		One()
	if err != nil || record.IsEmpty() {
		r.Response.WriteStatus(http.StatusNotFound)

		return
	}

	if record["status"].String() != "running" {
		r.Response.WriteStatus(http.StatusBadRequest)

		return
	}

	gatewayToken := record["gateway_token"].String()
	podName := record["pod_name"].String()
	podNamespace := record["pod_namespace"].String()
	podIP := record["pod_ip"].String()
	gatewayPort := g.Cfg().MustGet(ctx, "instance.gatewayPort", 18789).Int32()

	// 构建 Pod WS 地址
	podWsURL := buildPodWsURL(podName, podNamespace, podIP, gatewayPort)
	if podWsURL == "" {
		r.Response.WriteStatus(http.StatusBadGateway)

		return
	}

	// 升级前端连接为 WebSocket
	frontConn, err := upgrader.Upgrade(r.Response.Writer, r.Request, nil)
	if err != nil {
		g.Log().Errorf(ctx, "前端 WebSocket 升级失败: %v", err)

		return
	}
	defer frontConn.Close()

	// 连接 OpenClaw Pod
	podConn, helloOk, err := dialPod(ctx, podWsURL, gatewayToken)
	if err != nil {
		g.Log().Errorf(ctx, "连接 OpenClaw Pod 失败: %v", err)
		_ = frontConn.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, `{"type":"error","message":"%s"}`, err.Error()))

		return
	}
	defer podConn.Close()

	// 把 hello-ok 转发给前端
	if len(helloOk) > 0 {
		_ = frontConn.WriteMessage(websocket.TextMessage, helloOk)
	}

	g.Log().Infof(ctx, "WebSocket 桥接建立: instance=%d, user=%s", instanceID, ownerUpn)

	// 双向透传
	var wg sync.WaitGroup

	done := make(chan struct{})

	// Pod → 前端

	wg.Go(func() {
		for {
			msgType, msg, err := podConn.ReadMessage()
			if err != nil {
				g.Log().Debugf(ctx, "Pod WS 读取错误: %v", err)

				break
			}

			g.Log().Debugf(ctx, "Pod→前端 转发: %s", string(msg[:min(len(msg), 5000)]))

			if err := frontConn.WriteMessage(msgType, msg); err != nil {
				g.Log().Debugf(ctx, "前端 WS 写入错误: %v", err)

				break
			}
		}

		close(done)
	})

	// 前端 → Pod

	wg.Go(func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			msgType, msg, err := frontConn.ReadMessage()
			if err != nil {
				g.Log().Debugf(ctx, "前端 WS 读取错误: %v", err)
				podConn.Close()

				return
			}

			// RPC 方法白名单过滤：只允许聊天相关方法，阻止管理类操作
			if !isAllowedFrame(msg) {
				g.Log().Warningf(ctx, "拦截非白名单 RPC: %s", string(msg[:min(len(msg), 200)]))

				continue
			}

			g.Log().Debugf(ctx, "前端→Pod 转发: %s", string(msg[:min(len(msg), 200)]))

			if err := podConn.WriteMessage(msgType, msg); err != nil {
				g.Log().Debugf(ctx, "Pod WS 写入错误: %v", err)

				return
			}
		}
	})

	wg.Wait()
	g.Log().Infof(ctx, "WebSocket 桥接关闭: instance=%d", instanceID)

	r.ExitAll()
}

// dialPod 连接到 OpenClaw Pod 并完成 connect 握手
func dialPod(ctx context.Context, wsURL, token string) (*websocket.Conn, []byte, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// 设置 Origin header，OpenClaw Control UI 要求 origin 匹配
	headers := http.Header{}
	headers.Set("Origin", "http://127.0.0.1")

	conn, httpResp, err := dialer.DialContext(ctx, wsURL, headers)
	if httpResp != nil {
		_ = httpResp.Body.Close()
	}

	if err != nil {
		return nil, nil, fmt.Errorf("WebSocket dial 失败: %w", err)
	}

	// 发送 connect 握手
	if err = conn.WriteMessage(websocket.TextMessage, connectFrame(token)); err != nil {
		conn.Close()

		return nil, nil, fmt.Errorf("发送 connect 帧失败: %w", err)
	}

	// 等待 hello-ok
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()

		return nil, nil, fmt.Errorf("等待 hello-ok 超时: %w", err)
	}

	// 验证响应
	var resp map[string]any
	if err = json.Unmarshal(msg, &resp); err != nil {
		conn.Close()

		return nil, nil, fmt.Errorf("解析 hello-ok 失败: %w", err)
	}

	g.Log().Debugf(ctx, "OpenClaw 响应: %s", string(msg[:min(len(msg), 2000)]))

	// 可能收到 connect.challenge 事件，需要读下一条消息（hello-ok）
	for resp["type"] == "event" {
		_, msg, err = conn.ReadMessage()
		if err != nil {
			conn.Close()

			return nil, nil, fmt.Errorf("等待 hello-ok 超时: %w", err)
		}

		resp = map[string]any{}

		if err := json.Unmarshal(msg, &resp); err != nil {
			conn.Close()

			return nil, nil, fmt.Errorf("解析响应失败: %w", err)
		}

		g.Log().Debugf(ctx, "OpenClaw 响应: %s", string(msg[:min(len(msg), 2000)]))
	}

	if resp["ok"] != true {
		conn.Close()

		return nil, nil, fmt.Errorf("OpenClaw connect 失败: %v", resp["error"])
	}

	// 重置读超时
	_ = conn.SetReadDeadline(time.Time{})

	return conn, msg, nil
}

// buildPodWsURL 构建 Pod WebSocket URL
func buildPodWsURL(podName, podNamespace, podIP string, port int32) string {
	// 先尝试 port-forward（out-of-cluster）
	localPort, err := portforward.GetLocal(podName, podNamespace, port)
	if err == nil && localPort > 0 {
		url := fmt.Sprintf("ws://127.0.0.1:%d", localPort)
		g.Log().Infof(context.Background(), "Pod WS URL (port-forward): %s", url)

		return url
	}

	// in-cluster 直连 Pod IP
	if podIP != "" {
		return fmt.Sprintf("ws://%s:%d", podIP, port)
	}

	return ""
}

// isAllowedFrame 检查前端发来的帧是否在白名单内
func isAllowedFrame(msg []byte) bool {
	var frame struct {
		Type   string `json:"type"`
		Method string `json:"method"`
	}

	if err := json.Unmarshal(msg, &frame); err != nil {
		return false
	}

	// 只拦截 req 类型（RPC 调用），其他类型（如心跳）放行
	if frame.Type != "req" {
		return true
	}

	return allowedMethods[frame.Method]
}
