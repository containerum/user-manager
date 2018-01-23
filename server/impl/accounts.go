package impl

import (
	"context"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"github.com/sirupsen/logrus"
)

func (u *serverImpl) AddBoundAccount(ctx context.Context, request umtypes.OAuthLoginRequest) error {
	u.log.WithFields(logrus.Fields{
		"resource":     request.Resource,
		"access_token": request.AccessToken,
	}).Infof("Adding bound account: %#v", request)

	userID := server.MustGetUserID(ctx)
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BindAccount(ctx, user, umtypes.OAuthResource(request.Resource), request.AccessToken)
	})
	if err := u.handleDBError(err); err != nil {
		return err
	}
	return nil
}

func (u *serverImpl) GetBoundAccounts(ctx context.Context) (*umtypes.BoundAccountsResponce, error) {
	u.log.Infof("Getting bound accounts")

	userID := server.MustGetUserID(ctx)
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, err
	}

	var accounts *models.Accounts
	accounts, err = u.svc.DB.GetUserBoundAccounts(ctx, user)
	if err != nil {
		return nil, err
	}

	accs := make(map[string]string)

	if accounts.Google.String != "" {
		accs["google"] = accounts.Google.String
	}
	if accounts.Facebook.String != "" {
		accs["facebook"] = accounts.Facebook.String
	}
	if accounts.Github.String != "" {
		accs["github"] = accounts.Github.String
	}

	return &umtypes.BoundAccountsResponce{accs}, nil
}

func (u *serverImpl) DeleteBoundAccount(ctx context.Context, request umtypes.BoundAccountDeleteRequest) error {
	u.log.WithFields(logrus.Fields{
		"resource": request.Resource,
	}).Infof("Deleting bound account: %#v", request)

	userID := server.MustGetUserID(ctx)
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.DeleteBoundAccount(ctx, user, umtypes.OAuthResource(request.Resource))
	})
	if err := u.handleDBError(err); err != nil {
		return err
	}
	return nil
}
