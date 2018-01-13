package routes

import (
	"net/http"

	"time"

	"context"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	oneTimeTokenNotFound = "one-time token %s not exists or already used"
	resourceNotSupported = "resource %s not supported now"
	activationNeeded     = "Activate your account please. Check your email"
)

func basicLoginHandler(ctx *gin.Context) {
	var request umtypes.BasicLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.New(err.Error()))
		return
	}

	user, err := svc.DB.GetUserByLogin(ctx, request.Username)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, userGetFailed)
		return
	}
	if user == nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, errors.Format(userNotFound, request.Username))
		return
	}
	if user.IsInBlacklist {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.Format(userBanned, request.Username))
		return
	}

	if !utils.CheckPassword(request.Username, request.Password, user.Salt, user.PasswordHash) {
		ctx.AbortWithStatusJSON(http.StatusForbidden, errors.New(invalidPassword))
		return
	}

	if !user.IsActive {
		link, err := svc.DB.GetLinkForUser(ctx, umtypes.LinkTypeConfirm, user)
		if err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, linkGetFailed)
			return
		}

		if link == nil {
			link, err = svc.DB.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, user)
			if err != nil {
				ctx.Error(err)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, linkCreateFailed)
				return
			}
		}

		if tdiff := time.Now().UTC().Sub(link.SentAt.Time); link.SentAt.Valid && tdiff < 5*time.Minute {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.Format(waitForResend, int(tdiff.Seconds())))
			return
		}

		go func() {
			err := svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
				err := svc.MailClient.SendConfirmationMail(ctx, &mttypes.Recipient{
					ID:        user.ID,
					Name:      user.Login,
					Email:     user.Login,
					Variables: map[string]string{"CONFIRM": link.Link},
				})
				if err != nil {
					return err
				}
				link.SentAt.Time = time.Now().UTC()
				link.SentAt.Valid = true
				return tx.UpdateLink(ctx, link)
			})
			if err != nil {
				logrus.WithError(err).Error("email send failed")
			}
		}()

		ctx.AbortWithStatusJSON(http.StatusBadRequest, errors.Format(activationNeeded)) // TODO: may be other status code/message
		return
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		Fingerprint: ctx.GetHeader(umtypes.FingerprintHeader),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role(user.Role),
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenCreateFailed)
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

	token, err := svc.DB.GetTokenObject(ctx, request.Token)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, getTokenFailed)
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

	rctx := ctx
	err = svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		var err error
		tokens, err = svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
			UserAgent:   rctx.Request.UserAgent(),
			Fingerprint: rctx.GetHeader(umtypes.FingerprintHeader),
			UserId:      &common.UUID{Value: token.User.ID},
			UserIp:      rctx.ClientIP(),
			UserRole:    auth.Role(token.User.Role),
			RwAccess:    true,
			Access:      access,
			PartTokenId: nil,
		})
		if err != nil {
			return err
		}

		token.IsActive = false
		token.SessionID = rctx.GetHeader(umtypes.SessionIDHeader)
		if err := tx.UpdateToken(ctx, token); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenCreateFailed)
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

	info, err := resource.GetUserInfo(ctx, request.AccessToken)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, oauthUserInfoGetFailed)
		return
	}

	user, err := svc.DB.GetUserByLogin(ctx, info.Email)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, userGetFailed)
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

	accounts, err := svc.DB.GetUserBoundAccounts(ctx, user)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, boundAccountsGetFailed)
		return
	}
	if accounts == nil {
		if err := svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
			return tx.BindAccount(ctx, user, request.Resource, info.UserID)
		}); err != nil {
			ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, bindAccountFailed)
			return
		}
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		Fingerprint: ctx.GetHeader(umtypes.FingerprintHeader),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role(user.Role),
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenCreateFailed)
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

	resp, code, err := svc.WebAPIClient.Login(ctx, &request)
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(code, errors.New(err.Error()))
		return
	}

	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	tokens, err := svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   ctx.Request.UserAgent(),
		Fingerprint: ctx.GetHeader(umtypes.FingerprintHeader),
		UserId:      &common.UUID{Value: resp["user"].(map[string]interface{})["id"].(string)},
		UserIp:      ctx.ClientIP(),
		UserRole:    auth.Role_USER,
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
	if err != nil {
		ctx.Error(err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, tokenCreateFailed)
		return
	}

	resp["access_token"] = tokens.AccessToken
	resp["refresh_token"] = tokens.RefreshToken

	ctx.JSON(http.StatusOK, resp)
}
