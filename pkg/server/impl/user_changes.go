package impl

import (
	"context"

	"strings"
	"time"

	"database/sql"

	"fmt"

	"git.containerum.net/ch/auth/proto"
	mttypes "git.containerum.net/ch/mail-templater/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/utils"
	"github.com/containerum/utils/httputil"

	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	"github.com/lib/pq"
)

func (u *serverImpl) CreateUser(ctx context.Context, request models.RegisterRequest) (*models.UserLogin, error) {
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

		link, err = tx.CreateLink(ctx, models.LinkTypeConfirm, 24*time.Hour, newUser)
		if err != nil {
			return err
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableCreateUser()
	}

	go u.linkSend(ctx, link)

	if u.svc.TelegramClient != nil {
		err := u.svc.TelegramClient.SendRegistrationMessage(ctx, link.User.Login)
		if err != nil {
			u.log.WithError(err).Debug("telegram message send failed")
		}
	}

	return &models.UserLogin{
		ID:    newUser.ID,
		Login: newUser.Login,
	}, nil
}

func (u *serverImpl) ActivateUser(ctx context.Context, request models.Link) (*authProto.CreateTokenResponse, error) {
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
	} else if link.Type != models.LinkTypeConfirm {
		u.log.WithError(fmt.Errorf(linkNotFound, request.Link))
		return nil, cherry.ErrInvalidLink().AddDetailsErr(fmt.Errorf(linkNotFound, request.Link))
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		link.User.IsActive = true
		if updErr := tx.UpdateUser(ctx, link.User); updErr != nil {
			return cherry.ErrUnableActivate()
		}
		link.IsActive = false
		if updErr := tx.UpdateLink(ctx, link); updErr != nil {
			return cherry.ErrUnableActivate()
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableActivate()
	}

	tokens, err := u.createTokens(ctx, link.User)
	if err != nil {
		return nil, err
	}

	go func() {
		err := u.svc.MailClient.SendActivationMail(ctx, &mttypes.Recipient{
			ID:    link.User.ID,
			Name:  link.User.Login,
			Email: link.User.Login,
		})
		if err != nil {
			u.log.WithError(err).Error("activation email send failed")
		}
	}()

	if u.svc.TelegramClient != nil {
		err := u.svc.TelegramClient.SendActivationMessage(ctx, link.User.Login)
		if err != nil {
			u.log.WithError(err).Debug("telegram message send failed")
		}
	}

	return tokens, nil
}

func (u *serverImpl) BlacklistUser(ctx context.Context, request models.UserLogin) error {
	u.log.WithField("user_id", request.ID).Info("blacklisting user")

	userID := httputil.MustGetUserID(ctx)
	if request.ID == userID {
		return cherry.ErrRequestValidationFailed().AddDetails(blacklistYourself)
	}

	user, err := u.svc.DB.GetUserByID(ctx, request.ID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistUser()
	}
	if err := u.loginUserChecks(user); err != nil {
		u.log.WithError(err)
		return err
	}
	if user.Role == "admin" {
		return cherry.ErrRequestValidationFailed().AddDetails(blacklistAdmin)
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.DeleteGroupMemberFromAllGroups(ctx, user.ID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
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

func (u *serverImpl) UnBlacklistUser(ctx context.Context, request models.UserLogin) error {
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
	if !user.IsInBlacklist {
		u.log.WithError(cherry.ErrUserNotBlacklisted())
		return cherry.ErrUserNotBlacklisted()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
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

func (u *serverImpl) UpdateUser(ctx context.Context, newData map[string]interface{}) (*models.User, error) {
	userID := httputil.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("updating user profile data")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}
	if err := u.loginUserChecks(user); err != nil {
		u.log.WithError(err)
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil || profile == nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}

	profile.Data = newData
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateProfile(ctx, profile)
	})
	if err = u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableUpdateUserInfo()
	}

	return &models.User{
		UserLogin: &models.UserLogin{
			ID:    user.ID,
			Login: user.Login,
		},
		Profile: &models.Profile{
			Data:      profile.Data,
			CreatedAt: profile.CreatedAt.Time.Format(time.RFC3339),
		},
		Role:     user.Role,
		IsActive: user.IsActive,
	}, err
}

func (u *serverImpl) PartiallyDeleteUser(ctx context.Context) error {
	userID := httputil.MustGetUserID(ctx)
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

	user.IsDeleted = true
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	_, authErr := u.svc.AuthClient.DeleteUserTokens(ctx, &authProto.DeleteUserTokensRequest{
		UserId: user.ID,
	})
	if authErr != nil {
		return authErr
	}

	if err := u.svc.PermissionsClient.DeleteUserNamespaces(ctx, user); err != nil {
		u.log.WithError(err)
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.DeleteGroupMemberFromAllGroups(ctx, user.ID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	go func() {
		err := u.svc.MailClient.SendAccDeletedMail(ctx, &mttypes.Recipient{
			ID:    user.ID,
			Name:  user.Login,
			Email: user.Login,
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
		return tx.DeleteGroupMemberFromAllGroups(ctx, user.ID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}

	add, rngErr := utils.SecureRandomString(6)
	if rngErr != nil {
		return rngErr
	}
	user.Login = user.Login + "-" + add
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteUser()
	}
	return nil
}
