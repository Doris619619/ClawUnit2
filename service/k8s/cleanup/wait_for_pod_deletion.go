package cleanup

import (
	"context"
	"time"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WaitForPodDeletion 等待 Pod 删除完成
func WaitForPodDeletion(ctx context.Context, namespace, selector string, timeout time.Duration) {
	c := k8s.GetClient()
	ticker := time.NewTicker(2 * time.Second)

	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			g.Log().Warningf(ctx, "等待 Pod 删除超时: namespace=%s, selector=%s", namespace, selector)

			return
		case <-ticker.C:
			pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
			if err != nil || len(pods.Items) == 0 {
				return
			}
		}
	}
}
