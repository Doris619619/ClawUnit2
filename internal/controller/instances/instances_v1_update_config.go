package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/configmap"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) UpdateConfig(ctx context.Context, req *v1.UpdateConfigReq) (res *v1.UpdateConfigRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	record, err := dao.Instances.Ctx(ctx).Where("id", req.Id).Where("owner_upn", ownerUpn).One()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	if record.IsEmpty() {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "实例 %d 不存在", req.Id)
	}

	// 更新 DB 中的 LLM 配置（apiKey/baseUrl/modelId 变更需要重建 Pod 环境变量才能生效）
	updateDO := do.Instances{}

	if req.ApiKey != nil {
		updateDO.ApiKey = *req.ApiKey
	}

	if req.BaseUrl != nil {
		updateDO.BaseUrl = *req.BaseUrl
	}

	if updateDO.ApiKey != nil || updateDO.BaseUrl != nil {
		if _, err = dao.Instances.Ctx(ctx).Where("id", req.Id).Data(updateDO).Update(); err != nil {
			return nil, gerror.Wrapf(err, "更新实例记录失败")
		}
	}

	// 构建新的 OpenClaw 配置
	modelID := record["model_id"].String()
	if req.ModelID != nil {
		modelID = *req.ModelID
	}

	opts := configmap.OpenClawConfigOptions{
		ModelID: modelID,
	}

	if req.ToolProfile != nil {
		opts.ToolProfile = *req.ToolProfile
	}

	if req.SearchProvider != nil {
		opts.SearchProvider = *req.SearchProvider
	}

	if req.SearchApiKey != nil {
		opts.SearchApiKey = *req.SearchApiKey
	}

	if req.SystemPrompt != nil {
		opts.SystemPrompt = *req.SystemPrompt
	}

	if req.AllowPrivateNetwork != nil {
		opts.AllowPrivateNetwork = *req.AllowPrivateNetwork
	} else {
		opts.AllowPrivateNetwork = record["allow_private_network"].Bool()
	}

	configJSON := configmap.OpenClawConfig(opts)

	// 更新 ConfigMap（OpenClaw 热加载 agent/model/tools 配置）
	if err = configmap.Create(ctx, ownerUpn, req.Id, record["name"].String(), configJSON); err != nil {
		return nil, gerror.Wrapf(err, "更新配置失败")
	}

	return &v1.UpdateConfigRes{}, nil
}
