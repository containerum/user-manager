package impl

import (
	"context"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"git.containerum.net/ch/user-manager/pkg/models"
	"git.containerum.net/ch/user-manager/pkg/server"
	"github.com/pkg/errors"
)

func (u *serverImpl) GetBlacklistedDomain(ctx context.Context, domain string) (*umtypes.Domain, error) {
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

	return &umtypes.Domain{
		Domain:    blacklistedDomain.Domain,
		AddedBy:   blacklistedDomain.AddedBy.String,
		CreatedAt: blacklistedDomain.CreatedAt,
	}, nil
}

func (u *serverImpl) GetBlacklistedDomainsList(ctx context.Context) (*umtypes.DomainListResponse, error) {
	u.log.Info("get domains list")
	blacklistedDomains, err := u.svc.DB.GetBlacklistedDomainsList(ctx)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, cherry.ErrUnableGetDomainBlacklist()
	}

	resp := umtypes.DomainListResponse{
		DomainList: []umtypes.Domain{},
	}
	for _, v := range blacklistedDomains {
		resp.DomainList = append(resp.DomainList, umtypes.Domain{
			Domain:    v.Domain,
			AddedBy:   v.AddedBy.String,
			CreatedAt: v.CreatedAt,
		})
	}

	return &resp, nil
}

func (u *serverImpl) AddDomainToBlacklist(ctx context.Context, request umtypes.Domain) error {
	u.log.WithField("domain", request.Domain).Info("adding domain to blacklist")

	userID := server.MustGetUserID(ctx)

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
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

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.UnBlacklistDomain(ctx, domain)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return cherry.ErrUnableUnblacklistDomain()
	}

	return nil
}
