package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gogf/gf/v2/frame/g"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Initialize 根据 GoFrame 配置初始化全局 K8s 客户端单例。
//
// 配置 key：
//
//	k8s.mode             "incluster" / "outofcluster" / "auto"（默认 auto）
//	k8s.namespace        基础 namespace 前缀（默认 "clawunit"）
//	k8s.storageClass     新 PVC 的 StorageClass（默认 "standard"）
//	k8s.systemSkillsPVC  系统技能 PVC 名称（默认 "clawunit-system-skills"）
//	k8s.playwrightPVC    Playwright PVC 名称（默认 "clawunit-playwright-browsers"）
//	k8s.kubeconfig       仅 outofcluster 模式生效的 kubeconfig 路径
//
// auto 模式会先尝试 in-cluster，失败回退到 out-of-cluster。
// 应用启动时调用一次，失败返回 error；后续 GetClient 才有值。
func Initialize(ctx context.Context) error {
	mode := g.Cfg().MustGet(ctx, "k8s.mode", "auto").String()
	namespace := g.Cfg().MustGet(ctx, "k8s.namespace", "clawunit").String()
	storageClass := g.Cfg().MustGet(ctx, "k8s.storageClass", "standard").String()
	systemSkillsPVC := g.Cfg().MustGet(ctx, "k8s.systemSkillsPVC", "clawunit-system-skills").String()
	playwrightPVC := g.Cfg().MustGet(ctx, "k8s.playwrightPVC", "clawunit-playwright-browsers").String()
	kubeconfig := g.Cfg().MustGet(ctx, "k8s.kubeconfig", "").String()

	var (
		restConfig *rest.Config
		err        error
	)

	switch mode {
	case "incluster":
		restConfig, err = rest.InClusterConfig()
	case "outofcluster":
		restConfig, err = buildOutOfClusterConfig(kubeconfig)
	default: // auto
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			restConfig, err = buildOutOfClusterConfig(kubeconfig)
		}
	}

	if err != nil {
		return fmt.Errorf("K8s 客户端初始化失败: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("创建 K8s clientset 失败: %w", err)
	}

	globalClient = &Client{
		Clientset:       clientset,
		Config:          restConfig,
		Namespace:       namespace,
		StorageClass:    storageClass,
		SystemSkillsPVC: systemSkillsPVC,
		PlaywrightPVC:   playwrightPVC,
	}

	g.Log().Infof(ctx, "K8s 客户端初始化成功，mode=%s, namespace=%s", mode, namespace)

	return nil
}

// buildOutOfClusterConfig 从 kubeconfig 文件构建集群外配置（仅 Initialize 使用）
func buildOutOfClusterConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
			if kubeconfig = os.Getenv("K8S_KUBECONFIG"); kubeconfig == "" {
				kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
			}
		}
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) { //nolint:gosec // path comes from server config or env, not user input
		return nil, fmt.Errorf("kubeconfig 文件不存在: %s", kubeconfig)
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
