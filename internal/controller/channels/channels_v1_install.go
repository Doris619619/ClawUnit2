package channels

import (
	"context"
	"fmt"
	"strings"

	v1 "clawunit.cuhksz/api/channels/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/configmap"
	"clawunit.cuhksz/internal/service/k8s/exec"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Install 安装渠道插件
func (c *ControllerV1) Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	podName, namespace, err := getInstancePod(ctx, ownerUpn, req.Id)
	if err != nil {
		return nil, err
	}

	// 提取插件 ID
	pluginID := req.PluginSpec
	if idx := strings.LastIndex(pluginID, "/"); idx >= 0 {
		pluginID = pluginID[idx+1:]
	}

	// 检查是否已安装
	checkResult, checkErr := exec.InPod(ctx, namespace, podName, "gateway", []string{
		"sh", "-c", fmt.Sprintf("test -f /home/node/.openclaw/extensions/%s/package.json", pluginID),
	})
	if checkErr == nil && checkResult != nil {
		return &v1.InstallRes{Output: "插件已安装"}, nil
	}

	// 安装到 PVC 持久化的 extensions 目录
	result, err := exec.InPod(ctx, namespace, podName, "gateway", []string{
		"runuser", "-u", "node", "--",
		"node", "/app/dist/index.js", "plugins", "install", req.PluginSpec,
	})
	if err != nil {
		stderr := ""
		if result != nil {
			stderr = result.Stderr
		}

		return nil, gerror.Wrapf(err, "插件安装失败: %s", stderr)
	}

	// 更新 K8s ConfigMap 加入插件配置（下次 Pod 启动时生效）
	record, _ := dao.Instances.Ctx(ctx).Where("id", req.Id).One()
	if !record.IsEmpty() {
		if patchErr := configmap.PatchPlugin(ctx, ownerUpn, req.Id, pluginID, true); patchErr != nil {
			return nil, gerror.Wrapf(patchErr, "插件已安装但更新配置失败")
		}
	}

	return &v1.InstallRes{Output: result.Stdout}, nil
}
