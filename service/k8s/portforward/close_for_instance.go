package portforward

import (
	"context"

	"clawunit.cuhksz/internal/service/k8s"
	"clawunit.cuhksz/internal/service/k8s/pod"
)

// CloseForInstance 通过实例 ID 查找 Pod 并关闭其 port-forward
func CloseForInstance(ctx context.Context, ownerUpn string, instanceID int64) {
	p, err := pod.Get(ctx, ownerUpn, instanceID)
	if err != nil {
		return
	}

	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	Close(p.Name, namespace)
}
