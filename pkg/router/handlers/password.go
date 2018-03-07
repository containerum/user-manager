package handlers

import (
	"net/http"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	ch "git.containerum.net/ch/kube-client/pkg/cherry"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func PasswordChangeHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.PasswordRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidatePasswordChangeRequest(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.ChangePassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableChangePassword(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}

func PasswordResetHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserLogin(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.ResetPassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableResetPassword(), ctx)
		}
		return
	}

	ctx.Status(http.StatusOK)
}

func PasswordRestoreHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.PasswordRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidatePasswordRestoreRequest(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.RestorePassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableResetPassword(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}
