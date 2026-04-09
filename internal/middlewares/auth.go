package middlewares

import (
	"context"

	"clawunit.cuhksz/internal/service/uniauth"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OwnerFromCtx 从 ctx 中提取 InjectIdentity 中间件注入的 ownerUpn。
//
// 经过 InjectIdentity 中间件的请求保证返回非空 string；没经过时返回
// 空字符串。controller 应该用本函数取值而不是直接 ctx.Value(...).(string)，
// 这样可以避开 forcetypeassert linter 报错。
func OwnerFromCtx(ctx context.Context) string {
	upn, _ := ctx.Value("ownerUpn").(string)

	return upn
}

// InjectIdentity 从 X-User-ID 请求头提取 UPN，调用 UniAuth 校验
// (clawunit, access) 权限，成功后把 UPN 注入 ctx 的 "ownerUpn" key。
//
// 失败情况：
//   - 缺少 header：401
//   - UniAuth 服务异常：500
//   - 无 access 权限：401
//
// 后续 controller 用 OwnerFromCtx 取值。
func InjectIdentity(r *ghttp.Request) {
	upn := r.Header.Get("X-User-ID")
	if upn == "" {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "缺少 X-User-ID 请求头"))

		return
	}

	ctx := r.Context()

	// 验证用户是否有 ClawUnit 访问权限
	ok, err := uniauth.CheckPermission(ctx, upn, "clawunit", "access")
	if err != nil {
		g.Log().Errorf(ctx, "UniAuth 权限校验失败，upn: %s, error: %v", upn, err)
		r.SetError(gerror.NewCode(gcode.CodeInternalError, "权限校验服务异常"))

		return
	}

	if !ok {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "无权访问 ClawUnit"))

		return
	}

	r.SetCtxVar("ownerUpn", upn)
	r.Middleware.Next()
}

// RequireAdmin 校验当前用户是否拥有 (clawunit, admin) 权限。
//
// 必须挂在 InjectIdentity 之后才能从 ctx 取到 ownerUpn。失败返回 401。
func RequireAdmin(r *ghttp.Request) {
	ctx := r.Context()

	upn, ok := ctx.Value("ownerUpn").(string)
	if !ok || upn == "" {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "缺少用户身份"))

		return
	}

	isAdmin, err := uniauth.CheckPermission(ctx, upn, "clawunit", "admin")
	if err != nil {
		g.Log().Errorf(ctx, "UniAuth 管理员权限校验失败，upn: %s, error: %v", upn, err)
		r.SetError(gerror.NewCode(gcode.CodeInternalError, "权限校验服务异常"))

		return
	}

	if !isAdmin {
		r.SetError(gerror.NewCode(gcode.CodeNotAuthorized, "需要管理员权限"))

		return
	}

	r.Middleware.Next()
}
