package impl

import (
	"context"

	"strings"
	"time"

	"database/sql"

	"fmt"

	"git.containerum.net/ch/auth/proto"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	"git.containerum.net/ch/user-manager/pkg/db"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
	"git.containerum.net/ch/user-manager/pkg/utils"

	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"github.com/lib/pq"
)

func (u *serverImpl) CreateUser(ctx context.Context, request umtypes.RegisterRequest) (*umtypes.UserLogin, error) {
	u.log.WithField("login", request.Login).Info("creating user")
	if err := u.checkReCaptcha(ctx, request.ReCaptcha); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrInvalidRecaptcha()
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

	user, err := u.svc.DB.GetAnyUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateUser()
	}

	reactivatingOldUser := false

	newUser := &db.User{}

	if user != nil {
		if user.IsDeleted {
			//Activating previously partially deleted user
			reactivatingOldUser = true

			newUser = user
			newUser.IsDeleted = false
			newUser.IsActive = false
		} else {
			u.log.WithError(cherry.ErrUserAlreadyExists())
			return nil, cherry.ErrUserAlreadyExists()
		}
	}

	salt := utils.GenSalt(request.Login, request.Login, request.Login) // compatibility with old client db
	passwordHash := utils.GetKey(request.Login, request.Password, salt)
	if !reactivatingOldUser {
		newUser = &db.User{
			Login:        request.Login,
			PasswordHash: passwordHash,
			Salt:         salt,
			Role:         "user",
			IsActive:     false,
			IsDeleted:    false,
		}
	} else {
		newUser.Salt = salt
		newUser.PasswordHash = passwordHash
	}

	var link *db.Link

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		if !reactivatingOldUser {
			if createErr := tx.CreateUser(ctx, newUser); createErr != nil {
				return err
			}

			if createErr := tx.CreateProfile(ctx, &db.Profile{
				User:      newUser,
				Referral:  sql.NullString{String: request.Referral, Valid: true},
				Access:    sql.NullString{String: "rw", Valid: true},
				CreatedAt: pq.NullTime{Time: time.Now().UTC(), Valid: true},
			}); createErr != nil {
				return err
			}
		} else {
			if createErr := tx.UpdateUser(ctx, newUser); createErr != nil {
				return err
			}
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

	return &umtypes.UserLogin{
		ID:    newUser.ID,
		Login: newUser.Login,
	}, nil
}

func (u *serverImpl) ActivateUser(ctx context.Context, request umtypes.Link) (*authProto.CreateTokenResponse, error) {
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

	var tokens *authProto.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		// TODO: send request to resource manager
		return tx.BlacklistUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistUser()
	}

	_, err = u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
		UserId: user.ID,
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		user.IsDeleted = true
		if updErr := tx.UpdateUser(ctx, user); updErr != nil {
			return updErr
		}

		// TODO: send request to billing manager

		_, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
			UserId: user.ID,
		})
		return authErr
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	if err := u.svc.ResourceServiceClient.DeleteUserNamespaces(ctx, user); err != nil {
		u.log.WithError(err)
	}
	if err := u.svc.ResourceServiceClient.DeleteUserVolumes(ctx, user); err != nil {
		u.log.WithError(err)
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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
