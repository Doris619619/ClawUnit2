package instances

import (
	"context"

	v1 "clawunit.cuhksz/api/instances/v1"
	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/middlewares"
	"clawunit.cuhksz/internal/service/k8s/pvc"

	"github.com/gogf/gf/v2/errors/gerror"
	corev1 "k8s.io/api/core/v1"
)

func (c *ControllerV1) ListOrphanPVCs(ctx context.Context, _ *v1.ListOrphanPVCsReq) (res *v1.ListOrphanPVCsRes, err error) {
	ownerUpn := middlewares.OwnerFromCtx(ctx)

	// 查出该用户所有活跃实例的 pvc_name
	records, err := dao.Instances.Ctx(ctx).
		Where("owner_upn", ownerUpn).
		Where("status !=", "deleting").
		All()
	if err != nil {
		return nil, gerror.Wrapf(err, "查询实例失败")
	}

	var activePVCs []string

	for _, r := range records {
		if name := r["pvc_name"].String(); name != "" {
			activePVCs = append(activePVCs, name)
		}
	}

	orphans, err := pvc.ListOrphan(ctx, ownerUpn, activePVCs)
	if err != nil {
		return nil, gerror.Wrapf(err, "列出可用 PVC 失败")
	}

	var list []v1.OrphanPVC

	for _, p := range orphans {
		size := ""
		if qty, ok := p.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			size = qty.String()
		}

		// 从 label 获取原实例名，没有的话从 PVC 名推断
		instName := p.Labels["instance-name"]
		if instName == "" {
			instName = p.Name
		}

		list = append(list, v1.OrphanPVC{
			Name:         p.Name,
			InstanceName: instName,
			Size:         size,
			CreatedAt:    p.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return &v1.ListOrphanPVCsRes{List: list}, nil
}
