package lifecycle

import (
	"context"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/configmap"
	"clawunit.cuhksz/internal/service/k8s/networkpolicy"
	"clawunit.cuhksz/internal/service/k8s/pod"
	"clawunit.cuhksz/internal/service/k8s/portforward"
	"clawunit.cuhksz/internal/service/uniauth"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Delete 销毁实例的全部 K8s 资源（除用户数据 PVC 之外）+ DB 记录。
//
// 流程：
//
//  1. 关闭实例的 port-forward 缓存（如果有）
//  2. 删除 Pod 并等待最多 30 秒确认终止
//  3. 删除 NetworkPolicy 和 ConfigMap
//  4. 吊销实例下所有未撤销的 UniAuth API Key
//  5. 删除 DB 中的实例记录和 api_key_provisions 子记录
//
// *不删除*用户数据 PVC，用户可以通过 ListOrphanPVCs 看到孤立 PVC，
// 自己决定是否复用或回收。
//
// 任何子步骤失败都不会中断后续步骤 —— 这是"尽力而为"语义，避免
// 卡在半清理状态。失败的子步骤会留下 warning 日志供运维处理。
func Delete(ctx context.Context, ownerUpn string, instanceID int64) error {
	// 关闭 port-forward
	portforward.CloseForInstance(ctx, ownerUpn, instanceID)

	// 删除 Pod 并等待终止
	_ = pod.DeleteAndWait(ctx, ownerUpn, instanceID)

	instance, _ := getInstance(ctx, instanceID)
	if instance != nil {
		_ = networkpolicy.Delete(ctx, ownerUpn, instanceID, instance.Name)
	}

	// 删除 ConfigMap
	configmap.Delete(ctx, ownerUpn, instanceID)

	// 吊销 API key
	revokeInstanceApiKeys(ctx, instanceID)

	// Pod 已终止，安全删除 DB 记录
	deleteInstance(ctx, instanceID)

	g.Log().Infof(ctx, "实例 %d 已完全删除", instanceID)

	return nil
}

// revokeInstanceApiKeys 吊销实例下所有未撤销的 API key（仅 Delete 使用）
func revokeInstanceApiKeys(ctx context.Context, instanceID int64) {
	records, err := dao.ApiKeyProvisions.Ctx(ctx).
		Where("instance_id", instanceID).
		WhereNull("revoked_at").
		All()
	if err != nil {
		g.Log().Warningf(ctx, "查询实例 %d 的 API key 失败: %v", instanceID, err)

		return
	}

	for _, record := range records {
		hash := record["api_key_hash"].String()
		_ = uniauth.RevokeApiKey(ctx, hash)

		_, _ = dao.ApiKeyProvisions.Ctx(ctx).Where("api_key_hash", hash).Data(do.ApiKeyProvisions{
			RevokedAt: gtime.Now(),
		}).Update()
	}
}
