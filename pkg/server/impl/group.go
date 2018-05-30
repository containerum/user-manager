package impl

import (
	"context"

	"errors"

	"git.containerum.net/ch/user-manager/pkg/db"
	cherry "git.containerum.net/ch/user-manager/pkg/umErrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
)

func (u *serverImpl) CreateGroup(ctx context.Context, request kube_types.UserGroup) error {
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
		//TODO
		return cherry.ErrUnableBlacklistDomain()
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
		//TODO
		return cherry.ErrUnableBlacklistDomain()
	}

	_ = u.CreateGroupMembers(ctx, newGroup.ID, *request.UserGroupMembers)
	return nil
}

func (u *serverImpl) CreateGroupMembers(ctx context.Context, groupID string, request kube_types.UserGroupMembers) error {
	u.log.Info("adding group members")

	var created int

	for _, member := range request.Members {
		usr, err := u.svc.DB.GetUserByLogin(ctx, member.Username)
		if err != nil {
			u.log.WithError(err)
			continue
		}

		if usr == nil {
			u.log.WithError(errors.New("user not exists"))
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
			continue
		}
		created++
	}

	if created == 0 {
		return errors.New("no members were added to group")
	}

	return nil
}

func (u *serverImpl) GetGroup(ctx context.Context, groupID string) (*kube_types.UserGroup, error) {
	u.log.Info("adding group members")

	var group *db.UserGroup
	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		group, err = tx.GetGroup(ctx, groupID)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableBlacklistDomain()
	}
	ret := kube_types.UserGroup{
		ID:        group.ID,
		Label:     group.Label,
		OwnerID:   group.OwnerID,
		CreatedAt: group.CreatedAt.Time.String(),
	}

	var members []db.UserGroupMember
	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		var err error
		members, err = tx.GetGroupMembers(ctx, groupID)
		return err
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		//TODO
		return nil, cherry.ErrUnableBlacklistDomain()
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
