package postgres

import (
	"context"
	"errors"

	"git.containerum.net/ch/user-manager/pkg/db"
)

func (pgdb *pgDB) BlacklistDomain(ctx context.Context, domain string, userID string) error {
	pgdb.log.Infoln("Blacklisting domain", domain)
	_, err := pgdb.eLog.ExecContext(ctx, "INSERT INTO domains (domain, added_by) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		domain, userID)
	return err
}

func (pgdb *pgDB) UnBlacklistDomain(ctx context.Context, domain string) error {
	pgdb.log.Infoln("UnBlacklisting domain", domain)
	res, err := pgdb.eLog.ExecContext(ctx, "DELETE FROM domains WHERE domain = $1", domain)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return errors.New("domain is not in blacklist")
	}
	return nil
}

func (pgdb *pgDB) IsDomainBlacklisted(ctx context.Context, domain string) (bool, error) {
	pgdb.log.Infof("Checking if domain %s in blacklist", domain)
	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT COUNT(*) FROM domains WHERE domain = $1", domain)
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

func (pgdb *pgDB) GetBlacklistedDomain(ctx context.Context, domain string) (*db.DomainBlacklistEntry, error) {
	pgdb.log.Infof("Getting info about domain %s", domain)

	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT domain, created_at, added_by FROM domains WHERE domain = $1", domain)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()

	ret := db.DomainBlacklistEntry{}
	err = rows.Scan(&ret.Domain, &ret.CreatedAt, &ret.AddedBy)
	return &ret, err
}

func (pgdb *pgDB) GetBlacklistedDomainsList(ctx context.Context) ([]db.DomainBlacklistEntry, error) {
	pgdb.log.Infof("Checking domains list")
	resp := make([]db.DomainBlacklistEntry, 0)

	rows, err := pgdb.qLog.QueryxContext(ctx, "SELECT domain, created_at, added_by FROM domains")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var domain db.DomainBlacklistEntry
		err := rows.StructScan(&domain)
		if err != nil {
			return nil, err
		}
		resp = append(resp, domain)
	}

	return resp, rows.Err()
}
