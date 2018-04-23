package handlers

import (
	"net/http"

	ch "git.containerum.net/ch/kube-client/pkg/cherry"
	"git.containerum.net/ch/kube-client/pkg/cherry/adaptors/gonic"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/gin-gonic/gin"

	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/gin-gonic/gin/binding"
)

func LinkResendHandler(ctx *gin.Context) {
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

	err := um.LinkResend(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableResendLink(), ctx)
		}
		return
	}

	ctx.Status(http.StatusOK)
}

func ActivateHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.Link
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateLink(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	tokens, err := um.ActivateUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableActivate(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func LinksGetHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	resp, err := um.GetUserLinks(ctx.Request.Context(), ctx.Param("user_id"))
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetUserInfo(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
