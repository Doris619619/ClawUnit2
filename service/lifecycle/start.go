package lifecycle

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/pod"
	"clawunit.cuhksz/internal/service/uniauth"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Start 把 stopped 或 error 状态的实例重新拉起来。
//
// 仅创建 Pod —— ConfigMap、NetworkPolicy、用户数据 PVC 都假设已经存在
// （首次创建时建好的，Stop 不会删除）。如果是 auto 模式实例，会调
// ensureApiKey 重新申请一份 API Key（旧 key 同时被吊销，因为 raw key
// 没有持久化保存）。
//
// 实例会先被置成 creating 状态，等 sync 服务推进到 running。
// 当前状态不是 stopped/error 时返回错误。
func Start(ctx context.Context, ownerUpn string, instanceID int64) error {
	instance, err := getInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	if instance.Status != "stopped" && instance.Status != "error" {
		return fmt.Errorf("实例当前状态为 %s，无法启动", instance.Status)
	}

	gatewayPort := g.Cfg().MustGet(ctx, "instance.gatewayPort", 18789).Int32()
	defaultCPU := g.Cfg().MustGet(ctx, "instance.defaultCPU", "200m").String()
	defaultMemory := g.Cfg().MustGet(ctx, "instance.defaultMemory", "512Mi").String()

	// 构建环境变量
	extraEnv := map[string]string{
		"OPENCLAW_GATEWAY_TOKEN": instance.GatewayToken,
	}

	if instance.ApiMode != "auto" {
		extraEnv["CUSTOM_API_KEY"] = instance.ApiKey
		extraEnv["CUSTOM_BASE_URL"] = instance.BaseUrl
	} else {
		var rawApiKey string

		rawApiKey, err = ensureApiKey(ctx, ownerUpn, instanceID)
		if err != nil {
			return fmt.Errorf("API Key 恢复失败: %w", err)
		}

		opBaseURL := g.Cfg().MustGet(ctx, "client.openPlatform.url", "http://localhost:8032").String() + "/open/v1"
		extraEnv["CUSTOM_API_KEY"] = rawApiKey
		extraEnv["CUSTOM_BASE_URL"] = opBaseURL
	}

	// 创建 Pod（ConfigMap/PVC 已在首次创建时建好）
	_, err = pod.Create(ctx, pod.Config{
		InstanceID:    instanceID,
		InstanceName:  instance.Name,
		OwnerUpn:      ownerUpn,
		Image:         instance.Image,
		MountPath:     instance.MountPath,
		CPU:           defaultCPU,
		Memory:        defaultMemory,
		GPUCount:      instance.GPUCount,
		ContainerPort: instance.ContainerPort,
		GatewayPort:   gatewayPort,
		GPUEnabled:    instance.GPUEnabled,
		PVCName:       instance.PVCName,
		ExtraEnv:      extraEnv,
	})
	if err != nil {
		return fmt.Errorf("启动 Pod 失败: %w", err)
	}

	updateInstanceStatus(ctx, instanceID, "creating")

	return nil
}

// ensureApiKey 恢复 auto 模式实例的 API Key（无明文存储则重新创建）
func ensureApiKey(ctx context.Context, ownerUpn string, instanceID int64) (string, error) {
	// 查找现有有效的 API key
	record, err := dao.ApiKeyProvisions.Ctx(ctx).
		Where("instance_id", instanceID).
		WhereNull("revoked_at").
		One()
	if err != nil {
		return "", err
	}

	// 如果没有有效 key，重新创建
	if record.IsEmpty() {
		var rawKey string

		_, rawKey, err = provisionApiKey(ctx, ownerUpn, instanceID, "")
		if err != nil {
			return "", err
		}

		return rawKey, nil
	}

	// 已有 key，但我们没有保存明文，需要重新创建
	// 先吊销旧的，再创建新的
	oldHash := record["api_key_hash"].String()
	_ = uniauth.RevokeApiKey(ctx, oldHash)

	_, _ = dao.ApiKeyProvisions.Ctx(ctx).Where("api_key_hash", oldHash).Data(do.ApiKeyProvisions{
		RevokedAt: gtime.Now(),
	}).Update()

	_, rawKey, err := provisionApiKey(ctx, ownerUpn, instanceID, "")
	if err != nil {
		return "", err
	}

	return rawKey, nil
}

// updateInstanceStatus 更新实例状态（仅 Start 使用，其他状态变更有专用 helper）
func updateInstanceStatus(ctx context.Context, instanceID int64, status string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		Status: status,
	}).Update(); err != nil {
		g.Log().Warningf(ctx, "更新实例 %d 状态失败: %v", instanceID, err)
	}
}
