package impl

import (
	"context"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"github.com/sirupsen/logrus"
)

func (u *serverImpl) AddBoundAccount(ctx context.Context, request umtypes.OAuthLoginRequest) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithFields(logrus.Fields{
		"userID":       userID,
		"resource":     request.Resource,
		"access_token": request.AccessToken,
	}).Infof("adding bound account: %#v", request)

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return err
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BindAccount(ctx, user, umtypes.OAuthResource(request.Resource), request.AccessToken)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return bindAccountFailed
	}
	return nil
}

func (u *serverImpl) GetBoundAccounts(ctx context.Context) (*umtypes.BoundAccountsResponce, error) {
	userID := server.MustGetUserID(ctx)

	u.log.WithField("userId", userID).Infof("getting bound accounts")

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return nil, err
	}

	var accounts *models.Accounts
	accounts, err = u.svc.DB.GetUserBoundAccounts(ctx, user)
	if err != nil {
		u.log.WithError(err)
		return nil, boundAccountsGetFailed
	}

	accs := make(map[string]string)

	if accounts != nil {
		if accounts.Google.String != "" {
			accs["google"] = accounts.Google.String
		}
		if accounts.Facebook.String != "" {
			accs["facebook"] = accounts.Facebook.String
		}
		if accounts.Github.String != "" {
			accs["github"] = accounts.Github.String
		}
	}

	return &umtypes.BoundAccountsResponce{Accounts: accs}, nil
}

func (u *serverImpl) DeleteBoundAccount(ctx context.Context, request umtypes.BoundAccountDeleteRequest) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("userId", userID).WithFields(logrus.Fields{
		"resource": request.Resource,
	}).Infof("deleting bound account: %#v", request)

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return err
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.DeleteBoundAccount(ctx, user, umtypes.OAuthResource(request.Resource))
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return boundAccountsDeleteFailed
	}
	return nil
}
