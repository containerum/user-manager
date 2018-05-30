package impl

import (
	"context"

	"errors"

	"time"

	"git.containerum.net/ch/user-manager/pkg/db"
	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
)

func (u *serverImpl) CreateGroup(ctx context.Context, request kube_types.UserGroup) (*string, error) {
	u.log.WithField("label", request.Label).Info("creating group")

	newGroup := &db.UserGroup{
		Label:   request.Label,
		OwnerID: httputil.MustGetUserID(ctx),
	}

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.CreateGroup(ctx, newGroup)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateGroup()
	}

	newGroupAdmin := &db.UserGroupMember{
		UserID:  httputil.MustGetUserID(ctx),
		GroupID: newGroup.ID,
		Access:  string(kube_types.AdminAccess),
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.AddGroupMembers(ctx, newGroupAdmin)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateGroup()
	}

	if request.UserGroupMembers != nil {
		_ = u.AddGroupMembers(ctx, newGroup.ID, *request.UserGroupMembers)
	}
	return &newGroup.ID, nil
}

func (u *serverImpl) AddGroupMembers(ctx context.Context, groupID string, request kube_types.UserGroupMembers) error {
	u.log.Info("adding group members")

	var errs []error
	var created int
	for _, member := range request.Members {
		usr, err := u.svc.DB.GetUserByLogin(ctx, member.Username)
		if err != nil {
			u.log.WithError(err)
			errs = append(errs, err)
			continue
		}

		if usr == nil {
			u.log.WithError(cherry.ErrUserNotExist().AddDetails(member.Username))
			errs = append(errs, cherry.ErrUserNotExist().AddDetails(member.Username))
			continue
		}

		newGroupMember := &db.UserGroupMember{
			UserID:  usr.ID,
			GroupID: groupID,
			Access:  string(member.Access),
		}

		err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
			return tx.AddGroupMembers(ctx, newGroupMember)
		})
		if err := u.handleDBError(err); err != nil {
			u.log.WithError(err)
			errs = append(errs, err)
			continue
		}
		created++
	}

	if created == 0 {
		return cherry.ErrUnableAddGroupMember().AddDetailsErr(errs...)
	}

	return nil
}

func (u *serverImpl) GetGroup(ctx context.Context, groupID string) (*kube_types.UserGroup, error) {
	u.log.Info("adding group members")

	group, err := u.svc.DB.GetGroup(ctx, groupID)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetGroup()
	}

	if group == nil {
		return nil, cherry.ErrGroupNotExist()
	}

	ret := kube_types.UserGroup{
		ID:        group.ID,
		Label:     group.Label,
		OwnerID:   group.OwnerID,
		CreatedAt: group.CreatedAt.Time.Format(time.RFC3339),
	}

	var members []db.UserGroupMember
	members, err = u.svc.DB.GetGroupMembers(ctx, groupID)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetGroup()
	}

	ret.UserGroupMembers = &kube_types.UserGroupMembers{Members: make([]kube_types.UserGroupMember, 0)}
	for _, member := range members {
		usr, err := u.svc.DB.GetUserByID(ctx, member.UserID)
		if err != nil {
			u.log.WithError(err)
			continue
		}

		if usr == nil {
			u.log.WithError(errors.New("user not found"))
			continue
		}

		newMember := kube_types.UserGroupMember{
			Username: usr.Login,
			Access:   kube_types.UserGroupAccess(member.Access),
		}
		ret.Members = append(ret.Members, newMember)
	}

	return &ret, nil
}

func (u *serverImpl) DeleteGroupMember(ctx context.Context, groupID string, username string) error {
	u.log.Info("deleting group members")

	usr, err := u.svc.DB.GetUserByLogin(ctx, username)
	if err != nil {
		u.log.WithError(err)
		return err
	}

	if usr == nil {
		return cherry.ErrUserNotExist().AddDetails(username)
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.DeleteGroupMember(ctx, usr.ID, groupID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		if err.Error() == "user is not in this group" {
			return cherry.ErrNotInGroup().AddDetails(username)
		}
		return cherry.ErrUnableDeleteGroupMember().AddDetailsErr(err)
	}

	return nil
}
