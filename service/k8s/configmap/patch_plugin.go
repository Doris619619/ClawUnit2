package configmap

import (
	"context"
	"encoding/json"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PatchPlugin 把指定 plugin 加到（add=true）或从（add=false）实例
// ConfigMap 的 openclaw.json 中。
//
// 修改的字段：plugins.allow（白名单）和 plugins.entries（启用配置）。
// 注意：ConfigMap 改完后 *不会* 立即生效，因为 OpenClaw 运行时读的
// 是 PVC 上的 openclaw.json，需要 channels.RestartGateway 让 Gateway
// 进程重新加载配置文件。
func PatchPlugin(ctx context.Context, ownerUpn string, instanceID int64, pluginID string, add bool) error {
	c := k8s.GetClient()
	namespace := c.GetUserNamespace(ownerUpn)
	cmName := nameFor(instanceID)

	cm, err := c.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("读取 ConfigMap 失败: %w", err)
	}

	configJSON := cm.Data["openclaw.json"]

	var config g.Map
	if err = json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("解析 ConfigMap JSON 失败: %w", err)
	}

	// 获取或创建 plugins 部分
	plugins, _ := config["plugins"].(g.Map)
	if plugins == nil {
		plugins = g.Map{}
		config["plugins"] = plugins
	}

	// 处理 plugins.allow
	allowRaw, _ := plugins["allow"].([]any)
	allowSet := map[string]bool{}

	for _, v := range allowRaw {
		if s, ok := v.(string); ok {
			allowSet[s] = true
		}
	}

	// 处理 plugins.entries
	entries, _ := plugins["entries"].(g.Map)
	if entries == nil {
		entries = g.Map{}
		plugins["entries"] = entries
	}

	if add {
		allowSet[pluginID] = true
		entries[pluginID] = g.Map{"enabled": true}
	} else {
		delete(allowSet, pluginID)
		delete(entries, pluginID)
	}

	// 重建 allow 数组
	var newAllow []any
	for k := range allowSet {
		newAllow = append(newAllow, k)
	}

	plugins["allow"] = newAllow

	// 序列化回 JSON
	newJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 ConfigMap JSON 失败: %w", err)
	}

	cm.Data["openclaw.json"] = string(newJSON)

	_, err = c.Clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 ConfigMap 失败: %w", err)
	}

	g.Log().Infof(ctx, "ConfigMap %s 插件配置已更新: %s (add=%v)", cmName, pluginID, add)

	return nil
}
