package impl

import (
	"context"

	"strings"
	"time"

	"math/rand"
	"strconv"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"git.containerum.net/ch/user-manager/utils"
)

func (u *serverImpl) CreateUser(ctx context.Context, request umtypes.UserCreateRequest) (*umtypes.UserCreateResponse, error) {
	u.log.WithField("login", request.UserName).Info("creating user")
	if err := u.checkReCaptcha(ctx, request.ReCaptcha); err != nil {
		return nil, err
	}

	domain := strings.Split(request.UserName, "@")[1]
	blacklisted, err := u.svc.DB.IsDomainBlacklisted(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		return nil, blacklistDomainCheckFailed
	}
	if blacklisted {
		return nil, &server.AccessDeniedError{Err: errors.Format(domainInBlacklist, domain)}
	}

	user, err := u.svc.DB.GetUserByLogin(ctx, request.UserName)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if user != nil {
		return nil, &server.AlreadyExistsError{Err: errors.Format(userAlreadyExists, request.UserName)}
	}

	salt := utils.GenSalt(request.UserName, request.UserName, request.UserName) // compatibility with old client db
	passwordHash := utils.GetKey(request.UserName, request.Password, salt)
	newUser := &models.User{
		Login:        request.UserName,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         umtypes.RoleUser,
		IsActive:     false,
		IsDeleted:    false,
	}

	var link *models.Link
	var tokens *auth.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		if err := tx.CreateUser(ctx, newUser); err != nil {
			return userCreateFailed
		}

		if err := tx.CreateProfile(ctx, &models.Profile{
			User:      newUser,
			Referral:  request.Referral,
			Access:    "rw",
			CreatedAt: time.Now().UTC(),
		}); err != nil {
			return profileCreateFailed
		}

		link, err = tx.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, newUser)
		if err != nil {
			return linkCreateFailed
		}

		tokens, err = u.createTokens(ctx, user)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, err
	}

	go u.linkSend(ctx, link)

	return &umtypes.UserCreateResponse{
		ID:       user.ID,
		Login:    user.Login,
		IsActive: user.IsActive,
	}, nil
}

func (u *serverImpl) ActivateUser(ctx context.Context, request umtypes.ActivateRequest) (*auth.CreateTokenResponse, error) {
	u.log.WithField("link", request.Link).Info("activating user")
	link, err := u.svc.DB.GetLinkFromString(ctx, request.Link)
	if err := u.handleDBError(err); err != nil {
		return nil, linkGetFailed
	}
	if link == nil {
		return nil, &server.NotFoundError{Err: errors.Format(linkNotFound, request.Link)}
	}
	if link.Type != umtypes.LinkTypeConfirm {
		return nil, &server.AccessDeniedError{Err: errors.Format(linkNotForConfirm, request.Link)}
	}

	var tokens *auth.CreateTokenResponse

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		link.User.IsActive = true
		if err := tx.UpdateUser(ctx, link.User); err != nil {
			return userUpdateFailed
		}
		link.IsActive = false
		if err := tx.UpdateLink(ctx, link); err != nil {
			return linkUpdateFailed
		}

		// TODO: send request to billing manager

		var err error
		tokens, err = u.createTokens(ctx, link.User)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		return nil, err
	}

	go func() {
		err := u.svc.MailClient.SendActivationMail(ctx, &mttypes.Recipient{
			ID:        link.User.ID,
			Name:      link.User.Login,
			Email:     link.User.Login,
			Variables: map[string]string{},
		})
		if err != nil {
			u.log.WithError(err).Error("activation email send failed")
		}
	}()

	return tokens, nil
}

func (u *serverImpl) BlacklistUser(ctx context.Context, request umtypes.UserToBlacklistRequest) error {
	u.log.WithField("user_id", request.UserID).Info("blacklisting user")
	user, err := u.svc.DB.GetUserByID(ctx, request.UserID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		// TODO: send request to resource manager
		return tx.BlacklistUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		return blacklistUserFailed
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

func (u *serverImpl) UpdateUser(ctx context.Context, newData umtypes.ProfileData) (*umtypes.UserInfoGetResponse, error) {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("updating user profile data")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil || profile == nil {
		return nil, profileGetFailed
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		profile.Data = newData
		return tx.UpdateProfile(ctx, profile)
	})
	err = u.handleDBError(err)
	if err != nil {
		return nil, profileUpdateFailed
	}

	return &umtypes.UserInfoGetResponse{
		Login:     user.Login,
		Data:      profile.Data,
		ID:        user.ID,
		IsActive:  user.IsActive,
		CreatedAt: profile.CreatedAt,
	}, err
}

func (u *serverImpl) PartiallyDeleteUser(ctx context.Context) error {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("partially delete user")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if user == nil {
		return &server.NotFoundError{Err: errors.New(userNotFound)}
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		user.IsDeleted = true
		if err := tx.UpdateUser(ctx, user); err != nil {
			return userUpdateFailed
		}

		// TODO: send request to billing manager

		_, err := u.svc.AuthClient.DeleteUserTokens(ctx, &auth.DeleteUserTokensRequest{
			UserId: &common.UUID{Value: user.ID},
		})
		if err != nil {
			return tokenDeleteFailed
		}
		return nil
	})
	if err := u.handleDBError(err); err != nil {
		return err
	}

	go func() {
		err := u.svc.MailClient.SendAccDeletedMail(ctx, &mttypes.Recipient{
			ID:        user.ID,
			Name:      user.Login,
			Email:     user.Login,
			Variables: map[string]string{},
		})
		if err != nil {
			u.log.WithError(err).Error("delete account email send failed")
		}
	}()

	return nil
}

func (u *serverImpl) CompletelyDeleteUser(ctx context.Context, userID string) error {
	u.log.WithField("user_id", userID).Info("completely delete user")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if user == nil {
		return &server.NotFoundError{Err: errors.New(userNotFound)}
	}
	if !user.IsDeleted {
		return &server.BadRequestError{Err: errors.Format(userNotPartiallyDeleted, userID)}
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		user.Login = user.Login + strconv.Itoa(rand.Int())
		// TODO: send request to billing manager
		return tx.UpdateUser(ctx, user)
	})
	if err := u.handleDBError(err); err != nil {
		return userUpdateFailed
	}
	return nil
}
