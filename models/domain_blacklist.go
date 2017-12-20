package models

import "time"

type DomainBlacklistEntry struct {
	Domain    string
	CreatedAt time.Time
}

func (db *DB) BlacklistDomain(domain string) error {
	db.log.Debug("Blacklisting domain", domain)
	_, err := db.eLog.Exec("INSERT INTO domains (domain, created_at) VALUES ('$1', NOW()) ON CONFLICT DO NOTHING",
		domain)
	return err
}

func (db *DB) UnBlacklistDomain(domain string) error {
	db.log.Debug("UnBlacklisting domain", domain)
	_, err := db.eLog.Exec("DELETE FROM domains WHERE domain = '$1'", domain)
	return err
}

func (db *DB) IsInBlacklist(domain string) (bool, error) {
	db.log.Debugf("Checking if domain %s in blacklist", domain)
	rows, err := db.qLog.Queryx("SELECT * FROM domains WHERE domain = '$1'", domain)
	if err != nil {
		return false, err
	}
	// returns true if we have rows in result => domain blacklisted
	return rows.Next(), err
}
