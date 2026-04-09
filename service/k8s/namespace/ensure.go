package namespace

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ensure 确保用户 namespace 存在，不存在则创建
func Ensure(ctx context.Context, ownerUpn string) (*corev1.Namespace, error) {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)

	ns, err := c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return ns, nil
	}

	if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("查询 namespace %s 失败: %w", namespace, err)
	}

	newNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app":        "clawunit",
				"owner-upn":  k8s.UpnHash(ownerUpn),
				"managed-by": "clawunit",
			},
		},
	}

	created, err := c.Clientset.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 namespace %s 失败: %w", namespace, err)
	}

	g.Log().Infof(ctx, "已创建用户 namespace: %s", namespace)

	return created, nil
}
