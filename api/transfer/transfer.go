// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package transfer

import (
	"context"

	"clawunit.cuhksz/api/transfer/v1"
)

type ITransferV1 interface {
	Export(ctx context.Context, req *v1.ExportReq) (res *v1.ExportRes, err error)
	Import(ctx context.Context, req *v1.ImportReq) (res *v1.ImportRes, err error)
}
