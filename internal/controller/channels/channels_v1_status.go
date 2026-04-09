package channels

import (
	"context"
	"strings"

	v1 "clawunit.cuhksz/api/channels/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/exec"
)

// Status 查询渠道状态（含已安装插件列表）
func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	podName, namespace, err := getInstancePod(ctx, ownerUpn, req.Id)
	if err != nil {
		return nil, err
	}

	// 直接列 extensions 目录获取已安装插件
	var installedPlugins []string

	lsResult, _ := exec.InPod(ctx, namespace, podName, "gateway", []string{
		"sh", "-c", "ls /home/node/.openclaw/extensions/ 2>/dev/null",
	})
	if lsResult != nil {
		for name := range strings.SplitSeq(strings.TrimSpace(lsResult.Stdout), "\n") {
			name = strings.TrimSpace(name)
			if name != "" {
				installedPlugins = append(installedPlugins, name)
			}
		}
	}

	return &v1.StatusRes{
		InstalledPlugins: installedPlugins,
	}, nil
}
