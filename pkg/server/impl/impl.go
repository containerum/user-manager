package impl

import (
	"io"
	"reflect"

	"errors"

	"git.containerum.net/ch/user-manager/pkg/db"
	"git.containerum.net/ch/user-manager/pkg/server"

	"context"
	"time"

	"fmt"

	"git.containerum.net/ch/auth/proto"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"github.com/sirupsen/logrus"
)

type serverImpl struct {
	svc server.Services
	log *logrus.Entry
}

// NewUserManagerImpl returns a main UserManager implementation
func NewUserManagerImpl(services server.Services) server.UserManager {
	return &serverImpl{
		svc: services,
		log: logrus.WithField("component", "user_manager_impl"),
	}
}

func (u *serverImpl) Close() error {
	var errs []error
	s := reflect.ValueOf(u.svc)
	closer := reflect.TypeOf((*io.Closer)(nil)).Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.Type().ConvertibleTo(closer) {
			errs = append(errs, f.Convert(closer).Interface().(io.Closer).Close())
		}
	}
	var strerr string
	for _, v := range errs {
		if v != nil {
			strerr += v.Error() + ";"
		}
	}
	return errors.New(strerr)
}

func (u *serverImpl) checkLinkResendTime(ctx context.Context, link *db.Link) error {
	if tdiff := time.Now().UTC().Sub(link.SentAt.Time); link.SentAt.Valid && tdiff < 5*time.Minute {
		return fmt.Errorf(waitForResend, int(tdiff.Seconds()))
	}
	return nil
}

func (u *serverImpl) linkSend(ctx context.Context, link *db.Link) {
	if link == nil {
		return
	}
	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx db.DB) error {
		err := u.svc.MailClient.SendConfirmationMail(ctx, &mttypes.Recipient{
			ID:        link.User.ID,
			Name:      link.User.Login,
			Email:     link.User.Login,
			Variables: map[string]interface{}{"CONFIRM": link.Link},
		})
		if err != nil {
			return err
		}
		link.SentAt.Time = time.Now().UTC()
		link.SentAt.Valid = true
		return tx.UpdateLink(ctx, link)
	})
	err = u.handleDBError(err)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id":    link.User.ID,
			"login": link.User.Login,
		}).Error("link send failed")
	}
}

func (u *serverImpl) createTokens(ctx context.Context, user *db.User) (resp *authProto.CreateTokenResponse, err error) {
	access, err := u.svc.ResourceServiceClient.GetUserAccess(ctx, user)
	if err != nil {
		u.log.WithError(err).Warning(resourceAccessGetFailed)
		return nil, errors.New(resourceAccessGetFailed)
	}

	resp, err = u.svc.AuthClient.CreateToken(ctx, &authProto.CreateTokenRequest{
		UserAgent:   server.MustGetUserAgent(ctx),
		Fingerprint: server.MustGetFingerprint(ctx),
		UserId:      user.ID,
		UserIp:      server.MustGetClientIP(ctx),
		UserRole:    user.Role,
		RwAccess:    true,
		Access:      access,
		PartTokenId: "00000000-0000-0000-0000-000000000000",
	})
	return
}

func (u *serverImpl) loginUserChecks(ctx context.Context, user *db.User) error {
	if user == nil {
		u.log.Error(cherry.ErrUserNotExist())
		return cherry.ErrUserNotExist()
	} else if user.IsDeleted {
		u.log.Error(cherry.ErrInvalidLogin())
		return cherry.ErrInvalidLogin()
	} else if user.IsInBlacklist {
		u.log.Error(cherry.ErrAccountBlocked())
		return cherry.ErrAccountBlocked()
	}
	return nil
}

func (u *serverImpl) checkReCaptcha(ctx context.Context, clientResponse string) error {
	remoteIP := server.MustGetClientIP(ctx)
	u.log.WithFields(logrus.Fields{
		"remote_ip":       remoteIP,
		"client_response": clientResponse,
	}).Info("checking recaptcha")
	resp, err := u.svc.ReCaptchaClient.Check(ctx, remoteIP, clientResponse)
	if err != nil {
		return cherry.ErrLoginFailed()
	}
	if !resp.Success {
		return cherry.ErrInvalidRecaptcha()
	}

	return nil
}

func (u *serverImpl) handleDBError(err error) error {
	switch err {
	case nil:
		return nil
	case db.ErrTransactionRollback, db.ErrTransactionCommit, db.ErrTransactionBegin:
		u.log.WithError(err).Error("db transaction error")
		return err
	default:
		u.log.WithError(err).Error("db error")
		return err
	}
}
