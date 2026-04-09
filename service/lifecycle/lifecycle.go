package lifecycle

import (
	"context"
	"fmt"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/uniauth"

	"github.com/gogf/gf/v2/frame/g"
)

// instanceInfo 实例信息（getInstance 返回值，被 Start/Stop/Delete 复用）
type instanceInfo struct {
	Name          string
	Description   string
	Status        string
	Image         string
	MountPath     string
	ApiMode       string
	ApiKey        string
	BaseUrl       string
	GatewayToken  string
	PVCName       string
	GPUCount      int32
	ContainerPort int32
	GPUEnabled    bool
}

// getInstance 查询实例完整信息（被 Start/Stop/Delete 复用）
func getInstance(ctx context.Context, instanceID int64) (*instanceInfo, error) {
	record, err := dao.Instances.Ctx(ctx).Where("id", instanceID).One()
	if err != nil {
		return nil, fmt.Errorf("查询实例失败: %w", err)
	}

	if record.IsEmpty() {
		return nil, fmt.Errorf("实例 %d 不存在", instanceID)
	}

	return &instanceInfo{
		Name:          record["name"].String(),
		Description:   record["description"].String(),
		Status:        record["status"].String(),
		Image:         record["image"].String(),
		MountPath:     record["mount_path"].String(),
		ApiMode:       record["api_mode"].String(),
		ApiKey:        record["api_key"].String(),
		BaseUrl:       record["base_url"].String(),
		GatewayToken:  record["gateway_token"].String(),
		PVCName:       record["pvc_name"].String(),
		GPUCount:      record["gpu_count"].Int32(),
		ContainerPort: record["container_port"].Int32(),
		GPUEnabled:    record["gpu_enabled"].Bool(),
	}, nil
}

// deleteInstance 删除 DB 实例记录及关联的 API key 记录（被 Create 回滚和 Delete 复用）
func deleteInstance(ctx context.Context, instanceID int64) {
	if _, err := dao.ApiKeyProvisions.Ctx(ctx).Where("instance_id", instanceID).Delete(); err != nil {
		g.Log().Warningf(ctx, "删除实例 %d 的 API key 记录失败: %v", instanceID, err)
	}

	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Delete(); err != nil {
		g.Log().Warningf(ctx, "删除实例记录 %d 失败: %v", instanceID, err)
	}
}

// provisionApiKey 通过 UniAuth 创建 API Key 并记录到 DB（被 Create 和 ensureApiKey 复用）
func provisionApiKey(ctx context.Context, ownerUpn string, instanceID int64, quotaPool string) (apiKeyHash, rawApiKey string, err error) {
	if quotaPool == "" {
		quotaPool = g.Cfg().MustGet(ctx, "instance.quotaPool", "clawunit").String()
	}

	res, err := uniauth.CreateApiKey(ctx, ownerUpn, uniauth.CreateApiKeyReq{
		Nickname:  fmt.Sprintf("clawunit-instance-%d", instanceID),
		QuotaPool: quotaPool,
	})
	if err != nil {
		return "", "", err
	}

	// 记录 API key 配置
	if _, err := dao.ApiKeyProvisions.Ctx(ctx).Data(do.ApiKeyProvisions{
		InstanceId: instanceID,
		OwnerUpn:   ownerUpn,
		ApiKeyHash: res.ApiKeyHash,
		QuotaPool:  quotaPool,
	}).Insert(); err != nil {
		g.Log().Warningf(ctx, "记录 API key 配置失败: %v", err)
	}

	return res.ApiKeyHash, res.RawApiKey, nil
}
