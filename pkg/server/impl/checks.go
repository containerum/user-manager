package impl

import (
	"context"

	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	"github.com/containerum/utils/httputil"
)

func (u *serverImpl) CheckUserExist(ctx context.Context) error {
	userID := httputil.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("checking if user exists")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrPermissionsError()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	return nil
}

func (u *serverImpl) CheckAdmin(ctx context.Context) error {
	userID := httputil.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("checking if user is admin")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrPermissionsError()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	if user.Role != "admin" {
		u.log.WithError(cherry.ErrAdminRequired())
		return cherry.ErrAdminRequired()
	}

	return nil
}
