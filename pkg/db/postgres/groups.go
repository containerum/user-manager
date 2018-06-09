package postgres

import (
	"context"
	"errors"

	"git.containerum.net/ch/user-manager/pkg/db"
)

func (pgdb *pgDB) CreateGroup(ctx context.Context, group *db.UserGroup) error {
	pgdb.log.Infoln("Create group", group.Label)
	rows, err := pgdb.qLog.QueryxContext(ctx, "INSERT INTO groups (label, owner_login, owner_user_id) "+
		"VALUES ($1, $2, $3) RETURNING id",
		group.Label, group.OwnerLogin, group.OwnerID)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}
	return rows.Scan(&group.ID)
}

func (pgdb *pgDB) AddGroupMembers(ctx context.Context, member *db.UserGroupMember) error {
	pgdb.log.Infoln("Adding group member", member.UserID)
	rows, err := pgdb.qLog.QueryxContext(ctx, "INSERT INTO groups_members (group_id, user_id, default_access) "+
		"VALUES ($1, $2, $3) RETURNING id",
		member.GroupID, member.UserID, member.Access)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}
	return rows.Scan(&member.ID)
}

func (pgdb *pgDB) GetGroup(ctx context.Context, groupID string) (*db.UserGroup, error) {
	pgdb.log.Infoln("Get group", groupID)
	var group db.UserGroup
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT * FROM groups WHERE id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.StructScan(&group)
	return &group, err
}

func (pgdb *pgDB) GetGroupMembers(ctx context.Context, groupID string) ([]db.UserGroupMember, error) {
	pgdb.log.Infoln("Get group users", groupID)
	resp := make([]db.UserGroupMember, 0)

	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT * FROM groups_members WHERE group_id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var member db.UserGroupMember
		err := rows.StructScan(&member)
		if err != nil {
			return nil, err
		}
		resp = append(resp, member)
	}

	return resp, err
}

func (pgdb *pgDB) GetUserGroupsIDsAccesses(ctx context.Context, userID string) (map[string]string, error) {
	pgdb.log.Infoln("Get users groups", userID)
	resp := make(map[string]string, 0)

	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT group_id, default_access FROM groups_members WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var groupId string
		var access string
		err := rows.Scan(&groupId, &access)
		if err != nil {
			return nil, err
		}
		resp[groupId] = access
	}

	return resp, err
}

func (pgdb *pgDB) CountGroupMembers(ctx context.Context, groupID string) (*uint, error) {
	pgdb.log.Infoln("Count group members", groupID)

	var membersCount uint
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT count(id) FROM groups_members WHERE group_id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	err = rows.Scan(&membersCount)

	return &membersCount, err
}

func (pgdb *pgDB) DeleteGroupMember(ctx context.Context, userID string, groupID string) error {
	pgdb.log.Infoln("Delete member", userID)
	res, err := pgdb.eLog.ExecContext(ctx, "DELETE FROM groups_members WHERE group_id = $1 AND user_id = $2", groupID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return errors.New("user is not in this group")
	}
	return nil
}

func (pgdb *pgDB) DeleteGroupMemberFromAllGroups(ctx context.Context, userID string) error {
	pgdb.log.Infoln("Delete member", userID)
	_, err := pgdb.eLog.ExecContext(ctx, "DELETE FROM groups_members WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	return nil
}

func (pgdb *pgDB) UpdateGroupMember(ctx context.Context, userID string, groupID string, access string) error {
	pgdb.log.WithField("userID", userID).WithField("access", access).Infoln("Update member access")
	res, err := pgdb.eLog.ExecContext(ctx, "UPDATE groups_members SET default_access = $3 WHERE group_id = $1 AND user_id = $2", groupID, userID, access)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return errors.New("user is not in this group")
	}
	return nil
}

func (pgdb *pgDB) DeleteGroup(ctx context.Context, groupID string) error {
	pgdb.log.Infoln("Delete group", groupID)
	_, err := pgdb.eLog.ExecContext(ctx, "DELETE FROM groups WHERE id = $1", groupID)
	if err != nil {
		return err
	}
	return nil
}
