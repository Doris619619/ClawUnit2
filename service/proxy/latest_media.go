package proxy

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"

	"clawunit.cuhksz/internal/dao"
	"clawunit.cuhksz/internal/service/k8s/exec"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const mediaBase = "/home/node/.openclaw/media"

var (
	// 图片文件头 magic bytes
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	webpMagic = []byte{0x52, 0x49, 0x46, 0x46} // "RIFF"

	// 合法文件名：UUID + 图片扩展名
	safeFilenameRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.(png|jpg|jpeg|webp)$`)

	// 允许的媒体子目录白名单
	allowedMediaTypes = map[string]bool{
		"browser":   true, // 浏览器截图
		"generated": true, // AI 图片生成
		"camera":    true, // 摄像头捕获
		"canvas":    true, // 画布导出
	}
)

// LatestMedia 读取实例 Pod 上指定 mediaType 目录中的最新文件并直接
// 返回给前端。当前用于在聊天 UI 里渲染浏览器截图。
//
// 路由：GET /api/gateway/v1/media/{instanceId}/{mediaType}
//
// 安全检查（任一失败立即拒绝）：
//
//   - mediaType 必须在白名单内（browser/generated/camera/canvas）
//   - 实例必须属于当前用户且处于 running 状态
//   - 文件名必须是 UUID + 图片扩展名（防止读到任意文件）
//   - readlink -f 解析后的真实路径必须仍在 mediaBase 之下（防符号链接穿越）
//   - 文件内容必须有合法的 PNG/JPEG/WEBP magic bytes
//
// 失败情况都会写 [审计] warning 日志，方便事后追溯。
func LatestMedia(r *ghttp.Request) {
	ctx := r.Context()
	ownerUpn := r.Header.Get("X-User-ID")

	if ownerUpn == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)

		return
	}

	instanceID := r.Get("instanceId").Int64()
	mediaType := r.Get("mediaType").String()

	if instanceID <= 0 || mediaType == "" {
		r.Response.WriteStatus(http.StatusBadRequest)

		return
	}

	// 校验媒体类型白名单
	if !allowedMediaTypes[mediaType] {
		r.Response.WriteStatus(http.StatusForbidden)

		return
	}

	// 校验实例归属
	record, err := dao.Instances.Ctx(ctx).
		Where("id", instanceID).
		Where("owner_upn", ownerUpn).
		One()
	if err != nil || record.IsEmpty() {
		r.Response.WriteStatus(http.StatusNotFound)

		return
	}

	if record["status"].String() != "running" {
		r.Response.WriteStatus(http.StatusBadRequest)

		return
	}

	podName := record["pod_name"].String()
	podNamespace := record["pod_namespace"].String()
	mediaDir := mediaBase + "/" + mediaType

	// Step 1: 列出最新文件名
	lsResult, err := exec.InPod(ctx, podNamespace, podName, "gateway", []string{
		"ls", "-t", mediaDir,
	})
	if err != nil || lsResult == nil || strings.TrimSpace(lsResult.Stdout) == "" {
		r.Response.WriteStatus(http.StatusNotFound)

		return
	}

	filename := strings.SplitN(strings.TrimSpace(lsResult.Stdout), "\n", 2)[0]
	filename = strings.TrimSpace(filename)

	// Step 2: 校验文件名格式（UUID + 图片扩展名）
	if !safeFilenameRe.MatchString(filename) {
		g.Log().Warningf(ctx, "[审计] 媒体文件名不合法: user=%s instance=%d type=%s filename=%s", ownerUpn, instanceID, mediaType, filename)
		r.Response.WriteStatus(http.StatusForbidden)

		return
	}

	// Step 3: readlink -f 防止符号链接穿越
	realPathResult, err := exec.InPod(ctx, podNamespace, podName, "gateway", []string{
		"readlink", "-f", mediaDir + "/" + filename,
	})
	if err != nil || realPathResult == nil {
		r.Response.WriteStatus(http.StatusInternalServerError)

		return
	}

	realPath := strings.TrimSpace(realPathResult.Stdout)
	if !strings.HasPrefix(realPath, mediaDir+"/") {
		g.Log().Warningf(ctx, "[审计] 媒体路径穿越: user=%s instance=%d type=%s realpath=%s", ownerUpn, instanceID, mediaType, realPath)
		r.Response.WriteStatus(http.StatusForbidden)

		return
	}

	// Step 4: 读取文件
	var imgBuf, stderrBuf bytes.Buffer

	err = exec.InPodStream(ctx, podNamespace, podName, "gateway", []string{
		"cat", realPath,
	}, &imgBuf, &stderrBuf)
	if err != nil {
		g.Log().Errorf(ctx, "读取媒体文件失败: %v, stderr: %s", err, stderrBuf.String())
		r.Response.WriteStatus(http.StatusInternalServerError)

		return
	}

	// Step 5: 校验图片 magic bytes
	data := imgBuf.Bytes()
	contentType := detectImageType(data)

	if contentType == "" {
		g.Log().Warningf(ctx, "[审计] 媒体内容非图片: user=%s instance=%d type=%s file=%s", ownerUpn, instanceID, mediaType, filename)
		r.Response.WriteStatus(http.StatusUnprocessableEntity)

		return
	}

	g.Log().Infof(ctx, "[审计] 媒体访问: user=%s instance=%d type=%s file=%s size=%d", ownerUpn, instanceID, mediaType, filename, len(data))

	r.Response.Header().Set("Content-Type", contentType)
	r.Response.Header().Set("Cache-Control", "no-cache")
	r.Response.Write(data)
}

// detectImageType 通过 magic bytes 检测图片类型，非图片返回空
func detectImageType(data []byte) string {
	if len(data) >= len(pngMagic) && bytes.Equal(data[:len(pngMagic)], pngMagic) {
		return "image/png"
	}

	if len(data) >= len(jpegMagic) && bytes.Equal(data[:len(jpegMagic)], jpegMagic) {
		return "image/jpeg"
	}

	if len(data) >= 12 && bytes.Equal(data[:4], webpMagic) && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}

	return ""
}
