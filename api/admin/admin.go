// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package admin

import (
	"context"

	"clawunit.cuhksz/api/admin/v1"
)

type IAdminV1 interface {
	AdminListInstances(ctx context.Context, req *v1.AdminListInstancesReq) (res *v1.AdminListInstancesRes, err error)
	AdminUpdateQuota(ctx context.Context, req *v1.AdminUpdateQuotaReq) (res *v1.AdminUpdateQuotaRes, err error)
	SyncStatus(ctx context.Context, req *v1.SyncStatusReq) (res *v1.SyncStatusRes, err error)
	ForceSync(ctx context.Context, req *v1.ForceSyncReq) (res *v1.ForceSyncRes, err error)
}
