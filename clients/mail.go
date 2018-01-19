package clients

import (
	"context"

	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// MailClient is an interface to mail-templater service
type MailClient interface {
	SendConfirmationMail(ctx context.Context, recipient *mttypes.Recipient) error
	SendActivationMail(ctx context.Context, recipient *mttypes.Recipient) error
	SendBlockedMail(ctx context.Context, recipient *mttypes.Recipient) error
	SendPasswordChangedMail(ctx context.Context, recipient *mttypes.Recipient) error
	SendPasswordResetMail(ctx context.Context, recipient *mttypes.Recipient) error
	SendAccDeletedMail(ctx context.Context, recipient *mttypes.Recipient) error
}

type httpMailClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

// NewHTTPMailClient returns client for mail-templater service working via restful api
func NewHTTPMailClient(serverURL string) MailClient {
	log := logrus.WithField("component", "mail_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetError(errors.Error{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpMailClient{
		rest: client,
		log:  log,
	}
}

func (mc *httpMailClient) sendOneTemplate(ctx context.Context, tmplName string, recipient *mttypes.Recipient) error {
	req := &mttypes.SendRequest{}
	req.Delay = 0
	req.Message.Recipients = append(req.Message.Recipients, *recipient)
	resp, err := mc.rest.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(mttypes.SendResponse{}).
		Post("/templates/" + tmplName)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(*errors.Error)
	}
	return nil
}

func (mc *httpMailClient) SendConfirmationMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "confirm_reg", recipient)
}

func (mc *httpMailClient) SendActivationMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "activate_acc", recipient)
}

func (mc *httpMailClient) SendBlockedMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending blocked mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "blocked_acc", recipient)
}

func (mc *httpMailClient) SendPasswordChangedMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending password changed mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "pwd_changed", recipient)
}

func (mc *httpMailClient) SendPasswordResetMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending reset password mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "reset_pwd", recipient)
}

func (mc *httpMailClient) SendAccDeletedMail(ctx context.Context, recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending account deleted mail to", recipient.Email)
	return mc.sendOneTemplate(ctx, "delete_acc", recipient)
}
