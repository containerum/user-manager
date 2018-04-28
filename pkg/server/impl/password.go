package impl

import (
	"context"

	"time"

	"fmt"

	"git.containerum.net/ch/auth/proto"
	mttypes "git.containerum.net/ch/mail-templater/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"github.com/containerum/utils/httputil"
)

func (u *serverImpl) ChangePassword(ctx context.Context, request models.PasswordRequest) (*authProto.CreateTokenResponse, error) {
	userID := httputil.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("changing password")

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableChangePassword()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, err
	}

	if !utils.CheckPassword(user.Login, request.CurrentPassword, user.Salt, user.PasswordHash) {
		u.log.WithError(cherry.ErrInvalidLogin())
		return nil, cherry.ErrInvalidLogin()
	}

	var tokens *authProto.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		user.PasswordHash = utils.GetKey(user.Login, request.NewPassword, user.Salt)
		if updErr := tx.UpdateUser(ctx, user); updErr != nil {
			return updErr
		}

		_, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
			UserId: user.ID,
		})
		if authErr != nil {
			return authErr
		}

		tokens, authErr = u.createTokens(ctx, user)
		return authErr
	})
	if err = u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableChangePassword()
	}
	go func() {
		mailErr := u.svc.MailClient.SendPasswordChangedMail(ctx, &mttypes.Recipient{
			ID:        user.ID,
			Name:      user.Login,
			Email:     user.Login,
			Variables: map[string]interface{}{},
		})
		if mailErr != nil {
			u.log.WithError(mailErr).Error("password change email send failed")
		}
	}()

	return tokens, nil
}

func (u *serverImpl) ResetPassword(ctx context.Context, request models.UserLogin) error {
	u.log.WithField("login", request.Login).Info("resetting password")
	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)

	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableResetPassword()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	var link *db.Link
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		link, err = tx.CreateLink(ctx, models.LinkTypePwdChange, 24*time.Hour, user)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableResetPassword()
	}

	go func() {
		err := u.svc.MailClient.SendPasswordResetMail(ctx, &mttypes.Recipient{
			ID:        user.ID,
			Name:      user.Login,
			Email:     user.Login,
			Variables: map[string]interface{}{"TOKEN": link.Link},
		})
		if err != nil {
			u.log.WithError(err).Error("password reset email send failed")
		}
	}()

	return nil
}

func (u *serverImpl) RestorePassword(ctx context.Context, request models.PasswordRequest) (*authProto.CreateTokenResponse, error) {
	u.log.Info("restoring password")
	u.log.WithField("link", request.Link).Debug("restoring password details")

	link, err := u.svc.DB.GetLinkFromString(ctx, request.Link)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableResetPassword()
	}
	if link == nil {
		u.log.WithError(fmt.Errorf(linkNotFound, request.Link))
		return nil, cherry.ErrInvalidLink().AddDetailsErr(fmt.Errorf(linkNotFound, request.Link))
	}
	if link.Type != models.LinkTypePwdChange {
		u.log.WithError(fmt.Errorf(linkNotFound, request.Link))
		return nil, cherry.ErrInvalidLink().AddDetailsErr(fmt.Errorf(linkNotFound, request.Link))
	}

	var tokens *authProto.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		link.User.PasswordHash = utils.GetKey(link.User.Login, request.NewPassword, link.User.Salt)
		if updErr := tx.UpdateUser(ctx, link.User); updErr != nil {
			return updErr
		}
		link.IsActive = false

		_, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
			UserId: link.User.ID,
		})
		if authErr != nil {
			return authErr
		}

		if updErr := tx.UpdateLink(ctx, link); updErr != nil {
			return updErr
		}

		tokens, authErr = u.createTokens(ctx, link.User)
		return authErr
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableResetPassword()
	}

	go func() {
		err := u.svc.MailClient.SendPasswordChangedMail(ctx, &mttypes.Recipient{
			ID:        link.User.ID,
			Name:      link.User.Login,
			Email:     link.User.Login,
			Variables: map[string]interface{}{},
		})
		if err != nil {
			u.log.WithError(err).Error("password changed email send failed")
		}
	}()

	return tokens, nil
}
