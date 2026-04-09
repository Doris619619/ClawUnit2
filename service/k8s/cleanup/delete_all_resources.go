package cleanup

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteAllResources 删除实例的所有 K8s 资源
// 注意：不删除用户 PVC（跨实例共享）
func DeleteAllResources(ctx context.Context, ownerUpn string, instanceID int64) error {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	instanceLabel := strconv.FormatInt(instanceID, 10)
	selector := fmt.Sprintf("instance-id=%s,managed-by=clawunit", instanceLabel)

	// 删除 Pod
	//nolint:dupl // K8s 不同资源类型用不同 clientset 路径，结构相似但无法泛化
	deleteBySelector(ctx, "Pod", func() error {
		pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}

		for i := range pods.Items {
			if err := c.Clientset.CoreV1().Pods(namespace).Delete(ctx, pods.Items[i].Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				g.Log().Warningf(ctx, "删除 Pod %s 失败: %v", pods.Items[i].Name, err)
			}
		}

		return nil
	})

	// 删除 Service
	//nolint:dupl // 同上
	deleteBySelector(ctx, "Service", func() error {
		services, err := c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}

		for i := range services.Items {
			if err := c.Clientset.CoreV1().Services(namespace).Delete(ctx, services.Items[i].Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				g.Log().Warningf(ctx, "删除 Service %s 失败: %v", services.Items[i].Name, err)
			}
		}

		return nil
	})

	// 删除 NetworkPolicy
	//nolint:dupl // 同上
	deleteBySelector(ctx, "NetworkPolicy", func() error {
		policies, err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}

		for i := range policies.Items {
			if err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, policies.Items[i].Name, metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				g.Log().Warningf(ctx, "删除 NetworkPolicy %s 失败: %v", policies.Items[i].Name, err)
			}
		}

		return nil
	})

	// 等待 Pod 实际删除
	WaitForPodDeletion(ctx, namespace, selector, 30*time.Second)

	return nil
}
