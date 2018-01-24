package impl

import (
	"context"

	"git.containerum.net/ch/user-manager/server"
	"github.com/pkg/errors"
)

func (u *serverImpl) CheckUserExist(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("check if exists")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}

	if user == nil {
		return errors.New(userNotFound)
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

	if user == nil {
		return errors.New(userNotFound)
	}

	if user.Role != "admin" {
		return errors.New(adminRequired)
	}

	return nil
}
