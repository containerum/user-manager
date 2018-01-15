package routes

import (
	"net/http"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
)

func basicLoginHandler(ctx *gin.Context) {
	var request umtypes.BasicLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.BasicLogin(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func oneTimeTokenLoginHandler(ctx *gin.Context) {
	var request umtypes.OneTimeTokenLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.OneTimeTokenLogin(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func oauthLoginHandler(ctx *gin.Context) {
	var request umtypes.OAuthLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.OAuthLogin(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func webAPILoginHandler(ctx *gin.Context) {
	var request umtypes.WebAPILoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	resp, err := srv.WebAPILogin(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
