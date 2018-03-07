package postgres

import (
	"context"
	"errors"

	"git.containerum.net/ch/user-manager/pkg/models"
)

func (db *pgDB) BlacklistDomain(ctx context.Context, domain string, userID string) error {
	db.log.Infoln("Blacklisting domain", domain)
	_, err := db.eLog.ExecContext(ctx, "INSERT INTO domains (domain, added_by) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		domain, userID)
	return err
}

func (db *pgDB) UnBlacklistDomain(ctx context.Context, domain string) error {
	db.log.Infoln("UnBlacklisting domain", domain)
	res, err := db.eLog.ExecContext(ctx, "DELETE FROM domains WHERE domain = $1", domain)
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

func (db *pgDB) GetBlacklistedDomain(ctx context.Context, domain string) (*models.DomainBlacklistEntry, error) {
	db.log.Infof("Getting info about domain %s", domain)

	rows, err := db.qLog.QueryxContext(ctx, "SELECT domain, created_at, added_by FROM domains WHERE domain = $1", domain)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, rows.Err()
	}
	defer rows.Close()

	ret := models.DomainBlacklistEntry{}
	err = rows.Scan(&ret.Domain, &ret.CreatedAt, &ret.AddedBy)
	return &ret, err
}

func (db *pgDB) GetBlacklistedDomainsList(ctx context.Context) ([]models.DomainBlacklistEntry, error) {
	db.log.Infof("Checking domains list")
	resp := make([]models.DomainBlacklistEntry, 0)

	rows, err := db.qLog.QueryxContext(ctx, "SELECT domain, created_at, added_by FROM domains")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var domain models.DomainBlacklistEntry
		err := rows.StructScan(&domain)
		if err != nil {
			return nil, err
		}
		resp = append(resp, domain)
	}

	return resp, rows.Err()
}
