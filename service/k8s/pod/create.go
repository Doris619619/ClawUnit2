package pod

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"clawunit.cuhksz/internal/service/k8s"

	"github.com/gogf/gf/v2/frame/g"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config 是 Create 函数的入参，描述创建一个 OpenClaw Pod 所需的全部信息。
//
// CPU / Memory 是 K8s 资源格式字符串（"4"、"8Gi"），不要传 int。
// PVCName 是已经存在的实例数据 PVC 名称（由 pvc.EnsureInstance 创建）。
// ContainerPort/GatewayPort 都是 OpenClaw Gateway 监听端口，
// 默认 18789；零值会被填充为 18789。
//
// ExtraEnv 是要追加到容器的环境变量，典型用法是注入 OPENCLAW_GATEWAY_TOKEN
// 和 LLM 凭证（CUSTOM_API_KEY / CUSTOM_BASE_URL）。
type Config struct {
	ExtraEnv      map[string]string
	Memory        string
	OwnerUpn      string
	Image         string
	MountPath     string
	CPU           string
	PVCName       string
	InstanceName  string
	InstanceID    int64
	GPUCount      int32
	ContainerPort int32
	GatewayPort   int32
	GPUEnabled    bool
}

// Create 创建一个 OpenClaw 实例 Pod。
//
// Pod 包含一个 init container（busybox 复制 ConfigMap 到 PVC）和一个
// gateway container（运行 OpenClaw）。挂载点：
//
//   - /home/node/.openclaw       用户数据 PVC（RW）
//   - /skills/system             系统技能 PVC（RO）
//   - /home/node/.cache/ms-playwright  Playwright 浏览器 PVC（RO）
//   - /dev/shm                   memory-backed EmptyDir，给 Chrome 用
//
// 如果同名 Pod 已存在（通常是上次创建残留），会自动删除并重建一次。
func Create(ctx context.Context, cfg Config) (*corev1.Pod, error) {
	c := k8s.GetClient()
	podName := c.GetPodName(cfg.InstanceID, cfg.InstanceName)
	namespace := c.GetUserNamespace(cfg.OwnerUpn)
	instanceLabel := strconv.FormatInt(cfg.InstanceID, 10)

	if cfg.ContainerPort == 0 {
		cfg.ContainerPort = 18789
	}

	if cfg.GatewayPort == 0 {
		cfg.GatewayPort = 18789
	}

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cfg.CPU),
			corev1.ResourceMemory: resource.MustParse(cfg.Memory),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cfg.CPU),
			corev1.ResourceMemory: resource.MustParse(cfg.Memory),
		},
	}

	if cfg.GPUEnabled && cfg.GPUCount > 0 {
		gpuQty := resource.MustParse(strconv.Itoa(int(cfg.GPUCount)))
		resources.Limits["nvidia.com/gpu"] = gpuQty
		resources.Requests["nvidia.com/gpu"] = gpuQty
	}

	configMapName := fmt.Sprintf("clawunit-%d-config", cfg.InstanceID)
	nodeUser := int64(1000)

	// 环境变量
	envVars := make([]corev1.EnvVar, 0, 6+len(cfg.ExtraEnv))

	envVars = append(envVars,
		corev1.EnvVar{Name: "HOME", Value: "/home/node"},
		corev1.EnvVar{Name: "OPENCLAW_CONFIG_DIR", Value: "/home/node/.openclaw"},
		corev1.EnvVar{Name: "NODE_ENV", Value: "production"},
		corev1.EnvVar{Name: "INSTANCE_ID", Value: instanceLabel},
		corev1.EnvVar{Name: "OWNER_UPN", Value: cfg.OwnerUpn},
		corev1.EnvVar{Name: "PLAYWRIGHT_BROWSERS_PATH", Value: "/home/node/.cache/ms-playwright"},
	)
	for key, value := range cfg.ExtraEnv {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: value})
	}

	// 健康检查：用 exec 而非 httpGet（gateway 绑定 lan 但用 node 命令更可靠）
	healthCmd := func(path string) []string {
		return []string{
			"node", "-e",
			fmt.Sprintf("require('http').get('http://127.0.0.1:%d%s', r => process.exit(r.statusCode < 400 ? 0 : 1)).on('error', () => process.exit(1))", cfg.ContainerPort, path),
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "clawunit",
				"instance-id":   instanceLabel,
				"instance-name": cfg.InstanceName,
				"owner-upn":     k8s.UpnHash(cfg.OwnerUpn),
				"managed-by":    "clawunit",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			SecurityContext: &corev1.PodSecurityContext{
				FSGroup: &nodeUser,
			},
			InitContainers: []corev1.Container{
				{
					Name:            "init-config",
					Image:           "docker.gitfetch.dev/library/busybox:1.37",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"sh", "-c", "test -f /home/node/.openclaw/openclaw.json || cp /config/openclaw.json /home/node/.openclaw/openclaw.json; mkdir -p /home/node/.openclaw/workspace /home/node/.openclaw/memory /home/node/.openclaw/agents"},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser:  &nodeUser,
						RunAsGroup: &nodeUser,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("32Mi"),
							corev1.ResourceCPU:    resource.MustParse("50m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("64Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
					},
					VolumeMounts: initVolumeMounts(),
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "gateway",
					Image: cfg.Image,
					// Playwright 二进制来自 PVC，只需安装系统依赖（libnss3 等）
					Command: []string{"sh", "-c", "DEBIAN_FRONTEND=noninteractive node /app/node_modules/playwright-core/cli.js install-deps chromium > /dev/null 2>&1; exec runuser -u node -- node /app/dist/index.js gateway run"},
					Ports: []corev1.ContainerPort{
						{ContainerPort: cfg.ContainerPort, Name: "gateway", Protocol: corev1.ProtocolTCP},
					},
					StartupProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{Command: healthCmd("/healthz")},
						},
						FailureThreshold: 120,
						PeriodSeconds:    5,
						TimeoutSeconds:   5,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{Command: healthCmd("/readyz")},
						},
						InitialDelaySeconds: 15,
						PeriodSeconds:       10,
						TimeoutSeconds:      5,
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							Exec: &corev1.ExecAction{Command: healthCmd("/healthz")},
						},
						InitialDelaySeconds: 60,
						PeriodSeconds:       30,
						TimeoutSeconds:      10,
					},
					Resources:    resources,
					VolumeMounts: gatewayVolumeMounts(),
					Env:          envVars,
					SecurityContext: &corev1.SecurityContext{
						RunAsUser:                new(int64), // 0 = root，用于 apt-get 安装依赖
						AllowPrivilegeEscalation: new(false),
						ReadOnlyRootFilesystem:   new(false),
					},
				},
			},
			Volumes: podVolumes(configMapName, c.SystemSkillsPVC, c.PlaywrightPVC, cfg.PVCName),
		},
	}

	created, err := c.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// 删除已有 Pod 并重建
			if delErr := c.Clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{}); delErr != nil && !errors.IsNotFound(delErr) {
				return nil, fmt.Errorf("删除已存在的 Pod %s 失败: %w", podName, delErr)
			}

			time.Sleep(5 * time.Second)

			created, err = c.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
			if err != nil {
				return nil, fmt.Errorf("重建 Pod %s 失败: %w", podName, err)
			}

			return created, nil
		}

		return nil, fmt.Errorf("创建 Pod %s 失败: %w", podName, err)
	}

	g.Log().Infof(ctx, "已创建 Pod: %s/%s", namespace, podName)

	return created, nil
}
