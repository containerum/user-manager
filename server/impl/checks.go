package impl

import (
	"context"

	"git.containerum.net/ch/user-manager/server"
)

func (u *serverImpl) CheckUserExist(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("check if user exists")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	return nil
}

func (u *serverImpl) CheckAdmin(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("check if admin")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	if user.Role != "admin" {
		return adminRequired
	}

	return nil
}
