package routes

import (
	"net/http"

	chutils "git.containerum.net/ch/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type BasicLoginRequest struct {
	Login     string `json:"login" binding:"required;email"`
	Password  string `json:"password" binding:"required"`
	ReCaptcha string `json:"recaptcha" binding:"required"`
}

type OneTimeTokenLoginRequest struct {
	Token string `json:"token" binding:"required"`
}

type OAuthLoginRequest struct {
	Resource    string `json:"resource" binding:"required"`
	AccessToken string `json:"access_token" binding:"required"`
}

func basicLoginHandler(ctx *gin.Context) {
	var request BasicLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	user, err := svc.DB.GetUserByLogin(request.Login)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "user " + request.Login + " not exists"})
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, chutils.Error{Text: "user " + user.Login + " banned"})
		return
	}
}

func oneTimeTokenLoginHandler(ctx *gin.Context) {
	var request OneTimeTokenLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	user, err := svc.DB.GetUserByToken(request.Token)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "one-time " + request.Token + " not exists or invalid"})
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, chutils.Error{Text: "user " + user.Login + " banned"})
		return
	}
}

func oauthLoginHandler(ctx *gin.Context) {
	var request OAuthLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}
}
