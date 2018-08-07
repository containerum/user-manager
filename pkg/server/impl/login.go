package impl

import (
	"context"

	"time"

	"fmt"

	"git.containerum.net/ch/auth/proto"
	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"github.com/containerum/utils/httputil"
	"github.com/sirupsen/logrus"
)

func (u *serverImpl) BasicLogin(ctx context.Context, request models.LoginRequest) (resp *authProto.CreateTokenResponse, err error) {
	u.log.Infoln("Basic login")
	u.log.WithFields(logrus.Fields{
		"username": request.Login,
	}).Debugln("Basic login details")

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if dbErr := u.handleDBError(err); dbErr != nil {
		u.log.WithError(dbErr)
		return resp, cherry.ErrLoginFailed()
	}

	if err = u.loginUserChecks(user); err != nil {
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if dbErr := u.handleDBError(err); dbErr != nil {
		u.log.WithError(dbErr)
		return resp, cherry.ErrLoginFailed()
	}

	if !utils.CheckPassword(request.Login, request.Password, user.Salt, user.PasswordHash) {
		u.log.WithError(cherry.ErrInvalidLogin())
		return nil, cherry.ErrInvalidLogin()
	}
	if user.IsInBlacklist {
		return nil, cherry.ErrAccountBlocked()
	}

	if !user.IsActive {
		link, err := u.svc.DB.GetLinkForUser(ctx, models.LinkTypeConfirm, user)
		if err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrInvalidLogin()
		}
		if link == nil {
			err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
				var err error
				link, err = tx.CreateLink(ctx, models.LinkTypeConfirm, 24*time.Hour, user)
				return err
			})
			if err := u.handleDBError(err); err != nil {
				u.log.WithError(err)
				return nil, cherry.ErrInvalidLogin()
			}
		}
		if err := u.checkLinkResendTime(link); err != nil {
			u.log.WithError(err)
			return nil, err
		}
		if err := u.linkSend(ctx, link); err != nil {
			return nil, err
		}
		return nil, cherry.ErrNotActivated()
	}

	loginerr := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateLastLogin(ctx, profile.ID.String, time.Now().Format(time.RFC3339))
	})
	if loginerr := u.handleDBError(loginerr); loginerr != nil {
		u.log.WithError(loginerr)
	}
	return u.createTokens(ctx, user)
}

func (u *serverImpl) OneTimeTokenLogin(ctx context.Context, request models.OneTimeTokenLoginRequest) (*authProto.CreateTokenResponse, error) {
	u.log.Info("One-time token login")
	u.log.WithField("token", request.Token).Debug("One-time token login details")
	token, err := u.svc.DB.GetTokenObject(ctx, request.Token)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrLoginFailed()
	}
	if token != nil {
		if err := u.loginUserChecks(token.User); err != nil {
			return nil, err
		}

		profile, err := u.svc.DB.GetProfileByUser(ctx, token.User)
		if dbErr := u.handleDBError(err); dbErr != nil {
			u.log.WithError(dbErr)
			return nil, cherry.ErrLoginFailed()
		}
		var tokens *authProto.CreateTokenResponse
		token.IsActive = false
		//TODO Do something with session ID
		token.SessionID = ""
		err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
			return tx.UpdateToken(ctx, token)
		})

		if err := u.handleDBError(err); err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrLoginFailed()
		}
		tokens, err = u.createTokens(ctx, token.User)
		if err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrLoginFailed()
		}
		loginerr := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
			return tx.UpdateLastLogin(ctx, profile.ID.String, time.Now().Format(time.RFC3339))
		})
		if loginerr := u.handleDBError(loginerr); loginerr != nil {
			u.log.WithError(loginerr)
		}

		return tokens, nil
	}
	return nil, cherry.ErrInvalidLogin()
}

func (u *serverImpl) OAuthLogin(ctx context.Context, request models.OAuthLoginRequest) (*authProto.CreateTokenResponse, error) {
	u.log.WithFields(logrus.Fields{
		"resource": request.Resource,
	}).Infoln("OAuth login")
	u.log.WithFields(logrus.Fields{
		"resource":        request.Resource,
		"key_to_exchange": request.AccessToken,
	}).Debugln("OAuth login credentials")
	resource, exist := clients.OAuthClientByResource(request.Resource)
	if !exist {
		u.log.WithError(fmt.Errorf(resourceNotSupported, request.Resource))
		return nil, cherry.ErrInvalidLogin().AddDetailsErr(fmt.Errorf(resourceNotSupported, request.Resource))
	}
	info, err := resource.GetUserInfo(ctx, request.AccessToken)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableBindAccount()
	}
	user, err := u.svc.DB.GetUserByLogin(ctx, info.Email)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrLoginFailed()
	}

	if err := u.loginUserChecks(user); err != nil {
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if dbErr := u.handleDBError(err); dbErr != nil {
		u.log.WithError(dbErr)
		return nil, cherry.ErrLoginFailed()
	}
	if err = u.loginUserChecks(user); err != nil {
		u.log.Info("User is not found by email. Checking bound accounts")
		if info.UserID != "" {
			user, err = u.svc.DB.GetUserByBoundAccount(ctx, request.Resource, info.UserID)
			if err = u.handleDBError(err); err != nil {
				u.log.WithError(err)
				return nil, cherry.ErrLoginFailed()
			}
			if err := u.loginUserChecks(user); err != nil {
				return nil, err
			}
			return u.createTokens(ctx, user)
		}
		return nil, err
	}

	u.log.Info("User is found by email. Binding account")
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.BindAccount(ctx, user, request.Resource, info.UserID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrLoginFailed()
	}

	loginerr := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateLastLogin(ctx, profile.ID.String, time.Now().Format(time.RFC3339))
	})
	if loginerr := u.handleDBError(loginerr); loginerr != nil {
		u.log.WithError(loginerr)
	}
	return u.createTokens(ctx, user)
}

func (u *serverImpl) Logout(ctx context.Context) error {
	userID := httputil.MustGetUserID(ctx)
	tokenID := httputil.MustGetTokenID(ctx)
	//TODO Do something with session ID
	sessionID := ""
	u.log.WithFields(logrus.Fields{
		"user_id":    userID,
		"token_id":   tokenID,
		"session_id": sessionID,
	}).Info("Logout")

	_, err := u.svc.AuthClient.DeleteToken(ctx, &authProto.DeleteTokenRequest{
		UserId:  userID,
		TokenId: tokenID,
	})
	if err != nil {
		u.log.WithError(err)
		return cherry.ErrLogoutFailed()
	}

	oneTimeToken, err := u.svc.DB.GetTokenBySessionID(ctx, sessionID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrLogoutFailed()
	}
	if oneTimeToken != nil {
		if oneTimeToken.User.ID != userID {
			u.log.WithError(cherry.ErrInvalidLink())
			return cherry.ErrInvalidLink()
		}
		err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
			return u.svc.DB.DeleteToken(ctx, oneTimeToken.Token)
		})
		if err = u.handleDBError(err); err != nil {
			u.log.WithError(err)
			return cherry.ErrInvalidLink()
		}
	}
	return nil
}
