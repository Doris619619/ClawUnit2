package middlewares

import (
	"mime"
	"net/http"
	"slices"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	contentTypeEventStream  = "text/event-stream"
	contentTypeOctetStream  = "application/octet-stream"
	contentTypeMixedReplace = "multipart/x-mixed-replace"
)

var (
	streamContentType = []string{contentTypeEventStream, contentTypeOctetStream, contentTypeMixedReplace}
)

// UnifiedResponse 是 ClawUnit HTTP 接口的统一响应外壳。
//
// 字段命名与 CUHKSZ 内部 open-platform 保持一致，前端可以共用解析逻辑。
// ShowType 是给前端的提示（0=silent, 1=warn, 2=error），ClawUnit 当前
// 全部填 2。
type UnifiedResponse struct {
	Data         any    `dc:"响应数据" json:"data"`
	ErrorMessage string `dc:"错误信息" json:"message"`
	ErrorCode    int    `dc:"错误码"   json:"code"`
	ShowType     int    `dc:"展示类型" json:"showType"`
	Success      bool   `dc:"是否成功" json:"success"`
}

// UniResMiddleware 把 controller 返回值包装成 UnifiedResponse JSON。
//
// 流式响应（SSE / octet-stream / multipart）通过 Content-Type 检测
// 自动跳过包装。错误情况下 HTTP 状态码会根据 gerror.Code 设置成
// 401/404/500 之一，并把错误消息写入 message 字段。
func UniResMiddleware(r *ghttp.Request) {
	r.Middleware.Next()

	// 如果 Response 已经被修改，则直接返回不做处理
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	// 流式响应不处理
	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	if slices.Contains(streamContentType, mediaType) {
		return
	}

	var (
		msg  string
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}

		switch code {
		case gcode.CodeNotAuthorized:
			r.Response.Status = http.StatusForbidden
		case gcode.CodeNotFound:
			r.Response.Status = http.StatusNotFound
		default:
			r.Response.Status = http.StatusInternalServerError
		}

		msg = err.Error()
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden:
				code = gcode.CodeNotAuthorized
			default:
				code = gcode.CodeUnknown
			}

			err = gerror.NewCode(code, msg)
			r.SetError(err)
		} else {
			code = gcode.CodeOK
		}

		msg = code.Message()
	}

	r.Response.WriteJson(UnifiedResponse{
		Success:      code == gcode.CodeOK,
		ErrorCode:    code.Code(),
		ErrorMessage: msg,
		Data:         res,
		ShowType:     2,
	})
}
