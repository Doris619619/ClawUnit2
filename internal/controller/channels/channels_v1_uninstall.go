package channels

import (
	"context"
	"strings"

	v1 "clawunit.cuhksz/api/channels/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/configmap"
	"clawunit.cuhksz/internal/service/k8s/exec"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Uninstall 卸载渠道插件
func (c *ControllerV1) Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	podName, namespace, err := getInstancePod(ctx, ownerUpn, req.Id)
	if err != nil {
		return nil, err
	}

	pluginID := req.PluginSpec
	if idx := strings.LastIndex(pluginID, "/"); idx >= 0 {
		pluginID = pluginID[idx+1:]
	}

	// 删除 PVC 上的插件文件
	_, _ = exec.InPod(ctx, namespace, podName, "gateway", []string{
		"rm", "-rf", "/home/node/.openclaw/extensions/" + pluginID,
	})

	// 更新 K8s ConfigMap 移除插件配置
	if patchErr := configmap.PatchPlugin(ctx, ownerUpn, req.Id, pluginID, false); patchErr != nil {
		return nil, gerror.Wrapf(patchErr, "更新配置失败")
	}

	return &v1.UninstallRes{}, nil
}
