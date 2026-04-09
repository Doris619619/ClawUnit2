package lifecycle

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/service/k8s/pod"
)

// Status 是实例的当前状态快照。
//
// Status 来自 instances 表（由 sync 服务推进），PodPhase / Ready
// 是从 K8s API 实时读到的 Pod Phase 和 Ready condition —— 当 Pod
// 已被删除或不可达时这两个字段为零值。
type Status struct {
	Status   string
	PodPhase string
	Ready    bool
}

// GetStatus 返回实例的状态快照。
//
// 它同时读取 DB 中持久化的 Status 字段和 K8s 中的实时 Pod Phase。
// 如果实例不存在或不属于 ownerUpn，返回 error。Pod 已被删除时
// PodPhase 为空字符串、Ready 为 false，但 Status 字段仍然返回 DB 值
// （此时通常是 stopped/error）。
func GetStatus(ctx context.Context, ownerUpn string, instanceID int64) (*Status, error) {
	record, err := dao.Instances.Ctx(ctx).
		Where("id", instanceID).
		Where("owner_upn", ownerUpn).
		One()
	if err != nil {
		return nil, fmt.Errorf("查询实例失败: %w", err)
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("实例 %d 不存在", instanceID)
	}

	status := &Status{
		Status: record["status"].String(),
	}

	// Pod 不存在不视为错误 —— 实例可能处于 stopped 状态
	podStatus, err := pod.GetStatus(ctx, ownerUpn, instanceID)
	if err == nil {
		status.PodPhase = string(podStatus.Phase)

		for _, cond := range podStatus.Conditions {
			if cond.Type == "Ready" {
				status.Ready = cond.Status == "True"
			}
		}
	}

	return status, nil
}
