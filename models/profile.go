package models

import (
	"time"

	"github.com/json-iterator/go"
)

type ProfileData struct {
	Email          string `json:"email,omitempty" binding:"email"`
	Address        string `json:"address,omitempty"`
	Phone          string `json:"phone,omitempty"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	IsOrganization bool   `json:"is_organization,omitempty"`
	TaxCode        string `json:"tax_code,omitempty"`
	Company        string `json:"company,omitempty"`
}

type Profile struct {
	ID          string
	Referral    string
	Access      string
	CreatedAt   time.Time
	BlacklistAt time.Time
	DeletedAt   time.Time

	User *User

	Data ProfileData
}

const profileQueryColumns = "(profiles.id, profiles.referral, profiles.access, profiles.created_at, profiles.blacklisted_at, profiles.deleted_at, " +
	"users.id, users.login, users.password_hash, users.salt, users.role, users.is_active, users.is_deleted, users.is_in_blacklist, profiles.data)"

func (db *DB) CreateProfile(profile *Profile) error {
	db.log.Debug("Create profile for", profile.User.Login)
	profileData, err := jsoniter.MarshalToString(profile.Data)
	if err != nil {
		return err
	}
	rows, err := db.qLog.Queryx("INSERT INTO profiles (referral, access, created_at, user_id, data) VALUES "+
		"('$1', '$2', NOW(), '$3', '$4') RETURNING id, created_at", profile.Referral, profile.Access, profile.User.ID, profileData)
	if err != nil {
		return err
	}
	if rows.Next() {
		rows.Scan(&profile.ID)
	}
	if rows.Next() {
		rows.Scan(&profile.CreatedAt)
	}

	return rows.Err()
}

func (db *DB) GetProfileByID(id string) (*Profile, error) {
	db.log.Debug("Get profile by id", id)
	var profile Profile
	rows, err := db.qLog.Queryx("SELECT "+profileQueryColumns+" FROM profiles "+
		"JOIN users ON profiles.user_id = user.id WHERE profiles.id = '$1'", id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	rows.Scan(&profile.ID, &profile.Referral, &profile.Access, &profile.CreatedAt, &profile.BlacklistAt, &profile.DeletedAt)
	rows.StructScan(profile.User)

	var profileData string
	rows.Scan(&profileData)
	if err := jsoniter.UnmarshalFromString(profileData, &profile.Data); err != nil {
		return nil, err
	}

	return &profile, rows.Err()
}

func (db *DB) GetProfileByUser(user *User) (*Profile, error) {
	db.log.Debugf("Get profile by user %#v", user)
	var profile Profile
	rows, err := db.qLog.Queryx("SELECT "+profileQueryColumns+" FROM profiles "+
		"JOIN users ON profiles.user_id = user.id WHERE profiles.user_id = '$1'", user.ID)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, nil
	}
	rows.Scan(&profile.ID, &profile.Referral, &profile.Access, &profile.CreatedAt, &profile.BlacklistAt, &profile.DeletedAt)
	rows.StructScan(profile.User)

	var profileData string
	rows.Scan(&profileData)
	if err := jsoniter.UnmarshalFromString(profileData, &profile.Data); err != nil {
		return nil, err
	}

	return &profile, rows.Err()
}

func (db *DB) UpdateProfile(profile *Profile) error {
	db.log.Debugf("Update profile %#v", profile)
	_, err := db.eLog.Exec("UPDATE profiles SET referal = '$2', access = '$3', data = '$4 WHERE id = '$1'",
		profile.ID, profile.Referral, profile.Access, profile.Data)
	return err
}

func (db *DB) GetAllProfiles() ([]Profile, error) {
	db.log.Debug("Get all profiles")
	var profiles []Profile
	rows, err := db.qLog.Queryx("SELECT " + profileQueryColumns + " FROM profiles JOIN users ON profiles.user_id = user.id")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var profile Profile
		rows.Scan(&profile.ID, &profile.Referral, &profile.Access, &profile.CreatedAt, &profile.BlacklistAt, &profile.DeletedAt)
		rows.StructScan(profile.User)

		var profileData string
		rows.Scan(&profileData)
		if err := jsoniter.UnmarshalFromString(profileData, &profile.Data); err != nil {
			return profiles, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}
