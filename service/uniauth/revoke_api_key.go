package uniauth

import (
	"context"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// RevokeApiKey 通过 UniAuth 吊销已发放的 API Key。
//
// 实参是 hash 不是 raw key —— 因为 ClawUnit 没有持久化 raw key。
// 失败时 *只记 warning 不返回 error*：吊销通常出现在删除流程里，
// 失败不能阻塞 delete 完成。如果 UniAuth 真的挂了，运维需要从
// audit_events 或日志找到孤立的 hash 手动清理。
func RevokeApiKey(ctx context.Context, apiKeyHash string) error {
	resp, err := client.Delete(ctx, baseURL+"/openPlatform/apikey", g.Map{
		"apiKeyHash": apiKeyHash,
	})
	if err != nil {
		return err
	}

	defer resp.Close()

	result, err := gjson.DecodeToJson(resp.ReadAll())
	if err != nil {
		return err
	}

	if result.Get("code").Int() != 0 {
		g.Log().Warningf(ctx, "UniAuth 吊销 API Key 失败: %s", result.Get("message").String())
	}

	return nil
}
