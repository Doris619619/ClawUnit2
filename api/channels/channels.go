// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package channels

import (
	"context"

	"clawunit.cuhksz/api/channels/v1"
)

type IChannelsV1 interface {
	Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error)
	Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error)
	RestartGateway(ctx context.Context, req *v1.RestartGatewayReq) (res *v1.RestartGatewayRes, err error)
	Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error)
}
