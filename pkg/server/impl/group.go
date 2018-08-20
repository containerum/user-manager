package impl

import (
	"context"

	"time"

	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	cherry "git.containerum.net/ch/user-manager/pkg/umerrors"
	kube_types "github.com/containerum/kube-client/pkg/model"
	"github.com/containerum/utils/httputil"
)

func (u *serverImpl) CreateGroup(ctx context.Context, request kube_types.UserGroup) (*string, error) {
	u.log.WithField("label", request.Label).Info("creating group")

	newGroup := &db.UserGroup{
		Label:   request.Label,
		OwnerID: httputil.MustGetUserID(ctx),
	}

	usr, err := u.svc.DB.GetUserByID(ctx, newGroup.OwnerID)
	if err != nil {
		return nil, err
	}

	newGroup.OwnerLogin = usr.Login

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.CreateGroup(ctx, newGroup)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableCreateGroup()
	}

	newGroupAdmin := &db.UserGroupMember{
		UserID:  httputil.MustGetUserID(ctx),
		GroupID: newGroup.ID,
		Access:  string(kube_types.OwnerAccess),
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
	u.log.WithField("groupID", groupID).Info("adding group members")

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

		if usr.Role == "admin" {
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
	u.log.WithField("groupID", groupID).Info("getting group")

	group, err := u.svc.DB.GetGroup(ctx, groupID)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetGroup()
	}

	if group == nil {
		return nil, cherry.ErrGroupNotExist()
	}

	ret := kube_types.UserGroup{
		ID:         group.ID,
		Label:      group.Label,
		OwnerID:    group.OwnerID,
		OwnerLogin: group.OwnerLogin,
		CreatedAt:  group.CreatedAt.Time.Format(time.RFC3339),
	}

	members, err := u.svc.DB.GetGroupMembers(ctx, group.ID)
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetGroup()
	}

	ret.UserGroupMembers = &kube_types.UserGroupMembers{Members: make([]kube_types.UserGroupMember, 0)}
	for _, member := range members {
		ret.Members = append(ret.Members, kube_types.UserGroupMember{
			Username: member.Login,
			ID:       member.UserID,
			Access:   kube_types.AccessLevel(member.Access),
		})
	}
	return &ret, nil
}

func (u *serverImpl) GetGroupsList(ctx context.Context, userID string) (*kube_types.UserGroups, error) {
	role := httputil.MustGetUserRole(ctx)
	u.log.WithField("userID", userID).Info("getting groups list")

	groupsIDs, err := u.svc.DB.GetUserGroupsIDsAccesses(ctx, userID, role == "admin")
	if err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetGroup()
	}

	groups := make([]kube_types.UserGroup, 0)
	for gr, perm := range groupsIDs {
		group, err := u.svc.DB.GetGroup(ctx, gr)
		if err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrUnableGetGroup()
		}
		if group == nil {
			return nil, cherry.ErrGroupNotExist()
		}

		membersCount, err := u.svc.DB.CountGroupMembers(ctx, gr)
		if err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrUnableGetGroup()
		}
		userGroup := kube_types.UserGroup{
			UserAccess:   kube_types.AccessLevel(perm),
			ID:           group.ID,
			Label:        group.Label,
			OwnerID:      group.OwnerID,
			OwnerLogin:   group.OwnerLogin,
			CreatedAt:    group.CreatedAt.Time.Format(time.RFC3339),
			MembersCount: *membersCount,
		}
		groups = append(groups, userGroup)
	}
	return &kube_types.UserGroups{Groups: groups}, nil
}

func (u *serverImpl) DeleteGroupMember(ctx context.Context, group kube_types.UserGroup, username string) error {
	u.log.WithField("groupID", group.ID).WithField("username", username).Info("deleting group member")

	usr, err := u.svc.DB.GetUserByLogin(ctx, username)
	if err != nil {
		u.log.WithError(err)
		return err
	}

	if usr == nil {
		return cherry.ErrUserNotExist().AddDetails(username)
	}

	if usr.ID == group.OwnerID {
		return cherry.ErrUnableRemoveOwner()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.DeleteGroupMember(ctx, usr.ID, group.ID)
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

func (u *serverImpl) UpdateGroupMemberAccess(ctx context.Context, group kube_types.UserGroup, username, access string) error {
	u.log.WithField("groupID", group.ID).WithField("username", username).WithField("access", access).Info("updating group member access")

	usr, err := u.svc.DB.GetUserByLogin(ctx, username)
	if err != nil {
		u.log.WithError(err)
		return err
	}

	if usr == nil {
		return cherry.ErrUserNotExist().AddDetails(username)
	}

	if usr.ID == group.OwnerID {
		return cherry.ErrUnableChangeOwnerPermissions()
	}

	err = u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UpdateGroupMember(ctx, usr.ID, group.ID, access)
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

func (u *serverImpl) DeleteGroup(ctx context.Context, groupID string) error {
	u.log.WithField("groupID", groupID).Info("deleting group")

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.DeleteGroup(ctx, groupID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableDeleteGroup().AddDetailsErr(err)
	}

	return nil
}

func (u *serverImpl) GetGroupListLabelID(ctx context.Context, ids []string) (*models.LoginID, error) {
	u.log.Info("get groups list")
	groups, err := u.svc.DB.GetGroupListLabelID(ctx, ids)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUsersList()
	}

	resp := make(models.LoginID)

	for _, v := range groups {
		resp[v.ID] = v.Label
	}

	return &resp, nil
}

func (u *serverImpl) GetGroupListByIDs(ctx context.Context, ids []string) (*kube_types.UserGroups, error) {
	u.log.Info("get groups list by ids")
	groups, err := u.svc.DB.GetGroupListByIDs(ctx, ids)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetUsersList()
	}

	resp := make([]kube_types.UserGroup, 0)

	for _, v := range groups {
		group := kube_types.UserGroup{
			ID:         v.ID,
			Label:      v.Label,
			OwnerID:    v.OwnerID,
			OwnerLogin: v.OwnerLogin,
			CreatedAt:  v.CreatedAt.Time.Format(time.RFC3339),
		}

		members, err := u.svc.DB.GetGroupMembers(ctx, v.ID)
		if err != nil {
			u.log.WithError(err)
			return nil, cherry.ErrUnableGetGroup()
		}

		group.UserGroupMembers = &kube_types.UserGroupMembers{Members: make([]kube_types.UserGroupMember, 0)}
		for _, member := range members {
			group.Members = append(group.Members, kube_types.UserGroupMember{
				Username: member.Login,
				ID:       member.UserID,
				Access:   kube_types.AccessLevel(member.Access),
			})
		}

		resp = append(resp, group)
	}

	return &kube_types.UserGroups{Groups: resp}, nil
}
