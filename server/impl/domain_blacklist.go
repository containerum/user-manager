package impl

import (
	"context"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
)

func (u *serverImpl) GetBlacklistedDomain(ctx context.Context, domain string) (*umtypes.DomainResponce, error) {
	u.log.WithField("domain", domain).Info("get domain info")
	blacklistedDomain, err := u.svc.DB.GetBlacklistedDomain(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, userGetFailed
	}

	if blacklistedDomain == nil {
		u.log.WithError(domainNotBlacklist)
		return nil, domainNotBlacklist
	}

	return &umtypes.DomainResponce{
		Domain:    blacklistedDomain.Domain,
		AddedBy:   blacklistedDomain.AddedBy.String,
		CreatedAt: blacklistedDomain.CreatedAt.String(),
	}, nil
}

func (u *serverImpl) GetBlacklistedDomainsList(ctx context.Context) (*umtypes.DomainListResponce, error) {
	u.log.Info("get domains list")
	blacklistedDomains, err := u.svc.DB.GetBlacklistedDomainsList(ctx)
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return nil, userGetFailed
	}

	if len(blacklistedDomains) == 0 {
		u.log.WithError(domainNotBlacklist)
		return nil, domainNotBlacklist
	}

	resp := umtypes.DomainListResponce{
		DomainList: []umtypes.DomainResponce{},
	}
	for _, v := range blacklistedDomains {
		resp.DomainList = append(resp.DomainList, umtypes.DomainResponce{
			Domain:    v.Domain,
			AddedBy:   v.AddedBy.String,
			CreatedAt: v.CreatedAt.String(),
		})
	}

	return &resp, nil
}

func (u *serverImpl) AddDomainToBlacklist(ctx context.Context, request umtypes.DomainToBlacklistRequest) error {
	u.log.WithField("domain", request.Domain).Info("adding domain to blacklist")

	userID := server.MustGetUserID(ctx)

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BlacklistDomain(ctx, request.Domain, userID)
	})
	if err := u.handleDBError(err); err != nil {
		u.log.WithError(err)
		return blacklistDomainFailed
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
		return unblacklistDomainFailed
	}

	return nil
}
