package pvc

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListOrphan 列出用户命名空间下未被任何活跃实例绑定的 PVC
func ListOrphan(ctx context.Context, ownerUpn string, activePVCNames []string) ([]corev1.PersistentVolumeClaim, error) {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)

	pvcs, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "managed-by=clawunit,owner-upn=" + k8s.UpnHash(ownerUpn),
	})
	if err != nil {
		return nil, fmt.Errorf("列出 PVC 失败: %w", err)
	}

	activeSet := make(map[string]bool, len(activePVCNames))
	for _, name := range activePVCNames {
		activeSet[name] = true
	}

	var orphans []corev1.PersistentVolumeClaim

	for _, pvc := range pvcs.Items {
		if !activeSet[pvc.Name] {
			orphans = append(orphans, pvc)
		}
	}

	return orphans, nil
}
