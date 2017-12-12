package routes

import (
	"net/http"

	"fmt"
	"time"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/mail-templater/upstreams"
	"git.containerum.net/ch/user-manager/models"
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

	if !user.IsActive {
		link, err := svc.DB.GetLinkForUser(models.LinkTypeConfirm, user)
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if link == nil {
			link, err = svc.DB.CreateLink(models.LinkTypeConfirm, 24*time.Hour, user)
			if err != nil {
				ctx.Error(err)
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if tdiff := time.Now().UTC().Sub(link.SentAt); tdiff < 5*time.Minute {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{
				Text: fmt.Sprintf("can`t resend link, wait %f seconds", tdiff.Seconds()),
			})
			return
		}

		err = svc.MailClient.SendConfirmationMail(&upstreams.Recipient{
			ID:        user.ID,
			Name:      user.Login,
			Email:     user.Login,
			Variables: map[string]string{"CONFIRM": link.Link},
		})
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusOK)
		return
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role(user.Role),
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func oneTimeTokenLoginHandler(ctx *gin.Context) {
	var request OneTimeTokenLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}

	token, err := svc.DB.GetTokenObject(request.Token)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if token == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, chutils.Error{Text: "one-time " + request.Token + " not exists or invalid"})
		return
	}
	if token.User.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, chutils.Error{Text: "user " + token.User.Login + " banned"})
		return
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	var tokens *auth.CreateTokenResponse

	err = svc.DB.Transactional(func(tx *models.DB) error {
		var err error
		tokens, err = svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
			UserAgent:   ctx.Request.UserAgent(),
			UserId:      &common.UUID{Value: token.User.ID},
			UserIp:      ctx.ClientIP(),
			UserRole:    auth.Role(token.User.Role),
			RwAccess:    true,
			Access:      access,
			PartTokenId: nil,
		})
		if err != nil {
			return err
		}

		token.IsActive = false
		token.SessionID = "sid" // TODO: session ID here
		if err := tx.UpdateToken(token); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func oauthLoginHandler(ctx *gin.Context) {
	var request OAuthLoginRequest
	if err := ctx.ShouldBindWith(&request, binding.JSON); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, chutils.Error{Text: err.Error()})
		return
	}
}
