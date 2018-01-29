package routes

import (
	"net/http"

	"strings"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
)

var validate *validator.Validate

func userCreateHandler(ctx *gin.Context) {
	var request umtypes.UserCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
	resp, err := srv.GetUserInfo(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func userGetHandler(ctx *gin.Context) {

	resp, err := srv.GetUserInfoByID(ctx.Request.Context(), ctx.Param("user_id"))
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func userInfoUpdateHandler(ctx *gin.Context) {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)

	var newData map[string]interface{}
	if err := ctx.ShouldBindJSON(&newData); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
		return
	}

	if err := validate.Field(newData["email"], "omitempty,email"); err != nil {
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
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
	err := srv.PartiallyDeleteUser(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func completeDeleteHandler(ctx *gin.Context) {
	var request umtypes.CompleteDeleteHandlerRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
		return
	}

	err := srv.CompletelyDeleteUser(ctx.Request.Context(), request.UserID)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func addBoundAccountHandler(ctx *gin.Context) {
	var request umtypes.OAuthLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
		return
	}

	err := srv.AddBoundAccount(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

func getBoundAccountsHandler(ctx *gin.Context) {
	resp, err := srv.GetBoundAccounts(ctx.Request.Context())
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func deleteBoundAccountHandler(ctx *gin.Context) {
	var request umtypes.BoundAccountDeleteRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ParseBindErorrs(err))
		return
	}

	err := srv.DeleteBoundAccount(ctx.Request.Context(), request)
	if err != nil {
		ctx.AbortWithStatusJSON(errorWithHTTPStatus(err))
		return
	}

	ctx.Status(http.StatusAccepted)
}
