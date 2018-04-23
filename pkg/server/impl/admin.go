package impl

import (
	"context"
	"database/sql"
	"time"

	"git.containerum.net/ch/auth/proto"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"git.containerum.net/ch/user-manager/pkg/db"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/lib/pq"
)

func (u *serverImpl) AdminCreateUser(ctx context.Context, request umtypes.UserLogin) (*umtypes.UserLogin, error) {
	u.log.WithField("login", request.Login).Info("creating user (admin)")

	password, err := utils.SecureRandomString(10)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}

	errs := validation.ValidateUserLogin(request)
	if errs != nil {
		return nil, cherry.ErrRequestValidationFailed().AddDetailsErr(errs...)
	}

	user, err := u.svc.DB.GetAnyUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}

	if user != nil {
		return nil, cherry.ErrUserAlreadyExists()
	}

	salt := utils.GenSalt(request.Login, request.Login, request.Login) // compatibility with old client db
	passwordHash := utils.GetKey(request.Login, password, salt)
	newUser := &db.User{
		Login:        request.Login,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         "user",
		IsActive:     true,
		IsDeleted:    false,
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		if createErr := tx.CreateUser(ctx, newUser); createErr != nil {
			return err
		}

		if createErr := tx.CreateProfile(ctx, &db.Profile{
			User:      newUser,
			Access:    sql.NullString{String: "rw", Valid: true},
			CreatedAt: pq.NullTime{Time: time.Now().UTC(), Valid: true},
		}); createErr != nil {
			return err
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableCreateUser()
	}

	return &umtypes.UserLogin{
		ID:       newUser.ID,
		Login:    newUser.Login,
		Password: password,
	}, nil
}

func (u *serverImpl) AdminActivateUser(ctx context.Context, request umtypes.UserLogin) (*authProto.CreateTokenResponse, error) {
	u.log.Info("activating user (admin)")

	var tokens *authProto.CreateTokenResponse

	user, err := u.svc.DB.GetAnyUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableActivate()
	}

	if user.IsActive {
		return nil, cherry.ErrUserAlreadyActivated()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		user.IsActive = true
		if updErr := tx.UpdateUser(ctx, user); updErr != nil {
			u.log.WithError(updErr)
			return cherry.ErrUnableActivate()
		}

		var err error
		tokens, err = u.createTokens(ctx, user)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableActivate()
	}

	return tokens, nil
}

func (u *serverImpl) AdminDeactivateUser(ctx context.Context, request umtypes.UserLogin) error {
	u.log.Info("deactivating user (admin)")

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return cherry.ErrUserNotExist()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		user.IsDeleted = true
		if updErr := tx.UpdateUser(ctx, user); updErr != nil {
			u.log.WithError(updErr)
			return cherry.ErrUnableDeleteUser()
		}

	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableActivate()
	}

	return tokens, nil
}
