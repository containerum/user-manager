package models

import "time"

type DomainBlacklistEntry struct {
	Domain    string `gorm:"primary_key"`
	CreatedAt time.Time
}

func (db *DB) BlacklistDomain(domain string) error {
	db.log.Debug("Blacklisting domain", domain)
	return db.conn.Create(&DomainBlacklistEntry{Domain: domain, CreatedAt: time.Now().UTC()}).Error
}

func (db *DB) UnBlacklistDomain(domain string) error {
	db.log.Debug("UnBlacklisting domain", domain)
	return db.conn.Delete(&DomainBlacklistEntry{Domain: domain}).Error
}

func (db *DB) IsInBlacklist(domain string) (bool, error) {
	db.log.Debugf("Checking if domain %s in blacklist", domain)
	resp := db.conn.Where(&DomainBlacklistEntry{Domain: domain}).First(&Profile{})
	if resp.RecordNotFound() {
		return false, nil
	} else if resp.Error != nil {
		return false, resp.Error
	}
	return true, nil
}
