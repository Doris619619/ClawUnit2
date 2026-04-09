package configmap

import (
	"context"
	"fmt"
	"strconv"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create 为指定实例写入或更新 openclaw.json ConfigMap。
//
// 已存在的 ConfigMap 会被 Update 覆盖，幂等。configJSON 通常由
// OpenClawConfig 生成；ConfigMap 名称是 clawunit-{instanceID}-config。
func Create(ctx context.Context, ownerUpn string, instanceID int64, instanceName, configJSON string) error {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	cmName := nameFor(instanceID)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "clawunit",
				"instance-id":   strconv.FormatInt(instanceID, 10),
				"instance-name": instanceName,
				"managed-by":    "clawunit",
			},
		},
		Data: map[string]string{
			"openclaw.json": configJSON,
		},
	}

	_, err := c.Clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			_, err = c.Clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("更新 ConfigMap %s 失败: %w", cmName, err)
			}

			return nil
		}

		return fmt.Errorf("创建 ConfigMap %s 失败: %w", cmName, err)
	}

	g.Log().Infof(ctx, "已创建 ConfigMap: %s/%s", namespace, cmName)

	return nil
}
