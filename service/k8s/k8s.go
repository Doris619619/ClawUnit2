package k8s

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client 封装 Kubernetes 访问所需的全部状态。
//
// Clientset 和 Config 来自 client-go，可以同时支持 in-cluster 和
// out-of-cluster 模式。Namespace 是基础命名空间前缀（实际用户 namespace
// 是 {Namespace}-user-{upn_hash}）。StorageClass 给新建 PVC 用，
// SystemSkillsPVC 和 PlaywrightPVC 是管理员预创建的共享 PVC 名称。
//
// 通过 GetClient() 拿到全局单例，不要在子包里重复构造。
type Client struct {
	Clientset       *kubernetes.Clientset
	Config          *rest.Config
	Namespace       string
	StorageClass    string
	SystemSkillsPVC string
	PlaywrightPVC   string
}

// 包级共享：全局 Client 单例 + K8s 名称合法字符正则
var (
	globalClient        *Client
	k8sNameInvalidChars = regexp.MustCompile(`[^a-z0-9-]+`)
	k8sNameExtraDashes  = regexp.MustCompile(`-+`)
)

// GetUserNamespace 返回用户专属 namespace 名称。
//
// 格式: {Client.Namespace}-user-{upn_hash8}。upn_hash8 是 UPN 的 SHA256
// 前 4 字节（8 个 hex 字符），保证确定性、长度可控、不暴露用户邮箱。
func (c *Client) GetUserNamespace(ownerUpn string) string {
	return sanitizeK8sName(fmt.Sprintf("%s-user-%s", c.Namespace, UpnHash(ownerUpn)))
}

// GetSystemNamespace 返回平台共享的系统命名空间名称。
// 系统服务（open-platform、UniAuth 等）部署在这里，NetworkPolicy 会
// 显式放行用户实例到这个 namespace 的 80/443/8032 端口。
func (c *Client) GetSystemNamespace() string {
	return sanitizeK8sName(c.Namespace + "-system")
}

// GetPodName 返回实例 Pod 的名称：clawunit-{instanceID}-{instanceName}。
// 经过 sanitizeK8sName 规整后保证 ≤ 63 字符且符合 DNS-1123 规则。
func (c *Client) GetPodName(instanceID int64, instanceName string) string {
	return sanitizeK8sName(fmt.Sprintf("clawunit-%d-%s", instanceID, instanceName))
}

// GetUserPVCName 返回用户级共享 PVC 的名称（仅用于历史代码兼容）。
//
// 当前的 PVC 策略是 *每实例一个*，本函数返回的"每用户一个"PVC 名称
// 不再使用，但保留给即将迁移的代码做引用。新代码应该用
// pvc.GetInstanceName(instanceName, ownerUpn)。
func (c *Client) GetUserPVCName(ownerUpn string) string {
	return sanitizeK8sName("clawunit-user-" + UpnHash(ownerUpn))
}

// GetNetworkPolicyName 返回实例 NetworkPolicy 的名称：
// clawunit-{instanceID}-{instanceName}-netpol。
func (c *Client) GetNetworkPolicyName(instanceID int64, instanceName string) string {
	return sanitizeK8sName(fmt.Sprintf("clawunit-%d-%s-netpol", instanceID, instanceName))
}

// UpnHash 返回 UPN 的 SHA256 前 4 字节（8 个 hex 字符）。
//
// 用作 namespace、PVC、label 中的短标识 —— 确定性可哈希、长度可控、
// 不暴露用户邮箱明文。冲突概率极低（4 字节空间约 42 亿），且就算
// 冲突也只影响 K8s 名称，业务层用 owner_upn 字段做最终归属判断。
func UpnHash(upn string) string {
	h := sha256.Sum256([]byte(upn))

	return hex.EncodeToString(h[:4])
}

// sanitizeK8sName 把任意字符串规整为符合 K8s 命名规范的 DNS-1123 label
func sanitizeK8sName(name string) string {
	sanitized := strings.ToLower(name)
	sanitized = k8sNameInvalidChars.ReplaceAllString(sanitized, "-")
	sanitized = k8sNameExtraDashes.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	if sanitized == "" {
		return "clawunit"
	}

	if len(sanitized) > 63 {
		sanitized = strings.Trim(sanitized[:63], "-")
		if sanitized == "" {
			return "clawunit"
		}
	}

	return sanitized
}
