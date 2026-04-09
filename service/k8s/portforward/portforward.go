package portforward

import "sync"

// entry 缓存一条 port-forward 连接
type entry struct {
	stopChan  chan struct{}
	localPort int32
}

// 包级共享：port-forward 缓存（GetLocal 和 Close 共用）
var (
	cache = make(map[string]*entry) // key: "namespace/podName"
	mu    sync.Mutex
)
