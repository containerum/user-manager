package models

import "time"

type DomainBlacklistEntry struct {
	Domain    string
	CreatedAt time.Time
}

func (db *DB) BlacklistDomain(domain string) error {
	db.log.Debug("Blacklisting domain", domain)
	_, err := db.eLog.Exec("INSERT INTO domains (domain) VALUES ('$1') ON CONFLICT DO NOTHING",
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
	rows, err := db.qLog.Queryx("SELECT COUNT(*) FROM domains WHERE domain = '$1'", domain)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	count := 0
	if rows.Next() {
		rows.Scan(&count)
	}
	return count > 0, err
}
