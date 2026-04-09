package uniauth

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// CheckPermission 调用 UniAuth 校验 (subject, object, action) 三元组。
//
// 返回 (false, nil) 表示明确无权限（不是错误）；
// 返回 (_, err) 表示 UniAuth 服务异常 —— caller 应该把这种情况
// 视为暂时性失败，不要降级为"无权限"。
//
// ClawUnit 当前只用到两组：
//
//	CheckPermission(ctx, upn, "clawunit", "access")  // 普通用户
//	CheckPermission(ctx, upn, "clawunit", "admin")   // 管理员
func CheckPermission(ctx context.Context, upn, obj, act string) (bool, error) {
	resp, err := client.Post(ctx, baseURL+"/auth/check", g.Map{
		"sub": upn,
		"obj": obj,
		"act": act,
	})
	if err != nil {
		return false, err
	}

	defer resp.Close()

	result, err := gjson.DecodeToJson(resp.ReadAll())
	if err != nil {
		return false, err
	}

	if result.Get("code").Int() != 0 {
		return false, nil
	}

	return result.Get("data.allow").Bool(), nil
}
