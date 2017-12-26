package clients

import (
	mttypes "git.containerum.net/ch/json-types/mail-templater"
	"git.containerum.net/ch/utils"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type MailClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

func NewMailClient(serverUrl string) *MailClient {
	log := logrus.WithField("component", "mail_client")
	client := resty.New().SetHostURL(serverUrl).SetLogger(log.WriterLevel(logrus.DebugLevel))
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &MailClient{
		rest: client,
		log:  log,
	}
}

func (mc *MailClient) sendOneTemplate(tmplName string, recipient *mttypes.Recipient) error {
	req := &mttypes.SendRequest{}
	req.Delay = 0
	req.Message.Recipients = append(req.Message.Recipients, *recipient)
	resp, err := mc.rest.R().
		SetBody(req).
		SetResult(mttypes.SendResponse{}).
		SetError(utils.Error{}).
		Post("/templates/" + tmplName)
	if err != nil {
		return err
	}
	return resp.Error().(*utils.Error)
}

func (mc *MailClient) SendConfirmationMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate("confirm_reg", recipient)
}

func (mc *MailClient) SendActivationMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending confirmation mail to", recipient.Email)
	return mc.sendOneTemplate("activate_acc", recipient)
}

func (mc *MailClient) SendBlockedMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending blocked mail to", recipient.Email)
	return mc.sendOneTemplate("blocked_acc", recipient)
}

func (mc *MailClient) SendPasswordChangedMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending password changed mail to", recipient.Email)
	return mc.sendOneTemplate("pwd_changed", recipient)
}

func (mc *MailClient) SendPasswordResetMail(recipient *mttypes.Recipient) error {
	mc.log.Infoln("Sending reset password mail to", recipient.Email)
	return mc.sendOneTemplate("reset_pwd", recipient)
}
