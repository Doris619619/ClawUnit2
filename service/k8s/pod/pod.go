package pod

import (
	corev1 "k8s.io/api/core/v1"
)

// initVolumeMounts init container 的 volume 挂载（被 Create 使用）
func initVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: "instance-data", MountPath: "/home/node/.openclaw"},
		{Name: "openclaw-config", MountPath: "/config"},
	}
}

// gatewayVolumeMounts gateway 容器的 volume 挂载（被 Create 使用）
func gatewayVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: "instance-data", MountPath: "/home/node/.openclaw"},
		{Name: "system-skills", MountPath: "/skills/system", ReadOnly: true},
		{Name: "playwright-browsers", MountPath: "/home/node/.cache/ms-playwright", ReadOnly: true},
		{Name: "tmp-volume", MountPath: "/tmp"},
		{Name: "dshm", MountPath: "/dev/shm"},
	}
}

// podVolumes 生成 Pod 的 volume 列表（被 Create 使用）
func podVolumes(configMapName, systemSkillsPVC, playwrightPVC, instancePVCName string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "openclaw-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
				},
			},
		},
		{
			Name: "system-skills",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: systemSkillsPVC,
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "playwright-browsers",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: playwrightPVC,
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "tmp-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
		{
			Name: "instance-data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: instancePVCName,
				},
			},
		},
	}
}
