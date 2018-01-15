package routes

import (
	"net/http"

	"strings"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func userCreateHandler(ctx *gin.Context) {
	var request umtypes.UserCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	resp, err := srv.CreateUser(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

func linkResendHandler(ctx *gin.Context) {
	var request umtypes.ResendLinkRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	err := srv.LinkResend(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusOK)
}

func activateHandler(ctx *gin.Context) {
	var request umtypes.ActivateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	tokens, err := srv.ActivateUser(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func userToBlacklistHandler(ctx *gin.Context) {
	var request umtypes.UserToBlacklistRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	err := srv.BlacklistUser(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func blacklistGetHandler(ctx *gin.Context) {
	var params umtypes.UserListQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, errors.New(err.Error()))
		return
	}

	resp, err := srv.GetBlacklistedUsers(ctx.Request.Context(), params)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func linksGetHandler(ctx *gin.Context) {
	resp, err := srv.GetUserLinks(ctx.Request.Context(), ctx.Param("user_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func userInfoGetHandler(ctx *gin.Context) {
	resp, err := srv.GetUserInfo(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func userInfoUpdateHandler(ctx *gin.Context) {
	var newData umtypes.ProfileData
	if err := ctx.ShouldBindWith(&newData, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	resp, err := srv.UpdateUser(ctx.Request.Context(), newData)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func userListGetHandler(ctx *gin.Context) {
	var params umtypes.UserListQuery
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	filters := strings.Split(ctx.Query("filters"), ",")
	resp, err := srv.GetUsers(ctx.Request.Context(), params, filters...)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func partialDeleteHandler(ctx *gin.Context) {
	err := srv.PartiallyDeleteUser(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func completeDeleteHandler(ctx *gin.Context) {
	err := srv.CompletelyDeleteUser(ctx.Request.Context(), "") // TODO: take from request
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}
