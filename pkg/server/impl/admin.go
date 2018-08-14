package impl

import (
	"context"
	"database/sql"
	"time"

	"git.containerum.net/ch/auth/proto"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	m "git.containerum.net/ch/user-manager/pkg/router/middleware"
	cherry "git.containerum.net/ch/user-manager/pkg/umerrors"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"git.containerum.net/ch/user-manager/pkg/validation"
	"github.com/containerum/utils/httputil"
	"github.com/lib/pq"
)

func (u *serverImpl) AdminCreateUser(ctx context.Context, request models.UserLogin) (*models.UserLogin, error) {
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
		Role:         m.RoleUser,
		IsActive:     true,
		IsDeleted:    false,
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		if createErr := tx.CreateUser(ctx, newUser); createErr != nil {
			return createErr
		}

		if createErr := tx.CreateProfile(ctx, &db.Profile{
			User:      newUser,
			Access:    sql.NullString{String: "rw", Valid: true},
			CreatedAt: pq.NullTime{Time: time.Now().UTC(), Valid: true},
		}); createErr != nil {
			return createErr
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableCreateUser()
	}

	return &models.UserLogin{
		ID:       newUser.ID,
		Login:    newUser.Login,
		Password: password,
	}, nil
}

func (u *serverImpl) AdminActivateUser(ctx context.Context, request models.UserLogin) error {
	u.log.Info("activating user (admin)")

	user, err := u.svc.DB.GetAnyUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableActivate()
	}
	if user.IsDeleted {
		return cherry.ErrInvalidLogin()
	}

	if user.IsActive {
		return cherry.ErrUserAlreadyActivated()
	}

	user.IsActive = true
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		return cherry.ErrUnableActivate()
	}

	return nil
}

func (u *serverImpl) AdminDeactivateUser(ctx context.Context, request models.UserLogin) error {
	u.log.Info("deactivating user (admin)")

	user, err := u.svc.DB.GetAnyUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	if err := u.loginUserChecks(user); err != nil {
		return err
	}

	if user.ID == httputil.MustGetUserID(ctx) {
		return cherry.ErrChangeOwnPermissions()
	}

	user.IsDeleted = true
	user.IsActive = false
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	if _, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
		UserId: user.ID,
	}); authErr != nil {
		return authErr
	}

	if err := u.svc.PermissionsClient.DeleteUserNamespaces(ctx, user); err != nil {
		u.log.WithError(err)
	}
	return nil
}

func (u *serverImpl) AdminResetPassword(ctx context.Context, request models.UserLogin) (*models.UserLogin, error) {
	u.log.Info("reseting user password (admin)")

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableChangePassword()
	}
	if err := u.loginUserChecks(user); err != nil {
		return nil, err
	}

	password, err := utils.SecureRandomString(10)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableChangePassword()
	}

	user.PasswordHash = utils.GetKey(user.Login, password, user.Salt)
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err = u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableChangePassword()
	}

	if _, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
		UserId: user.ID,
	}); authErr != nil {
		return nil, authErr
	}

	return &models.UserLogin{
		ID:       user.ID,
		Login:    user.Login,
		Password: password,
	}, nil
}

func (u *serverImpl) AdminSetAdmin(ctx context.Context, request models.UserLogin) error {
	u.log.Info("giving admin permissions to user (admin)")

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUpdateUserInfo()
	}
	if err := u.loginUserChecks(user); err != nil {
		return err
	}

	user.Role = m.RoleAdmin
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err = u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUpdateUserInfo()
	}

	return nil
}

func (u *serverImpl) AdminUnsetAdmin(ctx context.Context, request models.UserLogin) error {
	u.log.Info("removing admin permissions from user (admin)")

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUpdateUserInfo()
	}
	if err := u.loginUserChecks(user); err != nil {
		return err
	}
	if user.ID == httputil.MustGetUserID(ctx) {
		return cherry.ErrChangeOwnPermissions()
	}

	user.Role = m.RoleUser
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err = u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUpdateUserInfo()
	}

	return nil
}

func (u *serverImpl) CreateFirstAdmin(password string) error {
	u.log.Info("creating first admin user")

	user, err := u.svc.DB.GetAnyUserByLoginWOContext("admin@local.containerum.io")
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableCreateUser()
	}

	if user != nil {
		u.log.Info("updating admin password")
		user.PasswordHash = utils.GetKey("admin@local.containerum.io", password, user.Salt)
		err = u.svc.DB.UpdateUserWOContext(user)
		if err != nil {
			return err
		}
		u.log.Info("admin password updated")
		return nil
	}

	salt := utils.GenSalt("admin@local.containerum.io", "admin@local.containerum.io", "admin@local.containerum.io") // compatibility with old client db
	passwordHash := utils.GetKey("admin@local.containerum.io", password, salt)
	newUser := &db.User{
		Login:        "admin@local.containerum.io",
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         m.RoleAdmin,
		IsActive:     true,
		IsDeleted:    false,
	}

	if createErr := u.svc.DB.CreateUserWOContext(newUser); createErr != nil {
		return createErr
	}

	if createErr := u.svc.DB.CreateProfileWOContext(&db.Profile{
		User:      newUser,
		Access:    sql.NullString{String: "rw", Valid: true},
		CreatedAt: pq.NullTime{Time: time.Now().UTC(), Valid: true},
	}); createErr != nil {
		return createErr
	}
	u.log.Info("admin created")
	return nil
}
