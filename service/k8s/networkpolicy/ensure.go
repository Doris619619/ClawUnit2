package networkpolicy

import (
	"context"
	"fmt"
	"strconv"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Ensure 创建或更新实例的出口网络策略
// 限制出口流量：仅允许 DNS + 系统命名空间 + 外网 80/443
func Ensure(ctx context.Context, ownerUpn string, instanceID int64, instanceName string) error {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	policyName := c.GetNetworkPolicyName(instanceID, instanceName)
	instanceLabel := strconv.FormatInt(instanceID, 10)

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "clawunit",
				"instance-id":   instanceLabel,
				"instance-name": instanceName,
				"owner-upn":     k8s.UpnHash(ownerUpn),
				"managed-by":    "clawunit",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":         "clawunit",
					"instance-id": instanceLabel,
					"managed-by":  "clawunit",
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				// 允许 DNS
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "kube-system",
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: new(corev1.ProtocolUDP), Port: new(intstr.FromInt(53))},
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(53))},
					},
				},
				// 允许访问系统命名空间（open-platform 等服务）
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": c.GetSystemNamespace(),
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(80))},
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(443))},
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(8032))},
					},
				},
				// 允许访问外网 HTTP/HTTPS（LLM API、apt 仓库、web_fetch 等）
				{
					Ports: []networkingv1.NetworkPolicyPort{
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(80))},
						{Protocol: new(corev1.ProtocolTCP), Port: new(intstr.FromInt(443))},
					},
				},
			},
		},
	}

	existing, err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, policyName, metav1.GetOptions{})
	if err == nil {
		policy.ResourceVersion = existing.ResourceVersion

		if _, err = c.Clientset.NetworkingV1().NetworkPolicies(namespace).Update(ctx, policy, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("更新 NetworkPolicy %s 失败: %w", policyName, err)
		}

		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("查询 NetworkPolicy %s 失败: %w", policyName, err)
	}

	if _, err := c.Clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("创建 NetworkPolicy %s 失败: %w", policyName, err)
	}

	g.Log().Infof(ctx, "已创建 NetworkPolicy: %s/%s", namespace, policyName)

	return nil
}
