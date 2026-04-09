package lifecycle

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/model/do"
	"clawunit.cuhksz/internal/service/k8s/configmap"
	"clawunit.cuhksz/internal/service/k8s/namespace"
	"clawunit.cuhksz/internal/service/k8s/networkpolicy"
	"clawunit.cuhksz/internal/service/k8s/pod"
	"clawunit.cuhksz/internal/service/k8s/pvc"
	"clawunit.cuhksz/internal/service/uniauth"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// CreateRequest 是 Create 函数的入参，把 controller 层的 v1.CreateReq
// 重新包装成 service 层友好的命名（不带 v1 命名空间，便于内部调用）。
//
// ApiMode 决定 LLM 接入方式：
//   - "manual"：使用 ApiKey + BaseUrl，直接打用户提供的 LLM
//   - "auto"：通过 UniAuth 创建 API Key 并绑定到 QuotaPool 配额池
//
// GPUEnabled + GPUCount 控制是否申请 nvidia.com/gpu 资源。
type CreateRequest struct {
	ApiKey              string
	Description         string
	OwnerUpn            string
	Image               string
	StorageClass        string
	ApiMode             string
	ModelID             string
	Name                string
	BaseUrl             string
	QuotaPool           string
	GPUCount            int32
	GPUEnabled          bool
	AllowPrivateNetwork bool
}

// CreateResult 是 Create 成功后返回的关键标识。
// InstanceID 用于后续 controller 返回给前端，PodName/PodNamespace 主要给
// 调用方写日志或追加 metrics 用 —— 注意此时 Pod 还在 creating 状态，
// 实际 ready 时间由 sync 服务推进。
type CreateResult struct {
	PodName      string
	PodNamespace string
	InstanceID   int64
}

