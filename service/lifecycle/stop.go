package lifecycle

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/pod"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Stop 删除实例 Pod，把状态置为 stopped。
//
// 不删除 ConfigMap、NetworkPolicy、PVC —— 这些保留是为了 Start 能
// 快速重新拉起实例。仅当前状态是 running 或 creating 时允许 Stop。
//
// Pod 删除是 best-effort，K8s 还在异步终止时本函数已经返回。如果
// 需要等 Pod 完全消失，调用方应该额外用 pod.DeleteAndWait（Delete 会用）。
func Stop(ctx context.Context, ownerUpn string, instanceID int64) error {
	instance, err := getInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	if instance.Status != "running" && instance.Status != "creating" {
		return fmt.Errorf("实例当前状态为 %s，无法停止", instance.Status)
	}

	_ = pod.Delete(ctx, ownerUpn, instanceID)

	updateInstanceStopped(ctx, instanceID)

	return nil
}

// updateInstanceStopped 更新实例为 stopped 状态并记录停止时间
func updateInstanceStopped(ctx context.Context, instanceID int64) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		Status:    "stopped",
		StoppedAt: gtime.Now(),
	}).Update(); err != nil {
		g.Log().Warningf(ctx, "更新实例 %d 停止状态失败: %v", instanceID, err)
	}
}
