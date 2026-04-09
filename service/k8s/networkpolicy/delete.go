package networkpolicy

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Delete 删除实例的网络策略
func Delete(ctx context.Context, ownerUpn string, instanceID int64, instanceName string) error {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	policyName := c.GetNetworkPolicyName(instanceID, instanceName)

	if err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policyName, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("删除 NetworkPolicy %s 失败: %w", policyName, err)
	}

	return nil
}
