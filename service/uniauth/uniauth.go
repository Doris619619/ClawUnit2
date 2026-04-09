package uniauth

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// 包内共享的 HTTP 客户端和 baseURL
var (
	client  = g.Client().ContentJson().SetRetry(3, time.Second*3)
	baseURL = g.Cfg().MustGet(context.Background(), "client.uniauth.url").String()
)
