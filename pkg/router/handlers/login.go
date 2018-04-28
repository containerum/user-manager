package handlers

import (
	"net/http"

	"git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/containerum/cherry"
	"github.com/containerum/cherry/adaptors/gonic"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BasicLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.LoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateLoginRequest(request); errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.BasicLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func OneTimeTokenLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.OneTimeTokenLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	tokens, err := um.OneTimeTokenLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func OAuthLoginHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.OAuthLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	if errs := validation.ValidateOAuthLoginRequest(request); errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.OAuthLogin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrLoginFailed(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func LogoutHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	err := um.Logout(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrLogoutFailed(), ctx)
		}
		return
	}

	ctx.Status(http.StatusOK)
}
