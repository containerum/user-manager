package impl

import (
	"context"

	"time"

	"git.containerum.net/ch/auth/storages"
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
	if err := u.checkReCaptcha(ctx, request.ReCaptcha); err != nil {
		return nil, err
	}
	user, err := u.svc.DB.GetUserByLogin(ctx, request.Username)
	if err := u.handleDBError(err); err != nil {
		return resp, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return resp, err
	}
	if !utils.CheckPassword(request.Username, request.Password, user.Salt, user.PasswordHash) {
		return resp, &server.AccessDeniedError{Err: errors.New(invalidPassword)}
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
		return resp, nil
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
		if err := tx.UpdateToken(ctx, token); err != nil {
			return oneTimeTokenUpdateFailed
		}

		var err error
		tokens, err = u.createTokens(ctx, token.User)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (u *serverImpl) OAuthLogin(ctx context.Context, request umtypes.OAuthLoginRequest) (*auth.CreateTokenResponse, error) {
	u.log.WithFields(logrus.Fields{
		"resource":        request.Resource,
		"key_to_exchange": request.AccessToken,
	}).Infof("OAuth login: %#v", request)
	resource, exist := clients.OAuthClientByResource(request.Resource)
	if !exist {
		return nil, &server.BadRequestError{Err: errors.Format(resourceNotSupported, request.Resource)}
	}
	info, err := resource.GetUserInfo(ctx, request.AccessToken)
	if err != nil {
		return nil, oauthUserInfoGetFailed
	}

	user, err := u.svc.DB.GetUserByLogin(ctx, info.Email)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, err
	}

	accounts, err := u.svc.DB.GetUserBoundAccounts(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, boundAccountsGetFailed
	}
	if accounts == nil {
		err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
			return tx.BindAccount(ctx, user, request.Resource, info.UserID)
		})
		if err := u.handleDBError(err); err != nil {
			return nil, bindAccountFailed
		}
	}

	return u.createTokens(ctx, user)
}

func (u *serverImpl) WebAPILogin(ctx context.Context, request umtypes.WebAPILoginRequest) (map[string]interface{}, error) {
	u.log.WithField("username", request.Username).Infof("Login through web-api")

	resp, code, err := u.svc.WebAPIClient.Login(ctx, &request)
	if err != nil {
		return nil, &server.WebAPIError{Err: err.(*errors.Error), StatusCode: code}
	}

	tokens, err := u.createTokens(ctx, &models.User{
		ID:   resp["user"].(map[string]interface{})["id"].(string),
		Role: umtypes.RoleUser,
	})
	if err != nil {
		return nil, err
	}

	resp["access_token"] = tokens.AccessToken
	resp["refresh_token"] = tokens.RefreshToken

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
	switch {
	case err == nil:
	case err.Error() == storages.ErrInvalidToken.Error():
		return &server.BadRequestError{Err: errors.New(err.Error())}
	case err.Error() == storages.ErrTokenNotOwnedBySender.Error():
		return &server.AccessDeniedError{Err: errors.New(err.Error())}
	default:
		return oneTimeTokenDeleteFailed
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