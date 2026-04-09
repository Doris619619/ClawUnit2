package pvc

import (
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"
)

// GetInstanceName 生成实例级 PVC 名称
func GetInstanceName(instanceName, ownerUpn string) string {
	return fmt.Sprintf("clawunit-%s-%s", instanceName, k8s.UpnHash(ownerUpn))
}
