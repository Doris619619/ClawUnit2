package pod

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get 通过 label 查找实例的 Pod
func Get(ctx context.Context, ownerUpn string, instanceID int64) (*corev1.Pod, error) {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	selector := fmt.Sprintf("instance-id=%d,managed-by=clawunit", instanceID)

	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("查询 Pod 失败: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("实例 %d 的 Pod 未找到", instanceID)
	}

	return &pods.Items[0], nil
}
