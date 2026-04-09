// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package instances

import (
	"context"

	"clawunit.cuhksz/api/instances/v1"
)

type IInstancesV1 interface {
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
	GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	UpdateConfig(ctx context.Context, req *v1.UpdateConfigReq) (res *v1.UpdateConfigRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	ListOrphanPVCs(ctx context.Context, req *v1.ListOrphanPVCsReq) (res *v1.ListOrphanPVCsRes, err error)
	DeletePVC(ctx context.Context, req *v1.DeletePVCReq) (res *v1.DeletePVCRes, err error)
	Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error)
	Stop(ctx context.Context, req *v1.StopReq) (res *v1.StopRes, err error)
	Restart(ctx context.Context, req *v1.RestartReq) (res *v1.RestartRes, err error)
	GetStatus(ctx context.Context, req *v1.GetStatusReq) (res *v1.GetStatusRes, err error)
	GetQuota(ctx context.Context, req *v1.GetQuotaReq) (res *v1.GetQuotaRes, err error)
}
