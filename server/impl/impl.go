package impl

import (
	"io"
	"reflect"

	"git.containerum.net/ch/json-types/errors"

	"context"
	"time"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/grpc-proto-files/common"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	"git.containerum.net/ch/user-manager/models"
	"git.containerum.net/ch/user-manager/server"
	"github.com/sirupsen/logrus"
)

type serverImpl struct {
	svc server.Services
	log *logrus.Entry
}

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

func (u *serverImpl) checkLinkResendTime(ctx context.Context, link *models.Link) error {
	if tdiff := time.Now().UTC().Sub(link.SentAt.Time); link.SentAt.Valid && tdiff < 5*time.Minute {
		return &server.BadRequestError{Err: errors.Format(waitForResend, int(tdiff.Seconds()))}
	}
	return nil
}

func (u *serverImpl) linkSend(ctx context.Context, link *models.Link) {
	if link == nil {
		return
	}
	err := u.svc.DB.Transactional(ctx, func(ctx context.Context, tx models.DB) error {
		err := u.svc.MailClient.SendConfirmationMail(ctx, &mttypes.Recipient{
			ID:        link.User.ID,
			Name:      link.User.Login,
			Email:     link.User.Login,
			Variables: map[string]string{"CONFIRM": link.Link},
		})
		if err != nil {
			return err
		}
		link.SentAt.Time = time.Now().UTC()
		link.SentAt.Valid = true
		return tx.UpdateLink(ctx, link)
	})
	err = handleDBError(err)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id":    link.User.ID,
			"login": link.User.Login,
		}).Error("link send failed")
	}
}

func (u *serverImpl) createTokens(ctx context.Context, user *models.User) (resp *auth.CreateTokenResponse, err error) {
	// TODO: get access data from resource manager
	access := &auth.ResourcesAccess{}

	resp, err = u.svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   server.MustGetUserAgent(ctx),
		Fingerprint: server.MustGetFingerprint(ctx),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      server.MustGetClientIP(ctx),
		UserRole:    auth.Role(user.Role),
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
	u.log.WithError(err).Error("token create failed")
	if err != nil {
		err = tokenCreateFailed
	}
	return
}

func (u *serverImpl) loginUserChecks(ctx context.Context, user *models.User) error {
	if user == nil || user.IsDeleted {
		return &server.NotFoundError{Err: errors.Format(userNotFound)}
	}
	if user.IsInBlacklist {
		return &server.AccessDeniedError{Err: errors.Format(userBanned)}
	}
	return nil
}

func handleDBError(err error) error {
	switch err.(type) {
	case *errors.Error:
		return err
	}
	switch err {
	case nil:
		return nil
	case models.ErrTransactionRollback, models.ErrTransactionCommit, models.ErrTransactionBegin:
		return &server.InternalError{Err: errors.New("error on db transaction")}
	default:
		return &server.InternalError{Err: errors.New(err.Error())}
	}
}