// Create 创建一个新的 OpenClaw 实例。
//
// 完整流程（11 步）：
//
//  1. 读取默认配置（image / cpu / memory / 端口 / 挂载路径）
//  2. 检查用户实例数量配额
//  3. 检查实例名在用户范围内唯一
//  4. 生成随机 gateway token
//  5. 写入 instances 表（status = "creating"）
//  6. 构建环境变量；auto 模式调 UniAuth 申请 API Key
//  7. 渲染 openclaw.json 配置
//  8. 确保用户 namespace 存在
//  9. 创建 ConfigMap、NetworkPolicy
//  10. 确保用户数据 PVC 存在
//  11. 创建 Pod，更新 DB 记录的 pod_name / pod_namespace / started_at
//
// 任何步骤失败都会反向回滚之前已经创建的资源 + 吊销 API Key + 删除
// DB 记录。函数返回时 Pod 仍处于 creating 状态，由 sync 服务推进到
// running。
func Create(ctx context.Context, req CreateRequest) (*CreateResult, error) {
	req.Name = strings.TrimSpace(req.Name)

	// 读取默认配置
	if req.Image == "" {
		req.Image = g.Cfg().MustGet(ctx, "instance.defaultImage").String()
	}

	if req.StorageClass == "" {
		req.StorageClass = g.Cfg().MustGet(ctx, "k8s.storageClass", "standard").String()
	}

	containerPort := g.Cfg().MustGet(ctx, "instance.containerPort", 3001).Int32()
	gatewayPort := g.Cfg().MustGet(ctx, "instance.gatewayPort", 18789).Int32()
	mountPath := g.Cfg().MustGet(ctx, "instance.mountPath", "/home/user/data").String()

	// 资源配额从配置文件读取，前期固定不允许用户自定义
	defaultCPU := g.Cfg().MustGet(ctx, "instance.defaultCPU", "200m").String()
	defaultMemory := g.Cfg().MustGet(ctx, "instance.defaultMemory", "512Mi").String()
	defaultDisk := g.Cfg().MustGet(ctx, "instance.defaultDisk", "500Mi").String()

	// 1. 配额检查
	if err := checkQuota(ctx, req); err != nil {
		return nil, err
	}

	// 2. 检查实例名唯一性
	if err := checkNameUnique(ctx, req.OwnerUpn, req.Name); err != nil {
		return nil, err
	}

	// 3. 生成 gateway token
	gatewayToken := generateToken()

	// 4. 插入数据库记录
	instanceID, err := insertInstance(ctx, req, mountPath, containerPort, defaultCPU, defaultMemory, defaultDisk, gatewayToken)
	if err != nil {
		return nil, fmt.Errorf("创建实例记录失败: %w", err)
	}

	// 5. 构建 Pod 环境变量和 OpenClaw 配置
	extraEnv := map[string]string{
		"OPENCLAW_GATEWAY_TOKEN": gatewayToken,
	}

	var apiKeyHash string

	if req.ApiMode == "manual" {
		extraEnv["CUSTOM_API_KEY"] = req.ApiKey
		extraEnv["CUSTOM_BASE_URL"] = req.BaseUrl
	} else {
		var rawApiKey string

		apiKeyHash, rawApiKey, err = provisionApiKey(ctx, req.OwnerUpn, instanceID, req.QuotaPool)
		if err != nil {
			deleteInstance(ctx, instanceID)

			return nil, fmt.Errorf("创建 API Key 失败: %w", err)
		}

		updateInstanceApiKeyHash(ctx, instanceID, apiKeyHash)

		extraEnv["CUSTOM_API_KEY"] = rawApiKey
		extraEnv["CUSTOM_BASE_URL"] = g.Cfg().MustGet(ctx, "client.openPlatform.url", "http://localhost:8032").String() + "/open/v1"
	}

	configJSON := configmap.OpenClawConfig(configmap.OpenClawConfigOptions{
		ModelID:             req.ModelID,
		AllowPrivateNetwork: req.AllowPrivateNetwork,
	})

	// 6. 确保用户 namespace 存在
	if _, err = namespace.Ensure(ctx, req.OwnerUpn); err != nil {
		revokeAndCleanup(ctx, apiKeyHash, instanceID)

		return nil, fmt.Errorf("创建 namespace 失败: %w", err)
	}

	// 7. 创建 ConfigMap
	if err = configmap.Create(ctx, req.OwnerUpn, instanceID, req.Name, configJSON); err != nil {
		revokeAndCleanup(ctx, apiKeyHash, instanceID)

		return nil, fmt.Errorf("创建 ConfigMap 失败: %w", err)
	}

	// 8. 创建 NetworkPolicy
	if err = networkpolicy.Ensure(ctx, req.OwnerUpn, instanceID, req.Name); err != nil {
		configmap.Delete(ctx, req.OwnerUpn, instanceID)
		revokeAndCleanup(ctx, apiKeyHash, instanceID)

		return nil, fmt.Errorf("创建 NetworkPolicy 失败: %w", err)
	}

	// 9. 确保用户数据 PVC 存在
	pvcName, err := pvc.EnsureInstance(ctx, req.OwnerUpn, req.Name, "", defaultDisk)
	if err != nil {
		_ = networkpolicy.Delete(ctx, req.OwnerUpn, instanceID, req.Name)
		configmap.Delete(ctx, req.OwnerUpn, instanceID)
		revokeAndCleanup(ctx, apiKeyHash, instanceID)

		return nil, fmt.Errorf("创建 PVC 失败: %w", err)
	}

	updateInstancePVCName(ctx, instanceID, pvcName)

	// 10. 创建 Pod
	createdPod, err := pod.Create(ctx, pod.Config{
		InstanceID:    instanceID,
		InstanceName:  req.Name,
		OwnerUpn:      req.OwnerUpn,
		Image:         req.Image,
		MountPath:     mountPath,
		CPU:           defaultCPU,
		Memory:        defaultMemory,
		GPUCount:      req.GPUCount,
		ContainerPort: containerPort,
		GatewayPort:   gatewayPort,
		GPUEnabled:    req.GPUEnabled,
		PVCName:       pvcName,
		ExtraEnv:      extraEnv,
	})
	if err != nil {
		_ = networkpolicy.Delete(ctx, req.OwnerUpn, instanceID, req.Name)
		configmap.Delete(ctx, req.OwnerUpn, instanceID)
		revokeAndCleanup(ctx, apiKeyHash, instanceID)

		return nil, fmt.Errorf("创建 Pod 失败: %w", err)
	}

	// 11. 更新数据库中的 Pod 信息
	updateInstancePodInfo(ctx, instanceID, createdPod.Name, createdPod.Namespace)

	g.Log().Infof(ctx, "实例 %d 创建成功，等待 SyncService 更新为 running", instanceID)

	return &CreateResult{
		InstanceID:   instanceID,
		PodName:      createdPod.Name,
		PodNamespace: createdPod.Namespace,
	}, nil
}

