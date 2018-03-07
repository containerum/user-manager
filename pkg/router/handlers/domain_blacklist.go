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

func BlacklistDomainAddHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	var request umtypes.Domain
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(err), ctx)
		return
	}

	errs := validation.ValidateDomain(request)
	if errs != nil {
		gonic.Gonic(cherry.ErrRequestValidationFailed().AddDetailsErr(errs...), ctx)
		return
	}

	err := um.AddDomainToBlacklist(ctx.Request.Context(), request)
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableBlacklistDomain(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func BlacklistDomainDeleteHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	err := um.RemoveDomainFromBlacklist(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableUnblacklistDomain(), ctx)
		}
		return
	}

	ctx.Status(http.StatusAccepted)
}

func BlacklistDomainGetHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	resp, err := um.GetBlacklistedDomain(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetDomainBlacklist(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func BlacklistDomainsListGetHandler(ctx *gin.Context) {
	ump := ctx.MustGet(m.UMServices).(*server.UserManager)
	um := *ump

	resp, err := um.GetBlacklistedDomainsList(ctx.Request.Context())
	if err != nil {
		if cherr, ok := err.(*ch.Err); ok {
			gonic.Gonic(cherr, ctx)
		} else {
			ctx.Error(err)
			gonic.Gonic(cherry.ErrUnableGetDomainBlacklist(), ctx)
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)

}
