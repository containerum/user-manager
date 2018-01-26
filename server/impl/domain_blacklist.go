package impl

import (
	"context"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"github.com/sirupsen/logrus"
)

func (u *serverImpl) GetBlacklistedDomain(ctx context.Context, domain string) (*umtypes.DomainResponce, error) {
	u.log.WithField("domain", domain).Info("get domain info")
	blacklistedDomain, err := u.svc.DB.GetBlacklistedDomain(ctx, domain)
	if err := u.handleDBError(err); err != nil {
		return nil, userGetFailed
	}

	if blacklistedDomain == nil {
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
		return nil, userGetFailed
	}

	if len(blacklistedDomains) == 0 {
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
	u.log.Info("adding domain to blacklist")
	u.log.WithFields(logrus.Fields{
		"domain": request.Domain,
	})

	userID := server.MustGetUserID(ctx)

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.BlacklistDomain(ctx, request.Domain, userID)
	})
	if err := u.handleDBError(err); err != nil {
		return blacklistDomainFailed
	}

	return nil
}

func (u *serverImpl) RemoveDomainFromBlacklist(ctx context.Context, domain string) error {
	u.log.Info("removing blacklisted domain")
	u.log.WithFields(logrus.Fields{
		"domain": domain,
	})

	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		return tx.UnBlacklistDomain(ctx, domain)
	})
	if err := u.handleDBError(err); err != nil {
		return unblacklistDomainFailed
	}

	return nil
}
