package clients

import (
	"context"

	"time"

	"fmt"

	"github.com/containerum/cherry"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// AuthClient is an extension of AuthClient interface with Closer interface to close connection
type TelegramClient interface {
	SendRegistrationMessage(ctx context.Context, userLogin string) error
	SendActivationMessage(ctx context.Context, userLogin string) error
}

type httpTelegramClient struct {
	log    *logrus.Entry
	client *resty.Client
	chatID string
	tgPath string
}

type TelegramResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		Chat      struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}

// NewTelegramClient returns client for auth service works using http protocol
func NewTelegramClient(botid, token, chatid string) (TelegramClient, error) {
	log := logrus.WithField("component", "telegram_client")

	client := resty.New().
		SetHostURL("https://api.telegram.org/").
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetTimeout(10 * time.Second).
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpTelegramClient{
		log:    log,
		client: client,
		chatID: chatid,
		tgPath: fmt.Sprintf("/bot%s:%s/sendMessage", botid, token),
	}, nil
}

func (c *httpTelegramClient) SendRegistrationMessage(ctx context.Context, userLogin string) error {
	c.log.Debugf("Sending registration message for %s", userLogin)

	msg := fmt.Sprintf("%s зарегистрирован на web.containerum.io %s", userLogin, time.Now().Format("02.01.2006 15:04:05 MST"))

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(TelegramResponse{}).
		SetQueryParam("chat_id", c.chatID).
		SetQueryParam("text", msg).
		Get(c.tgPath)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(error)
	}
	return nil
}

func (c *httpTelegramClient) SendActivationMessage(ctx context.Context, userLogin string) error {
	c.log.Debugf("Sending registration message for %s", userLogin)

	msg := fmt.Sprintf("%s подтвердил регистрацию на web.containerum.io %s", userLogin, time.Now().Format("02.01.2006 15:04:05 MST"))

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(TelegramResponse{}).
		SetQueryParam("chat_id", c.chatID).
		SetQueryParam("text", msg).
		Get(c.tgPath)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(error)
	}
	return nil
}
