package impl

import (
	"context"

	"fmt"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"git.containerum.net/ch/user-manager/pkg/clients"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
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
		return cherry.ErrUnableBindAccount()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	resource, exist := clients.OAuthClientByResource(request.Resource)
	if !exist {
		u.log.WithError(fmt.Errorf(resourceNotSupported, request.Resource))
		return cherry.ErrUnableBindAccount().AddDetailsErr(fmt.Errorf(resourceNotSupported, request.Resource))
	}
	info, err := resource.GetUserInfo(ctx, request.AccessToken)
	if err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBindAccount()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BindAccount(ctx, user, umtypes.OAuthResource(request.Resource), info.UserID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBindAccount()
	}
	return nil
}

func (u *serverImpl) GetBoundAccounts(ctx context.Context) (map[string]string, error) {
	userID := server.MustGetUserID(ctx)

	u.log.WithField("userId", userID).Infof("getting bound accounts")

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUserInfo()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return nil, err
	}

	var accounts *models.Accounts
	accounts, err = u.svc.DB.GetUserBoundAccounts(ctx, user)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUserInfo()
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

	return accs, nil
}

func (u *serverImpl) DeleteBoundAccount(ctx context.Context, request umtypes.BoundAccountDeleteRequest) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("userId", userID).WithFields(logrus.Fields{
		"resource": request.Resource,
	}).Infof("deleting bound account: %#v", request)

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return err
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
		return err
	}
	return nil
}
