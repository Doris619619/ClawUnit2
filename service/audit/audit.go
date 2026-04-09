package audit

import (
	"context"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Log 把一次操作记录写入 audit_events 表。
//
// 写失败只记 error 日志，绝不返回错误 —— 审计模块挂掉不应该阻塞
// 业务流程。caller 应该在所有副作用完成 *之后* 调用 Log，确保
// details 反映最终结果。
//
// details 中不要塞 API key 原文、密码、access token 等敏感数据。
func Log(ctx context.Context, actorUpn, action, resourceType string, resourceID int64, details g.Map) {
	if _, err := dao.AuditEvents.Ctx(ctx).Data(do.AuditEvents{
		ActorUpn:     actorUpn,
		Action:       action,
		ResourceType: resourceType,
		ResourceId:   resourceID,
		Details:      gjson.New(details),
	}).Insert(); err != nil {
		g.Log().Errorf(ctx, "审计日志写入失败: %v", err)
	}
}
