package postgres

import (
	"context"

	"git.containerum.net/ch/user-manager/pkg/db"
)

func (pgdb *pgDB) CreateGroup(ctx context.Context, group *db.UserGroup) error {
	pgdb.log.Infoln("Create group", group.Label)
	rows, err := pgdb.qLog.QueryxContext(ctx, "INSERT INTO groups (label, owner_user_id) "+
		"VALUES ($1, $2) RETURNING id",
		group.Label, group.OwnerID)
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
	pgdb.log.Infoln("Get group users", groupID)
	var group db.UserGroup // return empty slice instead of nil if no records found
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
	resp := make([]db.UserGroupMember, 0) // return empty slice instead of nil if no records found

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
