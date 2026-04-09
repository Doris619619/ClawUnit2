package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// GetStatus 获取实例 Pod 的状态
func GetStatus(ctx context.Context, ownerUpn string, instanceID int64) (*corev1.PodStatus, error) {
	pod, err := Get(ctx, ownerUpn, instanceID)
	if err != nil {
		return nil, err
	}

	return &pod.Status, nil
}
