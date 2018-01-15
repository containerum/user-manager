package postgres

import "context"

func (db *pgDB) BlacklistDomain(ctx context.Context, domain string) error {
	db.log.Infoln("Blacklisting domain", domain)
	_, err := db.eLog.ExecContext(ctx, "INSERT INTO domains (domain) VALUES ($1) ON CONFLICT DO NOTHING",
		domain)
	return err
}

func (db *pgDB) UnBlacklistDomain(ctx context.Context, domain string) error {
	db.log.Infoln("UnBlacklisting domain", domain)
	_, err := db.eLog.ExecContext(ctx, "DELETE FROM domains WHERE domain = $1", domain)
	return err
}

func (db *pgDB) IsDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	db.log.Infof("Checking if domain %s in blacklist", domain)
	rows, err := db.qLog.QueryxContext(ctx, "SELECT COUNT(*) FROM domains WHERE domain = $1", domain)
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
