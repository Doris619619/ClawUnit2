package uniauth

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// CreateApiKeyReq 是 CreateApiKey 函数的入参。
//
// QuotaPool 必填，绑定到 Open Platform 配额池决定计费来源。
// Permissions 留空表示采用配额池默认权限集。MaxLifetimeCost > 0
// 时给 key 设置消费上限（USD）。
type CreateApiKeyReq struct {
	Nickname        string   `json:"nickname"`
	QuotaPool       string   `json:"quotaPool"`
	Permissions     []string `json:"permissions"`
	MaxLifetimeCost float64  `json:"maxLifetimeCost"`
}

// CreateApiKeyRes 是 CreateApiKey 的返回值。
//
// RawApiKey 是完整 API Key 文本，*只在创建响应里出现一次*，必须立刻
// 注入到目标 Pod 的环境变量；ApiKeyHash 用于后续吊销/查询；MaskedKey
// 是打码版本（"sk-***"），可以安全存数据库或日志。
type CreateApiKeyRes struct {
	RawApiKey  string `json:"rawApiKey"`
	ApiKeyHash string `json:"apiKeyHash"`
	MaskedKey  string `json:"maskedApiKey"`
}

// CreateApiKey 通过 UniAuth 创建 API Key
func CreateApiKey(ctx context.Context, ownerUpn string, req CreateApiKeyReq) (*CreateApiKeyRes, error) {
	resp, err := client.Post(ctx, baseURL+"/openPlatform/apikey", g.Map{
		"ownerUpn":        ownerUpn,
		"nickname":        req.Nickname,
		"quotaPool":       req.QuotaPool,
		"permissions":     req.Permissions,
		"maxLifetimeCost": req.MaxLifetimeCost,
	})
	if err != nil {
		return nil, err
	}

	defer resp.Close()

	result, err := gjson.DecodeToJson(resp.ReadAll())
	if err != nil {
		return nil, err
	}

	if result.Get("code").Int() != 0 {
		return nil, fmt.Errorf("UniAuth 创建 API Key 失败: %s", result.Get("message").String())
	}

	return &CreateApiKeyRes{
		RawApiKey:  result.Get("data.rawApiKey").String(),
		ApiKeyHash: result.Get("data.apiKeyHash").String(),
		MaskedKey:  result.Get("data.maskedApiKey").String(),
	}, nil
}
