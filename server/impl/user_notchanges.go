package impl

import (
	"context"

	"time"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
)

func (u *serverImpl) GetUserLinks(ctx context.Context, userID string) (*umtypes.LinksGetResponse, error) {
	u.log.WithField("user_id", userID).Info("get user links")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if user == nil {
		return nil, &server.NotFoundError{Err: errors.New(userNotFound)}
	}

	links, err := u.svc.DB.GetUserLinks(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, linkGetFailed
	}

	resp := umtypes.LinksGetResponse{Links: []umtypes.Link{}}
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

func (u *serverImpl) GetUserInfo(ctx context.Context) (*umtypes.UserInfoGetResponse, error) {
	userID := server.MustGetUserID(ctx)
	u.log.WithField("user_id", userID).Info("get user info")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return nil, err
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, profileGetFailed
	}

	return &umtypes.UserInfoGetResponse{
		Login:     user.Login,
		Data:      profile.Data,
		ID:        user.ID,
		IsActive:  user.IsActive,
		CreatedAt: profile.CreatedAt,
	}, nil
}

func (u *serverImpl) GetUserInfoByID(ctx context.Context, userID string) (*umtypes.UserInfoByIDGetResponse, error) {
	u.log.WithField("user_id", userID).Info("get user info by id")
	user, err := u.svc.DB.GetUserByID(ctx, userID)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}
	if user == nil {
		return nil, &server.NotFoundError{Err: errors.New(userNotFound)}
	}

	profile, err := u.svc.DB.GetProfileByUser(ctx, user)
	if err := u.handleDBError(err); err != nil {
		return nil, profileGetFailed
	}

	return &umtypes.UserInfoByIDGetResponse{
		Login: user.Login,
		Data:  profile.Data,
	}, nil
}

func (u *serverImpl) GetBlacklistedUsers(ctx context.Context, params umtypes.UserListQuery) (*umtypes.BlacklistGetResponse, error) {
	u.log.WithField("per_page", params.PerPage).WithField("page", params.Page).Info("get blacklisted users")
	blacklisted, err := u.svc.DB.GetBlacklistedUsers(ctx, params.PerPage, (params.Page-1)*params.PerPage)
	if err := u.handleDBError(err); err != nil {
		return nil, blacklistUsersGetFailed
	}
	var resp umtypes.BlacklistGetResponse
	for _, v := range blacklisted {
		resp.BlacklistedUsers = append(resp.BlacklistedUsers, umtypes.BlacklistedUserEntry{
			Login: v.Login,
			ID:    v.ID,
		})
	}
	return &resp, nil
}

func (u *serverImpl) GetUsers(ctx context.Context, params umtypes.UserListQuery, filters ...string) (*umtypes.UserListGetResponse, error) {
	u.log.WithField("per_page", params.PerPage).WithField("page", params.Page).Info("get users")
	profiles, err := u.svc.DB.GetAllProfiles(ctx, params.PerPage, (params.Page-1)*params.PerPage)
	if err := u.handleDBError(err); err != nil {
		return nil, profileGetFailed
	}
	if profiles == nil {
		return nil, &server.NotFoundError{Err: errors.New(profilesNotFound)}
	}

	satisfiesFilter := server.CreateFilterFunc(filters...)

	resp := umtypes.UserListGetResponse{
		Users: []umtypes.UserListEntry{},
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

		resp.Users = append(resp.Users, umtypes.UserListEntry{
			ID:            v.User.ID,
			Login:         v.User.Login,
			Referral:      v.Referral,
			Role:          v.User.Role,
			Access:        v.Access,
			CreatedAt:     v.CreatedAt,
			DeletedAt:     v.DeletedAt.Time,
			BlacklistedAt: v.BlacklistAt.Time,
			Data:          v.Data,
			IsActive:      v.User.IsActive,
			IsInBlacklist: v.User.IsInBlacklist,
			IsDeleted:     v.User.IsDeleted,
			Accounts:      accs,
		})
	}

	return &resp, nil
}

func (u *serverImpl) LinkResend(ctx context.Context, request umtypes.ResendLinkRequest) error {
	u.log.WithField("login", request.UserName).Info("resend link")
	user, err := u.svc.DB.GetUserByLogin(ctx, request.UserName)
	if err := u.handleDBError(err); err != nil {
		return userGetFailed
	}
	if err := u.loginUserChecks(ctx, user); err != nil {
		return err
	}

	link, err := u.svc.DB.GetLinkForUser(ctx, umtypes.LinkTypeConfirm, user)
	if err := u.handleDBError(err); err != nil {
		return linkGetFailed
	}
	if link == nil {
		err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
			var err error
			link, err = tx.CreateLink(ctx, umtypes.LinkTypeConfirm, 24*time.Hour, user)
			return err
		})
		if err := u.handleDBError(err); err != nil {
			return linkCreateFailed
		}
	}
	if err := u.checkLinkResendTime(ctx, link); err != nil {
		return err
	}

	go u.linkSend(ctx, link)

	return nil
}
