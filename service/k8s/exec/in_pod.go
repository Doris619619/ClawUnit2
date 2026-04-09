package exec

import (
	"bytes"
	"context"
	"fmt"

	"clawunit.cuhksz/internal/service/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// Result 是 InPod 的返回值，分别保存 stdout 和 stderr 文本。
//
// 即使命令非零退出（InPod 返回 error），Result 仍然包含已收到的输出，
// caller 可以读 stderr 显示给用户。
type Result struct {
	Stdout string
	Stderr string
}

// InPod 在指定 Pod 的指定容器内执行命令，全部 stdout/stderr 收完后返回。
//
// 适合短命令（ls / cat 小文件 / git rev-parse 等）。需要流式输出或读
// 大文件时用 InPodStream。command[0] 是程序名，后续是参数 ——
// 这是 exec 模式，*不会经过 shell*，所以管道、重定向、glob 都不会展开。
// 需要 shell 时显式调用 sh -c "..."。
func InPod(ctx context.Context, namespace, podName, container string, command []string) (*Result, error) {
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
		return nil, fmt.Errorf("创建 exec 连接失败: %w", err)
	}

	var stdout, stderr bytes.Buffer

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return &Result{Stdout: stdout.String(), Stderr: stderr.String()}, fmt.Errorf("执行命令失败: %w", err)
	}

	return &Result{Stdout: stdout.String(), Stderr: stderr.String()}, nil
}
