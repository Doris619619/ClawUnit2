package configmap

import "fmt"

// nameFor 实例 ConfigMap 的名称（被 Create / PatchPlugin / Delete 复用）
func nameFor(instanceID int64) string {
	return fmt.Sprintf("clawunit-%d-config", instanceID)
}
