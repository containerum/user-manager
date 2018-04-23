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

func AdminUserCreateHandler(ctx *gin.Context) {
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

	resp, err := um.AdminCreateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableCreateUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func AdminUserActicate(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	resp, err := um.AdminActivateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, resp)
}

func AdminUserDeactivate(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	err := um.AdminDeactivateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func AdminSetAdmin(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	err := um.AdminSetAdmin(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func AdminResetPassword(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.UserLogin

	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	resp, err := um.AdminResetPassword(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableDeleteUser(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusAccepted, resp)
}
