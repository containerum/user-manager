package impl

import (
	"context"

	"strings"
	"time"

	"database/sql"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"git.containerum.net/ch/user-manager/pkg/validation"

	"fmt"

	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"github.com/lib/pq"
)

func (u *serverImpl) CreateUser(ctx context.Context, request umtypes.RegisterRequest) (*umtypes.User, error) {
	u.log.WithField("login", request.Login).Info("creating user")
	if err := u.checkReCaptcha(ctx, request.ReCaptcha); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrInvalidRecaptcha()
	}

	errs := validation.ValidateUserCreateRequest(request)
	if errs != nil {
		return nil, cherry.ErrRequestValidationFailed().AddDetailsErr(errs...)
	}

	domain := strings.Split(request.Login, "@")[1]
	blacklisted, err := u.svc.DB.IsDomainBlacklisted(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}
	if blacklisted {
		u.log.WithError(fmt.Errorf(domainInBlacklist, domain))
		return nil, cherry.ErrUnableCreateUser().AddDetailsErr(fmt.Errorf(domainInBlacklist, domain))
	}

	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}
	if user != nil {
		u.log.WithError(cherry.ErrUserAlreadyExists())
		return nil, cherry.ErrUserAlreadyExists()
	}

	salt := utils.GenSalt(request.Login, request.Login, request.Login) // compatibility with old client db
	passwordHash := utils.GetKey(request.Login, request.Password, salt)
	newUser := &models.User{
		Login:        request.Login,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         "user",
		IsActive:     false,
		IsDeleted:    false,
	}

	var link *models.Link

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		if createErr := tx.CreateUser(ctx, newUser); createErr != nil {
			return err
		}

		if createErr := tx.CreateProfile(ctx, &models.Profile{
			User:      newUser,
			Referral:  sql.NullString{String: request.Referral, Valid: true},
			Access:    sql.NullString{String: "rw", Valid: true},
			CreatedAt: pq.NullTime{Time: time.Now().UTC(), Valid: true},
		}); createErr != nil {
			return err
		}

		link, err = tx.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, newUser)
		if err != nil {
			return err
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableCreateUser()
	}

	go u.linkSend(ctx, link)

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			ID:    newUser.ID,
			Login: newUser.Login,
		},
		IsActive: newUser.IsActive,
	}, nil
}

func (u *serverImpl) ActivateUser(ctx context.Context, request umtypes.Link) (*auth.CreateTokenResponse, error) {
	u.log.Info("activating user")
	u.log.WithField("link", request.Link).Debugln("activating user details")
	link, err := u.svc.DB.GetLinkFromString(ctx, request.Link)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableActivate()
	}
	if link == nil {
		u.log.WithError(fmt.Errorf(linkNotFound, request.Link))
		return nil, cherry.ErrInvalidLink().AddDetailsErr(fmt.Errorf(linkNotFound, request.Link))
	} else if link.Type != umtypes.LinkTypeConfirm {
		u.log.WithError(fmt.Errorf(linkNotFound, request.Link))
		return nil, cherry.ErrInvalidLink().AddDetailsErr(fmt.Errorf(linkNotFound, request.Link))
	}

	var tokens *auth.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		link.User.IsActive = true
		if updErr := tx.UpdateUser(ctx, link.User); updErr != nil {
			return cherry.ErrUnableActivate()
		}
		link.IsActive = false
		if updErr := tx.UpdateLink(ctx, link); updErr != nil {
			return cherry.ErrUnableActivate()
		}

		// TODO: send request to billing manager

		var err error
		tokens, err = u.createTokens(ctx, link.User)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableActivate()
	}

	go func() {
		err := u.svc.MailClient.SendActivationMail(ctx, &mttypes.Recipient{
			ID:        link.User.ID,
			Name:      link.User.Login,
			Email:     link.User.Login,
			Variables: map[string]interface{}{},
		})
		if err != nil {
			u.log.WithError(err).Error("activation email send failed")
		}
	}()

	return tokens, nil
}

func (u *serverImpl) BlacklistUser(ctx context.Context, request umtypes.UserLogin) error {
	u.log.WithField("user_id", request.ID).Info("blacklisting user")

	userID := server.MustGetUserID(ctx)
	if request.ID == userID {
		return cherry.ErrRequestValidationFailed().AddDetails(blacklistYourself)
	}

	user, err := u.svc.DB.GetUserByID(ctx, request.ID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistUser()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return err
	}
	if user.Role == "admin" {
		return cherry.ErrRequestValidationFailed().AddDetails(blacklistAdmin)
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		// TODO: send request to resource manager
		return tx.BlacklistUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistUser()
	}

	_, err = u.svc.AuthClient.DeleteUserTokens(ctx, &auth.DeleteUserTokensRequest{
		UserId: &common.UUID{Value: user.ID},
	})
	if err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistUser()
	}

	go func() {
		err := u.svc.MailClient.SendBlockedMail(ctx, &mttypes.Recipient{
			ID:    user.ID,
			Name:  user.Login,
			Email: user.Login,
		})
		if err != nil {
			u.log.WithError(err).Error("email send failed")
		}
	}()

	return nil
}

