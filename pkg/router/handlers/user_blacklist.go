package handlers

import (
	"net/http"
	"strconv"

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

func UserToBlacklistHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserID(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.BlacklistUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableBlacklistUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func UserDeleteFromBlacklistHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	var request models.UserLogin
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateUserLogin(request)
	if errs != nil {
		gonic.Gonic(umErrors.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.UnBlacklistUser(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableUnblacklistUser(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func BlacklistGetHandler(ctx *gin.Context) {
	um := ctx.MustGet(m.UMServices).(server.UserManager)

	page := int64(1)
	pagestr, ok := ctx.GetQuery("page")
	if ok {
		var err error
		page, err = strconv.ParseInt(pagestr, 10, 64)
		if err != nil {
			ctx.Error(err)
		}
	}

	perPage := int64(10)
	perPagestr, ok := ctx.GetQuery("per_page")
	if ok {
		var err error
		perPage, err = strconv.ParseInt(perPagestr, 10, 64)
		if err != nil {
			ctx.Error(err)
		}
	}

	resp, err := um.GetBlacklistedUsers(ctx.Request.Context(), int(page), int(perPage))
	if err != nil {
		if cherr, ok := err.(*cherry.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(umErrors.ErrUnableGetUserBlacklist(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
