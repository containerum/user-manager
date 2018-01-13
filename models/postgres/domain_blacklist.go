package postgres

func (db *PgDB) BlacklistDomain(domain string) error {
	db.log.Infoln("Blacklisting domain", domain)
	_, err := db.eLog.Exec("INSERT INTO domains (domain) VALUES ($1) ON CONFLICT DO NOTHING",
		domain)
	return err
}

func (db *PgDB) UnBlacklistDomain(domain string) error {
	db.log.Infoln("UnBlacklisting domain", domain)
	_, err := db.eLog.Exec("DELETE FROM domains WHERE domain = $1", domain)
	return err
}

func (db *PgDB) IsDomainBlacklisted(domain string) (bool, error) {
	db.log.Infof("Checking if domain %s in blacklist", domain)
	rows, err := db.qLog.Queryx("SELECT COUNT(*) FROM domains WHERE domain = $1", domain)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	count := 0
	if !rows.Next() {
		return false, rows.Err()
	}
	err = rows.Scan(&count)
	return count > 0, err
}
