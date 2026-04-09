package sync

// LastRunTime 返回上次同步成功的 RFC3339 时间戳。
//
// 同步服务还没跑过任何一轮时返回空字符串。这个值由 admin 仪表盘
// 用来判断同步服务是否健康（差距 > 30 秒就告警）。
func LastRunTime() string {
	v := lastRun.Load()
	if v == nil {
		return ""
	}

	s, _ := v.(string)

	return s
}
