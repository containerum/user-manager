package clients

import (
	"git.containerum.net/ch/json-types/errors"
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type MailClient interface {
	SendConfirmationMail(recipient *mttypes.Recipient) error
	SendActivationMail(recipient *mttypes.Recipient) error
	SendBlockedMail(recipient *mttypes.Recipient) error
	SendPasswordChangedMail(recipient *mttypes.Recipient) error
	SendPasswordResetMail(recipient *mttypes.Recipient) error
	SendAccDeletedMail(recipient *mttypes.Recipient) error
}

type httpMailClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

func NewHTTPMailClient(serverUrl string) MailClient {
	log := logrus.WithField("component", "mail_client")
	client := resty.New().
		SetHostURL(serverUrl).
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

func (mc *httpMailClient) sendOneTemplate(tmplName string, recipient *mttypes.Recipient) error {
	req := &mttypes.SendRequest{}
	req.Delay = 0
	req.Message.Recipients = append(req.Message.Recipients, *recipient)
	resp, err := mc.rest.R().
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

func (mc *httpMailClient) SendConfirmationMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate("confirm_reg", recipient)
}

func (mc *httpMailClient) SendActivationMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate("activate_acc", recipient)
}

func (mc *httpMailClient) SendBlockedMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending blocked mail to", recipient.Email)
	return mc.sendOneTemplate("blocked_acc", recipient)
}

func (mc *httpMailClient) SendPasswordChangedMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending password changed mail to", recipient.Email)
	return mc.sendOneTemplate("pwd_changed", recipient)
}

func (mc *httpMailClient) SendPasswordResetMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending reset password mail to", recipient.Email)
	return mc.sendOneTemplate("reset_pwd", recipient)
}

func (mc *httpMailClient) SendAccDeletedMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending account deleted mail to", recipient.Email)
	return mc.sendOneTemplate("delete_acc", recipient)
}
