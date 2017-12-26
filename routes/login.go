package routes

import (
	"net/http"

	"time"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"github.com/gin-gonic/gin"
)

const (
	oneTimeTokenNotFound = "one-time token %s not exists or already used"
	resourceNotSupported = "resource %s not supported now"
)

func basicLoginHandler(ctx *gin.Context) {
	var request umtypes.BasicLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	user, err := svc.DB.GetUserByLogin(request.Login)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userNotFound, request.Login))
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(userBanned, request.Login))
		return
	}

	if !user.IsActive {
		link, err := svc.DB.GetLinkForUser(umtypes.LinkTypeConfirm, user)
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if link == nil {
			link, err = svc.DB.CreateLink(umtypes.LinkTypeConfirm, 24*time.Hour, user)
			if err != nil {
				ctx.Error(err)
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if tdiff := time.Now().UTC().Sub(link.SentAt.Time); link.SentAt.Valid && tdiff < 5*time.Minute {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.Format(waitForResend, int(tdiff.Seconds())))
			return
		}

		err = svc.MailClient.SendConfirmationMail(&mttypes.Recipient{
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
	var request umtypes.OneTimeTokenLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	token, err := svc.DB.GetTokenObject(request.Token)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if token == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(oneTimeTokenNotFound, request.Token))
		return
	}
	if token.User.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(userBanned, token.User.Login))
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
		token.SessionID = ctx.GetHeader(umtypes.SessionIDHeader)
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
	var request umtypes.OAuthLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	resource, exist := clients.OAuthClientByResource(request.Resource)
	if !exist {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.Format(resourceNotSupported, request.Resource))
		return
	}

	info, err := resource.GetUserInfo(request.AccessToken)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user, err := svc.DB.GetUserByLogin(info.Email)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userNotFound, info.Email))
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(userBanned, user.Login))
		return
	}

	accounts, err := svc.DB.GetUserBoundAccounts(user)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if accounts == nil {
		if err := svc.DB.Transactional(func(tx *models.DB) error {
			return tx.BindAccount(user, string(request.Resource), info.UserID)
		}); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
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

func webAPILoginHandler(ctx *gin.Context) {
	var request umtypes.WebAPILoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	resp, code, err := svc.WebAPIClient.Login(&request)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(code, errors.New(err.Error()))
		return
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		UserId:      &common.UUID{Value: resp["user"].(map[string]interface{})["id"].(string)},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role_USER,
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})

	resp["access_token"] = tokens.AccessToken
	resp["refresh_token"] = tokens.RefreshToken

	ctx.JSON(http.StatusOK, resp)
}
