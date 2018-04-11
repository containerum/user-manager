package impl

import (
	"context"

	"time"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
)

func (u *serverImpl) GetUserLinks(ctx context.Context, userID string) (*umtypes.Links, error) {
	u.log.WithField("user_id", userID).Info("get user links")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUserLinks()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return nil, cherry.ErrUserNotExist()
	}

	links, err := u.svc.DB.GetUserLinks(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUserLinks()
	}

	resp := umtypes.Links{Links: []umtypes.Link{}}
	for _, v := range links {
		var sentAt time.Time
		if v.SentAt.Valid {
			sentAt = v.SentAt.Time
		}
		resp.Links = append(resp.Links, umtypes.Link{
			Link:      v.Link,
			Type:      v.Type,
			CreatedAt: v.CreatedAt,
			ExpiredAt: v.ExpiredAt,
			IsActive:  v.IsActive,
			SentAt:    sentAt,
		})
	}

	return &resp, nil
}

func (u *serverImpl) GetUserInfo(ctx context.Context) (*umtypes.User, error) {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("get user info")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUserInfo()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return nil, cherry.ErrUserNotExist()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, cherry.ErrUnableGetUserInfo()
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUserInfo()
	}

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			Login: user.Login,
			ID:    user.ID,
		},
		Profile: &umtypes.Profile{
			Data:      profile.Data,
			CreatedAt: profile.CreatedAt.Time.String(),
		},
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}

func (u *serverImpl) GetUserInfoByID(ctx context.Context, userID string) (*umtypes.User, error) {
	u.log.WithField("user_id", userID).Info("get user info by id")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUserInfo()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return nil, cherry.ErrUserNotExist()
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUserInfo()
	}

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			Login: user.Login,
		},
		Profile: &umtypes.Profile{
			Data: profile.Data,
		},
		Role: user.Role,
	}, nil
}

func (u *serverImpl) GetUserInfoByLogin(ctx context.Context, login string) (*umtypes.User, error) {
	u.log.WithField("login", login).Info("get user info by login")
	user, err := u.svc.DB.GetUserByLogin(ctx, login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUserInfo()
	}
	if user == nil {
		u.log.WithError(cherry.ErrUserNotExist())
		return nil, cherry.ErrUserNotExist()
	}
	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUserInfo()
	}

	return &umtypes.User{
		UserLogin: &umtypes.UserLogin{
			ID: user.ID,
		},
		Role: user.Role,
		Profile: &umtypes.Profile{
			Data: profile.Data,
		},
	}, nil
}

func (u *serverImpl) GetBlacklistedUsers(ctx context.Context, page int, perPage int) (*umtypes.UserList, error) {
	u.log.WithField("per_page", perPage).WithField("page", page).Info("get blacklisted users")
	blacklisted, err := u.svc.DB.GetBlacklistedUsers(ctx, perPage, (page-1)*perPage)
	if err := u.handleDBError(err); err != nil {
		return nil, cherry.ErrUnableGetUsersList()
	}
	var resp umtypes.UserList
	for _, v := range blacklisted {
		resp.Users = append(resp.Users, umtypes.User{
			UserLogin: &umtypes.UserLogin{
				Login: v.Login,
				ID:    v.ID,
			},
		})
	}
	return &resp, nil
}

func (u *serverImpl) GetUsers(ctx context.Context, page int, perPage int, filters ...string) (*umtypes.UserList, error) {
	u.log.WithField("per_page", perPage).WithField("page", page).Info("get users")
	profiles, err := u.svc.DB.GetAllProfiles(ctx, perPage, (page-1)*perPage)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUsersList()
	}

	satisfiesFilter := server.CreateFilterFunc(filters...)

	resp := umtypes.UserList{
		Users: []umtypes.User{},
	}
	for _, v := range profiles {
		if !satisfiesFilter(v) {
			continue
		}

		accs := make(map[string]string)

		if v.Accounts.Google.String != "" {
			accs["google"] = v.Accounts.Google.String
		}

		if v.Accounts.Facebook.String != "" {
			accs["facebook"] = v.Accounts.Facebook.String
		}

		if v.Accounts.Github.String != "" {
			accs["github"] = v.Accounts.Github.String
		}

		user := umtypes.User{
			UserLogin: &umtypes.UserLogin{
				ID:    v.User.ID,
				Login: v.User.Login,
			},
			Profile: &umtypes.Profile{
				Access:   v.Profile.Access.String,
				Data:     v.Profile.Data,
				Referral: v.Profile.Referral.String,
			},
			Accounts: &umtypes.Accounts{
				Accounts: accs,
			},
			Role:          v.User.Role,
			IsActive:      v.User.IsActive,
			IsInBlacklist: v.User.IsInBlacklist,
			IsDeleted:     v.User.IsDeleted,
		}

		if !v.Profile.CreatedAt.Time.IsZero() {
			user.CreatedAt = v.Profile.CreatedAt.Time.String()
		}

		if !v.Profile.DeletedAt.Time.IsZero() {
			user.DeletedAt = v.Profile.DeletedAt.Time.String()
		}

		if !v.Profile.BlacklistAt.Time.IsZero() {
			user.BlacklistedAt = v.Profile.BlacklistAt.Time.String()
		}

		resp.Users = append(resp.Users, user)
	}

	return &resp, nil
}

func (u *serverImpl) GetUsersLoginID(ctx context.Context) (*map[string]string, error) {
	u.log.Info("get users list")
	users, err := u.svc.DB.GetAllUsersLoginID(ctx)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUsersList()
	}

	resp := make(map[string]string, 0)

	for _, v := range users {
		resp[v.ID] = v.Login
	}

	return &resp, nil
}

func (u *serverImpl) LinkResend(ctx context.Context, request umtypes.UserLogin) error {
	u.log.WithField("login", request.Login).Info("resending link")
	user, err := u.svc.DB.GetUserByLogin(ctx, request.Login)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableResendLink()
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		u.log.WithError(err)
		return err
	}
	if user.IsActive {
		return cherry.ErrUserAlreadyActivated()
	}

	link, err := u.svc.DB.GetLinkForUser(ctx, umtypes.LinkTypeConfirm, user)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableResendLink()
	}
	if link == nil {
		err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
			var err error
			link, err = tx.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, user)
			return err
		})
		if err := u.handleDBError(err); err != nil {
			u.log.WithError(err)
			return cherry.ErrUnableResendLink()
		}
	}
	if err := u.checkLinkResendTime(ctx, link); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableResendLink()
	}

	go u.linkSend(ctx, link)

	return nil
}
