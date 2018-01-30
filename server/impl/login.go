package impl

import (
	"context"

	"time"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/clients"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"git.containerum.net/ch/user-manager/utils"
	"github.com/sirupsen/logrus"
)

func (u *serverImpl) BasicLogin(ctx context.Context, request umtypes.BasicLoginRequest) (resp *auth.CreateTokenResponse, err error) {
	u.log.Infof("Basic login: %#v", request)
	if rcErr := u.checkReCaptcha(ctx, request.ReCaptcha); rcErr != nil {
		return nil, rcErr
	}
	user, err := u.svc.DB.GetUserByLogin(ctx, request.Username)
	if dbErr := u.handleDBError(err); dbErr != nil {
		return resp, userGetFailed
	}
	if checksErr := u.loginUserChecks(ctx, user); checksErr != nil {
		return resp, checksErr
	}
	if !utils.CheckPassword(request.Username, request.Password, user.Salt, user.PasswordHash) {
		return resp, invalidPassword
	}
	if !user.IsActive {
		link, err := u.svc.DB.GetLinkForUser(ctx, umtypes.LinkTypeConfirm, user)
		if err != nil {
			return resp, linkGetFailed
		}
		if link == nil {
			err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
				var err error
				link, err = tx.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, user)
				return err
			})
			if err := u.handleDBError(err); err != nil {
				return resp, linkCreateFailed
			}
		}
		if err := u.checkLinkResendTime(ctx, link); err != nil {
			return resp, err
		}
		go u.linkSend(ctx, link)
		return resp, activationNeeded
	}
	resp, err = u.createTokens(ctx, user)
	return
}

func (u *serverImpl) OneTimeTokenLogin(ctx context.Context, request umtypes.OneTimeTokenLoginRequest) (*auth.CreateTokenResponse, error) {
	u.log.WithField("token", request.Token).Info("One-time token login")
	token, err := u.svc.DB.GetTokenObject(ctx, request.Token)
	if err != nil {
		return nil, oneTimeTokenGetFailed
	}
	if err := u.loginUserChecks(ctx, token.User); err != nil {
		return nil, err
	}

	var tokens *auth.CreateTokenResponse
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		token.IsActive = false
		token.SessionID = server.MustGetSessionID(ctx)
		if updErr := tx.UpdateToken(ctx, token); updErr != nil {
			return oneTimeTokenUpdateFailed
		}

		var err error
		tokens, err = u.createTokens(ctx, token.User)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, oneTimeTokenCreateFailed
	}
	return tokens, nil
}

//nolint: gocyclo
func (u *serverImpl) OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (*auth.CreateTokenResponse, error) {
	u.log.WithFields(logrus.Fields{
		"resource":        request.Resource,
		"key_to_exchange": request.AccessToken,
	}).Infof("OAuth login: %#v", request)
	resource, exist := clients.OAuthClientByResource(request.Resource)
	if !exist {
		return nil, &server.BadRequestError{Err: errors.Format(resourceNotSupported, request.Resource)}
	}
	info, oauthError := resource.GetUserInfo(ctx, request.AccessToken)
	if oauthError != nil {
		switch oauthError.Code {
		case 403, 401:
			return nil, oauthLoginFailed
		default:
			return nil, oauthUserInfoGetFailed
		}
	}

	user, err := u.svc.DB.GetUserByLogin(ctx, info.Email)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if err = u.loginUserChecks(ctx, user); err != nil {
		u.log.Info("User is not found by email. Checking bound accounts")
		if info.UserID != "" {
			user, err = u.svc.DB.GetUserByBoundAccount(ctx, request.Resource, info.UserID)
			if err = u.handleDBError(err); err != nil {
				return nil, userGetFailed
			}
			if err := u.loginUserChecks(ctx, user); err != nil {
				return nil, err
			}
		}
		return nil, userNotFound
	}

	u.log.Info("User is found by email. Binding account")
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BindAccount(ctx, user, request.Resource, info.UserID)
	})
	if err := u.handleDBError(err); err != nil {
		return nil, bindAccountFailed
	}
	return u.createTokens(ctx, user)
}

func (u *serverImpl) WebAPILogin(ctx context.Context, request umtypes.WebAPILoginRequest) (map[string]interface{}, error) {
	u.log.WithField("username", request.Username).Infof("Login through web-api")

	resp, _, err := u.svc.WebAPIClient.Login(ctx, &request)
	if err != nil {
		return nil, err
	}

	tokens, err := u.createTokens(ctx, &models.User{
		ID:   resp["user"].(map[string]interface{})["id"].(string),
		Role: "user",
	})
	if err != nil {
		return nil, tokenCreateFailed
	}

	resp["access_token"] = tokens.AccessToken
	resp["refresh_token"] = tokens.RefreshToken

	newUser := umtypes.UserCreateWebAPIRequest{ID: resp["user"].(map[string]interface{})["id"].(string),
		UserName:  resp["user"].(map[string]interface{})["login"].(string),
		Password:  request.Password,
		Data:      resp["user"].(map[string]interface{})["data"].(map[string]interface{}),
		CreatedAt: resp["user"].(map[string]interface{})["created_at"].(string),
		IsActive:  resp["user"].(map[string]interface{})["is_active"].(bool)}

	if _, err = u.CreateUserWebAPI(ctx, newUser); err != nil {
		u.log.WithError(err).Warnf("Unable to add user to new db")
	}

	return resp, nil

}

func (u *serverImpl) Logout(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	tokenID := server.MustGetTokenID(ctx)
	sessionID := server.MustGetSessionID(ctx)
	u.log.WithFields(logrus.Fields{
		"user_id":    userID,
		"token_id":   tokenID,
		"session_id": sessionID,
	}).Info("Logout")

	_, err := u.svc.AuthClient.DeleteToken(ctx, &auth.DeleteTokenRequest{
		UserId:  &common.UUID{Value: userID},
		TokenId: &common.UUID{Value: tokenID},
	})
	if err != nil {
		return tokenDeleteFailed
	}

	oneTimeToken, err := u.svc.DB.GetTokenBySessionID(ctx, sessionID)
	if err := u.handleDBError(err); err != nil {
		return oneTimeTokenGetFailed
	}
	if oneTimeToken != nil {
		if oneTimeToken.User.ID != userID {
			return &server.AccessDeniedError{Err: errors.Format(tokenNotOwnedByUser, oneTimeToken.Token, oneTimeToken.User.Login)}
		}
		err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
			return u.svc.DB.DeleteToken(ctx, oneTimeToken.Token)
		})
		if err = u.handleDBError(err); err != nil {
			return oneTimeTokenDeleteFailed
		}
	}
	return nil
}
