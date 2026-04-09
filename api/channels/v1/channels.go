package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// InstallReq 安装渠道插件
type InstallReq struct {
	g.Meta `path:"/install" method:"post" tags:"Channels" summary:"安装渠道插件" dc:"在实例中安装指定的渠道插件"`

	PluginSpec string `json:"pluginSpec" v:"required" dc:"插件包名，如 @tencent-weixin/openclaw-weixin"`
	Id         int64  `json:"id" v:"required|min:1" dc:"实例ID"`
}

// InstallRes 安装结果
type InstallRes struct {
	Output string `json:"output" dc:"安装输出"`
}

// UninstallReq 卸载渠道插件
type UninstallReq struct {
	g.Meta `path:"/uninstall" method:"post" tags:"Channels" summary:"卸载渠道插件" dc:"从实例中卸载指定的渠道插件"`

	PluginSpec string `json:"pluginSpec" v:"required" dc:"插件包名"`
	Id         int64  `json:"id" v:"required|min:1" dc:"实例ID"`
}

// UninstallRes 卸载结果
type UninstallRes struct{}

// LoginReq 渠道扫码登录（SSE 流式）
// 注意：此接口通过 SSE 流式返回，不走标准 controller 绑定
type LoginReq struct {
	Channel string `json:"channel" v:"required" dc:"渠道名称，如 openclaw-weixin"`
	Id      int64  `json:"id" v:"required|min:1" dc:"实例ID"`
}

// RestartGatewayReq 重启 Gateway 进程（不重启容器）
type RestartGatewayReq struct {
	g.Meta `path:"/restart-gateway" method:"post" tags:"Channels" summary:"重启 Gateway" dc:"重启 OpenClaw Gateway 进程以加载新插件，不会丢失已安装的插件"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

// RestartGatewayRes 重启结果
type RestartGatewayRes struct{}

// StatusReq 查询渠道状态
type StatusReq struct {
	g.Meta `path:"/status" method:"get" tags:"Channels" summary:"渠道状态" dc:"查询实例的渠道连接状态"`

	Id int64 `json:"id" v:"required|min:1" dc:"实例ID"`
}

// StatusRes 渠道状态
type StatusRes struct {
	InstalledPlugins []string      `json:"installedPlugins" dc:"已安装的插件 ID 列表"`
	Channels         []ChannelInfo `json:"channels" dc:"渠道列表"`
}

// ChannelInfo 单个渠道信息
type ChannelInfo struct {
	Name      string `json:"name" dc:"渠道名称"`
	Account   string `json:"account" dc:"账号信息"`
	Connected bool   `json:"connected" dc:"是否已连接"`
}
