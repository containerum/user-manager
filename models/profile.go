package models

import "time"

type Profile struct {
	ID          string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"` // use UUID v4 as primary key (good support in psql)
	User        User
	Referral    string
	Access      string
	Data        string
	CreatedAt   time.Time
	BlacklistAt time.Time
	DeletedAt   time.Time
}

func (db *DB) CreateProfile(profile *Profile) error {
	db.log.Debug("Create profile for", profile.User.Login)
	return db.conn.Create(profile).Error
}

func (db *DB) GetProfileByID(id string) (*Profile, error) {
	db.log.Debug("Get profile by id", id)
	var profile Profile
	resp := db.conn.Where(&Profile{ID: id}).First(&profile)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &profile, resp.Error
}

func (db *DB) GetProfileByUser(user *User) (*Profile, error) {
	db.log.Debugf("Get profile by user %#v", user)
	var profile Profile
	resp := db.conn.Where(&Profile{User: *user}).First(&profile)
	if resp.RecordNotFound() {
		return nil, nil
	}
	return &profile, resp.Error
}

func (db *DB) UpdateProfile(profile *Profile) error {
	db.log.Debugf("Update profile %#v", profile)
	return db.conn.Save(profile).Error
}
