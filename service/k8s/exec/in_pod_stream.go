package exec

import (
	"context"
	"fmt"
	"io"

	"clawunit.cuhksz/internal/service/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// InPodStream 在 Pod 内执行命令，把 stdout/stderr 流式写入传入的 writer。
//
// 适合长输出（SSE 流、tar 打包、大文件下载）。和 InPod 一样不经过 shell。
// stdout 和 stderr 可以指向同一个 writer，调用方需自行处理交错问题。
//
// 阻塞直到命令退出或 ctx 取消。
func InPodStream(ctx context.Context, namespace, podName, container string, command []string, stdout, stderr io.Writer) error {
	c := k8s.GetClient()

	req := c.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建 exec 连接失败: %w", err)
	}

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: stderr,
	})
}
