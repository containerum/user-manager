package clients

import (
	"time"

	"net/url"

	"context"

	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

const reCaptchaAPI = "https://www.google.com/recaptcha/api"

// ReCaptchaClient is an interface to Google`s ReCaptcha service
type ReCaptchaClient interface {
	Check(ctx context.Context, remoteIP, clientResponse string) (r *ReCaptchaResponse, err error)
}

type httpReCaptchaClient struct {
	client     *resty.Client
	log        *logrus.Entry
	privateKey string
}

// ReCaptchaResponse describes response from ReCaptcha service.
// It`s enough to check only "Success" field. Other fields is for logging purposes.
type ReCaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []int     `json:"error-codes"`
}

// NewHTTPReCaptchaClient returns client for ReCaptcha service working via HTTP
func NewHTTPReCaptchaClient(privateKey string) ReCaptchaClient {
	log := logrus.WithField("component", "recaptcha")
	client := resty.New().SetLogger(log.WriterLevel(logrus.DebugLevel)).SetHostURL(reCaptchaAPI).SetDebug(true)
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpReCaptchaClient{
		log:        log,
		client:     client,
		privateKey: privateKey,
	}
}

func (c *httpReCaptchaClient) Check(ctx context.Context, remoteIP, clientResponse string) (r *ReCaptchaResponse, err error) {
	c.log.Infoln("Checking ReCaptcha from", remoteIP)
	r = new(ReCaptchaResponse)
	_, err = c.client.R().SetContext(ctx).SetResult(r).SetMultiValueFormData(url.Values{
		"secret":   {c.privateKey},
		"remoteip": {remoteIP},
		"response": {clientResponse},
	}).Post("/siteverify")
	return
}

type dummyReCaptchaClient struct {
	log *logrus.Entry
}

// NewDummyReCaptchaClient returns a dummy client.
// It returns success on any request and log actions. Useful for testing purposes.
func NewDummyReCaptchaClient() ReCaptchaClient {
	return &dummyReCaptchaClient{
		log: logrus.WithField("component", "dummy_recaptcha_client"),
	}
}

func (c *dummyReCaptchaClient) Check(ctx context.Context, remoteIP, clientResponse string) (r *ReCaptchaResponse, err error) {
	c.log.Infoln("Checking ReCaptcha from", remoteIP)
	return &ReCaptchaResponse{
		Success:     true,
		ChallengeTS: time.Now(),
		Hostname:    "dummy",
	}, nil
}
