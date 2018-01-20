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
	"google.golang.org/grpc/status"
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

func (u *serverImpl) createTokens(ctx context.Context, user *models.User) (resp *auth.CreateTokenResponse, err error) {
	access, err := u.svc.ResourceServiceClient.GetUserAccess(ctx, user)
	if err != nil {
		u.log.WithError(err).Error("resource access get failed")
		err = resourceAccessGetFailed
		return
	}

	resp, err = u.svc.AuthClient.CreateToken(ctx, &auth.CreateTokenRequest{
		UserAgent:   server.MustGetUserAgent(ctx),
		Fingerprint: server.MustGetFingerprint(ctx),
		UserId:      &common.UUID{Value: user.ID},
		UserIp:      server.MustGetClientIP(ctx),
		UserRole:    user.Role,
		RwAccess:    true,
		Access:      access,
		PartTokenId: nil,
	})
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

func (u *serverImpl) checkReCaptcha(ctx context.Context, clientResponse string) error {
	remoteIP := server.MustGetClientIP(ctx)
	u.log.WithFields(logrus.Fields{
		"remote_ip":       remoteIP,
		"client_response": clientResponse,
	}).Info("checking recaptcha")
	resp, err := u.svc.ReCaptchaClient.Check(ctx, remoteIP, clientResponse)
	if err != nil {
		return reCaptchaRequestFailed
	}
	if !resp.Success {
		return &server.InternalError{Err: errors.New(invalidReCaptcha)}
	}

	return nil
}

func (u *serverImpl) handleDBError(err error) error {
	if _, ok := status.FromError(err); ok { // forward grpc errors up
		return err
	}
	switch err.(type) {
	case *errors.Error:
		return err
	case *server.InternalError, *server.WebAPIError, *server.AlreadyExistsError,
		*server.BadRequestError, *server.NotFoundError, *server.AccessDeniedError:
		return err
	}
	switch err {
	case nil:
		return nil
	case models.ErrTransactionRollback, models.ErrTransactionCommit, models.ErrTransactionBegin:
		u.log.WithError(err).Error("db transaction error")
		return err
	default:
		u.log.WithError(err).Error("db error")
		return err
	}
}
