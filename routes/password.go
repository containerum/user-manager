package routes

import (
	"net/http"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
)

func passwordChangeHandler(ctx *gin.Context) {
	var request umtypes.PasswordChangeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.ChangePassword(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}

func passwordResetHandler(ctx *gin.Context) {
	var request umtypes.PasswordResetRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	err := srv.ResetPassword(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusOK)
}

func passwordRestoreHandler(ctx *gin.Context) {
	var request umtypes.PasswordRestoreRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.RestorePassword(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusAccepted, tokens)
}
