package pvc

import (
	"context"
	"time"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// monitorBinding 异步监控 PVC 绑定状态（被 EnsureInstance 使用）
func monitorBinding(ctx context.Context, namespace, pvcName string) {
	c := k8s.GetClient()
	timeout := 30 * time.Second
	ticker := time.NewTicker(2 * time.Second)

	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			g.Log().Warningf(ctx, "PVC %s/%s 绑定超时", namespace, pvcName)

			return
		case <-ticker.C:
			pvc, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
			if err != nil {
				g.Log().Errorf(ctx, "监控 PVC %s 绑定状态失败: %v", pvcName, err)

				return
			}

			if pvc.Status.Phase == corev1.ClaimBound {
				g.Log().Infof(ctx, "PVC %s/%s 已绑定到 %s", namespace, pvcName, pvc.Spec.VolumeName)

				return
			}
		}
	}
}
