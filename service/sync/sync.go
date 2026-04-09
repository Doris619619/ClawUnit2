package sync

import (
	"context"
	"sync/atomic"
	"time"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/pod"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
)

const syncInterval = 5 * time.Second

// 包内共享：上次同步时间，原子读写
var lastRun atomic.Value // string

// syncInstanceStatus 同步所有活跃实例的 Pod 状态（被 Start 和 ForceSync 复用）
func syncInstanceStatus(ctx context.Context) {
	lastRun.Store(time.Now().Format(time.RFC3339))

	// 查询活跃实例（含 error 状态，可能会恢复）
	records, err := dao.Instances.Ctx(ctx).
		WhereIn("status", []string{"creating", "running", "error"}).
		All()
	if err != nil {
		g.Log().Errorf(ctx, "查询活跃实例失败: %v", err)

		return
	}

	for _, record := range records {
		instanceID := record["id"].Int64()
		ownerUpn := record["owner_upn"].String()
		currentStatus := record["status"].String()

		podStatus, err := pod.GetStatus(ctx, ownerUpn, instanceID)
		if err != nil {
			// Pod 不存在，标记为 stopped 或 error
			switch currentStatus {
			case "running":
				g.Log().Warningf(ctx, "实例 %d 的 Pod 已消失，标记为 stopped", instanceID)
				updateStatus(ctx, instanceID, "stopped")
			case "creating":
				// creating 状态下 Pod 消失可能是创建失败
				g.Log().Warningf(ctx, "实例 %d 创建中但 Pod 消失，标记为 error", instanceID)
				updateStatus(ctx, instanceID, "error")
			}

			continue
		}

		newStatus := mapPodPhaseToStatus(podStatus, currentStatus)
		if newStatus != currentStatus {
			g.Log().Infof(ctx, "实例 %d 状态变更: %s → %s (Pod phase: %s)", instanceID, currentStatus, newStatus, podStatus.Phase)
			updateStatus(ctx, instanceID, newStatus)
		}

		// running 状态下始终同步 Pod IP（可能之前为空）
		if newStatus == "running" && podStatus.PodIP != "" {
			currentIP := record["pod_ip"].String()
			if currentIP != podStatus.PodIP {
				updatePodIP(ctx, instanceID, podStatus.PodIP)
			}
		}
	}
}

// mapPodPhaseToStatus 将 K8s Pod 状态映射为实例状态
func mapPodPhaseToStatus(podStatus *corev1.PodStatus, currentStatus string) string {
	switch podStatus.Phase {
	case corev1.PodRunning:
		// 检查所有容器是否 Ready
		for _, cond := range podStatus.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return "running"
			}
		}

		return currentStatus // 还没 Ready，保持当前状态
	case corev1.PodPending:
		return currentStatus // 保持 creating
	case corev1.PodSucceeded:
		return "stopped"
	case corev1.PodFailed:
		return "error"
	case corev1.PodUnknown:
		return currentStatus
	default:
		return currentStatus
	}
}

func updateStatus(ctx context.Context, instanceID int64, status string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		Status: status,
	}).Update(); err != nil {
		g.Log().Errorf(ctx, "更新实例 %d 状态失败: %v", instanceID, err)
	}
}

func updatePodIP(ctx context.Context, instanceID int64, podIP string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		PodIp: podIP,
	}).Update(); err != nil {
		g.Log().Errorf(ctx, "更新实例 %d Pod IP 失败: %v", instanceID, err)
	}
}
