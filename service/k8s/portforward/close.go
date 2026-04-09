package portforward

import "fmt"

// Close 关闭指定 Pod 的 port-forward
func Close(podName, podNamespace string) {
	key := fmt.Sprintf("%s/%s", podNamespace, podName)

	mu.Lock()
	defer mu.Unlock()

	if e, ok := cache[key]; ok {
		close(e.stopChan)
		delete(cache, key)
	}
}
