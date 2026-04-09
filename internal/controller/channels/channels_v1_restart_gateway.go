package channels

import (
	"context"

	v1 "clawunit.cuhksz/api/channels/v1"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/exec"

	"github.com/gogf/gf/v2/errors/gerror"
)

// RestartGateway 通过 config.apply 触发 gateway 进程内重启（SIGUSR1），不丢失已安装插件
func (c *ControllerV1) RestartGateway(ctx context.Context, req *v1.RestartGatewayReq) (res *v1.RestartGatewayRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	podName, namespace, err := getInstancePod(ctx, ownerUpn, req.Id)
	if err != nil {
		return nil, err
	}

	_, err = exec.InPod(ctx, namespace, podName, "gateway", []string{
		"runuser", "-u", "node", "--",
		"node", "/app/dist/index.js", "config", "apply",
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "Gateway 重启失败")
	}

	return &v1.RestartGatewayRes{}, nil
}
