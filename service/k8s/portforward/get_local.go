package portforward

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// GetLocal 获取 Pod 的本地转发端口
// out-of-cluster 时自动建立 port-forward，in-cluster 时返回 0（调用方用 Pod IP）
func GetLocal(podName, podNamespace string, remotePort int32) (int32, error) {
	c := k8s.GetClient()
	if c == nil {
		return 0, errors.New("K8s 客户端未初始化")
	}

	// in-cluster 模式不需要 port-forward
	if c.Config.Host == "" {
		return 0, nil
	}

	key := fmt.Sprintf("%s/%s", podNamespace, podName)

	mu.Lock()
	defer mu.Unlock()

	// 复用已有的 port-forward
	if e, ok := cache[key]; ok {
		// 检查端口是否还活着
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", e.localPort))
		if err == nil {
			conn.Close()

			return e.localPort, nil
		}
		// 端口不通了，清理旧的
		close(e.stopChan)
		delete(cache, key)
	}

	// 分配一个随机本地端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("分配本地端口失败: %w", err)
	}

	// net.Listen("tcp") 保证 Addr 是 *TCPAddr，端口是 16 位无符号整数，永远适合 int32
	localPort := int32(listener.Addr().(*net.TCPAddr).Port) //nolint:forcetypeassert,gosec
	listener.Close()

	// 建立 port-forward
	reqURL := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(podNamespace).
		Name(podName).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(c.Config)
	if err != nil {
		return 0, fmt.Errorf("创建 SPDY transport 失败: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", reqURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("创建 port-forward 失败: %w", err)
	}

	errChan := make(chan error, 1)

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-readyChan:
		// port-forward 就绪
	case err := <-errChan:
		return 0, fmt.Errorf("port-forward 启动失败: %w", err)
	}

	cache[key] = &entry{localPort: localPort, stopChan: stopChan}
	g.Log().Infof(context.Background(), "已建立 port-forward: 127.0.0.1:%d -> %s/%s:%d", localPort, podNamespace, podName, remotePort)

	return localPort, nil
}
