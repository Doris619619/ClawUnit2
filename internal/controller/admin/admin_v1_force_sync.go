package admin

import (
	"context"

	v1 "clawunit.cuhksz/api/admin/v1"
	"clawunit.cuhksz/internal/service/sync"
)

func (c *ControllerV1) ForceSync(ctx context.Context, _ *v1.ForceSyncReq) (res *v1.ForceSyncRes, err error) {
	sync.ForceSync(ctx)

	return &v1.ForceSyncRes{}, nil
}