// generateToken 生成随机 token
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)

	return hex.EncodeToString(b)
}

// checkQuota 校验用户实例数量配额
func checkQuota(ctx context.Context, req CreateRequest) error {
	quota, err := dao.UserQuotas.Ctx(ctx).Where("owner_upn", req.OwnerUpn).One()
	if err != nil {
		return fmt.Errorf("查询用户配额失败: %w", err)
	}

	if quota.IsEmpty() {
		return nil
	}

	count, err := dao.Instances.Ctx(ctx).
		Where("owner_upn", req.OwnerUpn).
		WhereNotIn("status", []string{"deleting", "error"}).
		Count()
	if err != nil {
		return fmt.Errorf("查询实例数量失败: %w", err)
	}

	maxInstances := quota["max_instances"].Int()
	if maxInstances > 0 && count >= maxInstances {
		return fmt.Errorf("实例数量超限: %d/%d", count, maxInstances)
	}

	return nil
}

// checkNameUnique 校验实例名在用户范围内唯一
func checkNameUnique(ctx context.Context, ownerUpn, name string) error {
	count, err := dao.Instances.Ctx(ctx).Where("owner_upn", ownerUpn).Where("name", name).Count()
	if err != nil {
		return fmt.Errorf("检查实例名称失败: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("实例名称 '%s' 已存在", name)
	}

	return nil
}

// insertInstance 写入实例记录
func insertInstance(ctx context.Context, req CreateRequest, mountPath string, containerPort int32, cpu, memory, disk, gatewayToken string) (int64, error) {
	result, err := dao.Instances.Ctx(ctx).Data(do.Instances{
		OwnerUpn:            req.OwnerUpn,
		Name:                req.Name,
		Description:         req.Description,
		Status:              "creating",
		Image:               req.Image,
		StorageClass:        req.StorageClass,
		MountPath:           mountPath,
		CpuCores:            cpu,
		MemoryGb:            memory,
		DiskGb:              disk,
		GpuEnabled:          req.GPUEnabled,
		AllowPrivateNetwork: req.AllowPrivateNetwork,
		GpuCount:            req.GPUCount,
		ContainerPort:       containerPort,
		ApiMode:             req.ApiMode,
		ModelId:             req.ModelID,
		ApiKey:              req.ApiKey,
		BaseUrl:             req.BaseUrl,
		GatewayToken:        gatewayToken,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return result, nil
}

// updateInstanceApiKeyHash 更新 API key hash
func updateInstanceApiKeyHash(ctx context.Context, instanceID int64, apiKeyHash string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		ApiKeyHash: apiKeyHash,
	}).Update(); err != nil {
		g.Log().Warningf(ctx, "更新实例 %d API key hash 失败: %v", instanceID, err)
	}
}

// updateInstancePVCName 更新实例的 PVC 名称
func updateInstancePVCName(ctx context.Context, instanceID int64, pvcName string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		PvcName: pvcName,
	}).Update(); err != nil {
		g.Log().Warningf(ctx, "更新实例 %d PVC 名称失败: %v", instanceID, err)
	}
}

// updateInstancePodInfo 更新实例的 Pod 信息和启动时间
func updateInstancePodInfo(ctx context.Context, instanceID int64, podName, podNamespace string) {
	if _, err := dao.Instances.Ctx(ctx).Where("id", instanceID).Data(do.Instances{
		PodName:      podName,
		PodNamespace: podNamespace,
		StartedAt:    gtime.Now(),
	}).Update(); err != nil {
		g.Log().Warningf(ctx, "更新实例 %d Pod 信息失败: %v", instanceID, err)
	}
}

// revokeAndCleanup 创建失败时的清理：吊销 API key + 删除 DB 记录
func revokeAndCleanup(ctx context.Context, apiKeyHash string, instanceID int64) {
	if apiKeyHash != "" {
		_ = uniauth.RevokeApiKey(ctx, apiKeyHash)
	}

	deleteInstance(ctx, instanceID)
}
