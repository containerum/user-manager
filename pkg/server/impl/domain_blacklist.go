package impl

import (
	"context"

	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/models"
	cherry "git.containerum.net/ch/user-manager/pkg/umerrors"
	"github.com/containerum/utils/httputil"
	"github.com/pkg/errors"
)

func (u *serverImpl) GetBlacklistedDomain(ctx context.Context, domain string) (*models.Domain, error) {
	u.log.WithField("domain", domain).Info("get domain info")
	blacklistedDomain, err := u.svc.DB.GetBlacklistedDomain(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetDomainBlacklist()
	}

	if blacklistedDomain == nil {
		u.log.WithError(errors.New("dom not blacklist"))
		return nil, cherry.ErrDomainNotBlacklisted()
	}

	return &models.Domain{
		Domain:    blacklistedDomain.Domain,
		AddedBy:   blacklistedDomain.AddedBy.String,
		CreatedAt: blacklistedDomain.CreatedAt,
	}, nil
}

func (u *serverImpl) GetBlacklistedDomainsList(ctx context.Context) (*models.DomainListResponse, error) {
	u.log.Info("get domains list")
	blacklistedDomains, err := u.svc.DB.GetBlacklistedDomainsList(ctx)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetDomainBlacklist()
	}

	resp := models.DomainListResponse{
		DomainList: []models.Domain{},
	}
	for _, v := range blacklistedDomains {
		resp.DomainList = append(resp.DomainList, models.Domain{
			Domain:    v.Domain,
			AddedBy:   v.AddedBy.String,
			CreatedAt: v.CreatedAt,
		})
	}

	return &resp, nil
}

func (u *serverImpl) AddDomainToBlacklist(ctx context.Context, request models.Domain) error {
	u.log.WithField("domain", request.Domain).Info("adding domain to blacklist")

	userID := httputil.MustGetUserID(ctx)

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.BlacklistDomain(ctx, request.Domain, userID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableBlacklistDomain()
	}

	return nil
}

func (u *serverImpl) RemoveDomainFromBlacklist(ctx context.Context, domain string) error {
	u.log.WithField("domain", domain).Info("removing domain from blacklist")

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		return tx.UnBlacklistDomain(ctx, domain)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUnblacklistDomain()
	}

	return nil
}