func (u *serverImpl) UnBlacklistUser(ctx context.Context, request umtypes.UserLogin) error {
	u.log.WithField("user_id", request.ID).Info("unblacklisting user")

	user, err := u.svc.DB.GetUserByID(ctx, request.ID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUnblacklistUser()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return cherry.ErrUserNotExist()
	}
	if user.IsInBlacklist != true {
		u.log.WithError(cherry.ErrUserNotBlacklisted())
		return cherry.ErrUserNotBlacklisted()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		// TODO: send request to resource manager
		return tx.UnBlacklistUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		return cherry.ErrUnableUnblacklistUser()
	}

	go func() {
		err := u.svc.MailClient.SendUnBlockedMail(ctx, &mttypes.Recipient{
			ID:    user.ID,
			Name:  user.Login,
			Email: user.Login,
		})
		if err != nil {
			u.log.WithError(err).Error("email send failed")
		}
	}()

	return nil
}

func (u *serverImpl) UpdateUser(ctx context.Context, newData map[string]interface{}) (*umtypes.User, error) {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("updating user profile data")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil || profile == nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		profile.Data = newData
		return tx.UpdateProfile(ctx, profile)
	})
	if err = u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			ID:    user.ID,
			Login: user.Login,
		},
		Profile: &umtypes.Profile{
			Data:      profile.Data,
			CreatedAt: profile.CreatedAt.Time.String(),
		},
		Role:     user.Role,
		IsActive: user.IsActive,
	}, err
}

func (u *serverImpl) PartiallyDeleteUser(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("partially deleting user")

	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return cherry.ErrUserNotExist()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		user.IsDeleted = true
		if updErr := tx.UpdateUser(ctx, user); updErr != nil {
			return updErr
		}

		// TODO: send request to billing manager

		_, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &auth.DeleteUserTokensRequest{
			UserId: &common.UUID{Value: user.ID},
		})
		return authErr
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	go func() {
		err := u.svc.MailClient.SendAccDeletedMail(ctx, &mttypes.Recipient{
			ID:        user.ID,
			Name:      user.Login,
			Email:     user.Login,
			Variables: map[string]interface{}{},
		})
		if err != nil {
			u.log.WithError(err).Error("delete account email send failed")
		}
	}()

	return nil
}

func (u *serverImpl) CompletelyDeleteUser(ctx context.Context, userID string) error {
	u.log.WithField("user_id", userID).Info("completely deleting user")
	user, err := u.svc.DB.GetAnyUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return cherry.ErrUserNotExist()
	}
	if !user.IsDeleted {
		u.log.WithError(cherry.ErrUnableDeleteUser())
		return cherry.ErrUnableDeleteUser()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		add, rngErr := utils.SecureRandomString(6)
		if rngErr != nil {
			return rngErr
		}
		user.Login = user.Login + "-" + add
		// TODO: send request to billing manager
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	return nil
}

func (u *serverImpl) CreateUserWebAPI(ctx context.Context, userName string, password string, id string, createdAtStr string, data map[string]interface{}) (*umtypes.User, error) {
	u.log.WithField("login", userName).Info("creating user from old api")

	domain := strings.Split(userName, "@")[1]
	blacklisted, err := u.svc.DB.IsDomainBlacklisted(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}
	if blacklisted {
		u.log.WithError(fmt.Errorf(domainInBlacklist, domain))
		return nil, cherry.ErrUnableCreateUser().AddDetailsErr(fmt.Errorf(domainInBlacklist, domain))
	}

	user, err := u.svc.DB.GetUserByLogin(ctx, userName)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}
	if user != nil {
		u.log.WithError(cherry.ErrUserAlreadyExists())
		return nil, cherry.ErrUserAlreadyExists()
	}

	salt := utils.GenSalt(userName, userName, userName) // compatibility with old client db
	passwordHash := utils.GetKey(userName, password, salt)
	newUser := &models.User{
		ID:           id,
		Login:        userName,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         "user",
		IsActive:     true,
		IsDeleted:    false,
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		if createErr := tx.CreateUserWebAPI(ctx, newUser); createErr != nil {
			u.log.WithError(createErr)
			return createErr
		}

		var createdAt time.Time
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			u.log.WithError(err).Warnf("Error parsing time")
			createdAt = time.Now().UTC()
		}

		if createErr := tx.CreateProfile(ctx, &models.Profile{
			User:      newUser,
			Access:    sql.NullString{String: "rw", Valid: true},
			CreatedAt: pq.NullTime{Time: createdAt, Valid: true},
			Data:      data,
		}); createErr != nil {
			return createErr
		}

		return nil
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, err
	}

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			ID:    newUser.ID,
			Login: newUser.Login,
		},
		IsActive: newUser.IsActive,
	}, nil
}
