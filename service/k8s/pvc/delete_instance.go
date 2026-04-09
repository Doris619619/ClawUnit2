package pvc

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteInstance 删除实例数据 PVC
func DeleteInstance(ctx context.Context, ownerUpn, pvcName string) error {
	if pvcName == "" {
		return nil
	}

	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)

	if err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除 PVC %s 失败: %w", pvcName, err)
	}

	g.Log().Infof(ctx, "已删除实例 PVC: %s/%s", namespace, pvcName)

	return nil
}
