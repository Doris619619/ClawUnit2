package pvc

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureInstance 确保实例数据 PVC 存在（挂载到 /home/node/.openclaw/，持久化 session/memory/workspace）
// 如果 existingPVC 不为空，验证其存在后直接返回（复用旧数据）
func EnsureInstance(ctx context.Context, ownerUpn, instanceName, existingPVC, storageSize string) (string, error) {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	storageClass := c.StorageClass

	// 复用已有 PVC
	if existingPVC != "" {
		existing, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, existingPVC, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("指定的 PVC %s 不存在: %w", existingPVC, err)
		}

		// 验证归属
		if existing.Labels["managed-by"] != "clawunit" || existing.Labels["owner-upn"] != k8s.UpnHash(ownerUpn) {
			return "", fmt.Errorf("PVC %s 不属于当前用户", existingPVC)
		}

		g.Log().Infof(ctx, "复用已有 PVC: %s/%s", namespace, existingPVC)

		return existingPVC, nil
	}

	// 创建新 PVC
	pvcName := GetInstanceName(instanceName, ownerUpn)

	existing, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err == nil {
		return existing.Name, nil
	}

	if !errors.IsNotFound(err) {
		return "", fmt.Errorf("查询 PVC %s 失败: %w", pvcName, err)
	}

	storageSizeQty := resource.MustParse(storageSize)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "clawunit",
				"owner-upn":     k8s.UpnHash(ownerUpn),
				"instance-name": instanceName,
				"managed-by":    "clawunit",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageSizeQty,
				},
			},
			StorageClassName: &storageClass,
		},
	}

	created, err := c.Clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return pvcName, nil
		}

		return "", fmt.Errorf("创建 PVC %s 失败: %w", pvcName, err)
	}

	g.Log().Infof(ctx, "已创建实例 PVC: %s/%s (%s)", namespace, created.Name, storageSize)

	// 后台监控不绑定请求 context，避免请求结束后 goroutine 被取消
	go monitorBinding(context.Background(), namespace, pvcName) //nolint:gosec,contextcheck // intentional: monitor outlives request

	return pvcName, nil
}
