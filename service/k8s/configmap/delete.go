package configmap

import (
	"context"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Delete 删除实例的 ConfigMap
func Delete(ctx context.Context, ownerUpn string, instanceID int64) {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	cmName := nameFor(instanceID)

	if err := c.Clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, cmName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		g.Log().Warningf(ctx, "删除 ConfigMap %s 失败: %v", cmName, err)
	}
}
