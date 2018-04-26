package handlers

import (
	"net/http"

	ch "git.containerum.net/ch/kube-client/pkg/cherry"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BasicLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request umtypes.LoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateLoginRequest(request); errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.BasicLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func OneTimeTokenLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request umtypes.OneTimeTokenLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	tokens, err := um.OneTimeTokenLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func OAuthLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request umtypes.OAuthLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateOAuthLoginRequest(request); errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.OAuthLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func LogoutHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.Logout(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrLogoutFailed(), ctx)
		}
		return
	}

	ctx.Status(http.StatusOK)
}
