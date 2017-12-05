package clients

import (
	"git.containerum.net/ch/mail-templater/upstreams"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type MailClient struct {
	rest *resty.Client
	log  *logrus.Logger
}

func NewMailClient(serverUrl string) *MailClient {
	log := logrus.WithField("component", "mail_client").Logger
	client := resty.New().SetHostURL(serverUrl).SetLogger(log.Writer())
	return &MailClient{
		rest: client,
		log:  log,
	}
}

func (mc *MailClient) SendConfirmationMail(recipient *upstreams.Recipient) error {
	mc.log.Info("Sending confirmation mail to", recipient.Email)
	req := &upstreams.SendRequest{}
	req.Delay = 0
	req.Message.Recipients = append(req.Message.Recipients, *recipient)
	_, err := mc.rest.R().
		SetBody(req).
		SetResult(&upstreams.SendResponse{}).
		Post("/templates/confirm_reg")
	if err != nil {
		return err
	}
	return nil
}
