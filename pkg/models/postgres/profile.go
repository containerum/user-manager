package postgres

import (
	"git.containerum.net/ch/user-manager/pkg/models"
	"github.com/json-iterator/go"

	"context"

	"database/sql"
)

const profileQueryColumnsWithUserAndAccounts = "profiles.id, profiles.referral, profiles.access, profiles.created_at, profiles.blacklist_at, profiles.deleted_at, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist, profiles.data, accounts.github, accounts.google, accounts.facebook"
const profileQueryColumnsWithUser = "profiles.id, profiles.referral, profiles.access, profiles.created_at, profiles.blacklist_at, profiles.deleted_at, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist, profiles.data"

const profileQueryColumns = "id, referral, access, created_at, blacklist_at, deleted_at, data"

func (db *pgDB) CreateProfile(ctx context.Context, profile *models.Profile) error {
	db.log.Infoln("Create profile for", profile.User.Login)
	profileData, err := jsoniter.MarshalToString(profile.Data)
	if err != nil {
		return err
	}
	rows, err := db.qLog.QueryxContext(ctx, "INSERT INTO profiles (referral, access, user_id, data, created_at) VALUES "+
		"($1, $2, $3, $4, $5) RETURNING id, created_at", profile.Referral, profile.Access, profile.User.ID, profileData, profile.CreatedAt)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return rows.Err()
	}

	err = rows.Scan(&profile.ID, &profile.CreatedAt)
	return err
}

func (db *pgDB) GetProfileByID(ctx context.Context, id string) (*models.Profile, error) {
	db.log.Infoln("Get profile by id", id)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+profileQueryColumnsWithUser+" FROM profiles "+
		"JOIN users ON profiles.user_id = user.id "+
		"WHERE profiles.id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	profile := models.Profile{User: &models.User{}}
	var profileData string
	err = rows.Scan(
		&profile.ID, &profile.Referral, &profile.Access, &profile.CreatedAt, &profile.BlacklistAt, &profile.DeletedAt,
		&profile.User.ID, &profile.User.Login, &profile.User.PasswordHash, &profile.User.Salt, &profile.User.Role,
		&profile.User.IsActive, &profile.User.IsDeleted, &profile.User.IsInBlacklist,
		&profileData,
	)
	if err != nil {
		return nil, err
	}
	if err := jsoniter.UnmarshalFromString(profileData, &profile.Data); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (db *pgDB) GetProfileByUser(ctx context.Context, user *models.User) (*models.Profile, error) {
	db.log.Infof("Get profile by user %#v", user)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+profileQueryColumns+" FROM profiles "+
		"WHERE profiles.user_id = $1", user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, rows.Err()
	}
	profile := models.Profile{User: user}
	var profileData string

	err = rows.Scan(&profile.ID, &profile.Referral, &profile.Access, &profile.CreatedAt, &profile.BlacklistAt, &profile.DeletedAt, &profileData)
	if err != nil {
		return nil, err
	}
	if err := jsoniter.UnmarshalFromString(profileData, &profile.Data); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (db *pgDB) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	db.log.Infof("Update profile %#v", profile)
	profileData, err := jsoniter.MarshalToString(profile.Data)
	if err != nil {
		return err
	}
	_, err = db.eLog.ExecContext(ctx, "UPDATE profiles SET referral = $2, access = $3, data = $4 WHERE id = $1",
		profile.ID, profile.Referral, profile.Access, profileData)
	return err
}

func (db *pgDB) GetAllProfiles(ctx context.Context, perPage, offset int) ([]models.UserProfileAccounts, error) {
	db.log.Infoln("Get all profiles")
	profiles := make([]models.UserProfileAccounts, 0) // return empty slice instead of nil if no records found

	rows, err := db.qLog.QueryxContext(ctx, "SELECT "+profileQueryColumnsWithUserAndAccounts+" FROM users "+
		"LEFT JOIN profiles ON users.id = profiles.user_id "+
		"LEFT JOIN accounts ON users.id = accounts.user_id "+
		"LIMIT $1 OFFSET $2", perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		profile := models.UserProfileAccounts{User: &models.User{}, Accounts: &models.Accounts{}, Profile: &models.Profile{}}
		var profileData sql.NullString
		err := rows.Scan(
			&profile.Profile.ID, &profile.Profile.Referral, &profile.Profile.Access, &profile.Profile.CreatedAt, &profile.Profile.BlacklistAt, &profile.Profile.DeletedAt,
			&profile.User.ID, &profile.User.Login, &profile.User.PasswordHash, &profile.User.Salt, &profile.User.Role,
			&profile.User.IsActive, &profile.User.IsDeleted, &profile.User.IsInBlacklist,
			&profileData, &profile.Accounts.Github, &profile.Accounts.Google, &profile.Accounts.Facebook,
		)

		if err != nil {
			return nil, err
		}
		if profileData.Valid {
			if err := jsoniter.UnmarshalFromString(profileData.String, &profile.Profile.Data); err != nil {
				return nil, err
			}
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}
