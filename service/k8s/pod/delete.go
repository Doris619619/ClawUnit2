package pod

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Delete 删除实例 Pod
func Delete(ctx context.Context, ownerUpn string, instanceID int64) error {
	pod, err := Get(ctx, ownerUpn, instanceID)
	if err != nil {
		// Pod 不存在则跳过
		return nil
	}

	c := k8s.GetClient()
	if err := c.Clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除 Pod %s 失败: %w", pod.Name, err)
	}

	g.Log().Infof(ctx, "已删除 Pod: %s/%s", pod.Namespace, pod.Name)

	return nil
}
