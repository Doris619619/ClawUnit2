package admin

import (
	"context"

	v1 "clawunit.cuhksz/api/admin/v1"
	"clawunit.cuhksz/internal/service/sync"
)

func (c *ControllerV1) SyncStatus(_ context.Context, _ *v1.SyncStatusReq) (res *v1.SyncStatusRes, err error) {
	return &v1.SyncStatusRes{
		Running: true,
		LastRun: sync.LastRunTime(),
	}, nil
}
