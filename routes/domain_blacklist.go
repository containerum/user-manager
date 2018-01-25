package routes

import (
	"net/http"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
)

func blacklistDomainAddHandler(ctx *gin.Context) {
	var request umtypes.DomainToBlacklistRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	err := srv.AddDomainToBlacklist(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func blacklistDomainDeleteHandler(ctx *gin.Context) {
	err := srv.RemoveDomainFromBlacklist(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func blacklistDomainGetHandler(ctx *gin.Context) {
	resp, err := srv.GetBlacklistedDomain(ctx.Request.Context(), ctx.Param("domain"))
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func blacklistDomainsListGetHandler(ctx *gin.Context) {
	resp, err := srv.GetBlacklistedDomainsList(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)

}
