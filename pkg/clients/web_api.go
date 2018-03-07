package clients

import (
	"context"

	"fmt"

	"crypto/sha256"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"

	"net/http"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"git.containerum.net/ch/grpc-proto-files/auth"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
	"github.com/pkg/errors"
)

type WebAPIError struct {
	Error string `json:"message"`
}

// WebAPIClient is an interface for web-api service from old architecture.
type WebAPIClient interface {
	Login(ctx context.Context, request *umtypes.LoginRequest) (ret *umtypes.WebAPILoginResponse, err error)
	GetVolumes(ctx context.Context, token string, userID string) (ret []*auth.AccessObject, err error)
	GetNamespaces(ctx context.Context, token string) (ret []*auth.AccessObject, err error)
}

type httpWebAPIClient struct {
	log    *logrus.Entry
	client *resty.Client
}

// NewHTTPWebAPIClient returns client for web-api service working via restful api
func NewHTTPWebAPIClient(serverURL string) WebAPIClient {
	log := logrus.WithField("component", "web_api_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetHeader("Content-Type", "application/json").
		SetError(WebAPIError{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpWebAPIClient{
		log:    log,
		client: client,
	}
}

// returns raw answer from web-api
func (c *httpWebAPIClient) Login(ctx context.Context, request *umtypes.LoginRequest) (ret *umtypes.WebAPILoginResponse, err error) {
	c.log.WithField("login", request.Login).Infoln("Signing in through web-api")

	resp, err := c.client.R().SetContext(ctx).SetBody(kube_types.Login{Username: request.Login, Password: request.Password}).SetResult(&ret).Post("/api/login")
	if err != nil {
		c.log.WithError(err).Errorln("Sign in through web-api request failed")
		return nil, cherry.ErrLoginFailed()
	}
	if resp.Error() != nil {
		switch resp.StatusCode() {
		case http.StatusForbidden, http.StatusUnauthorized:
			return nil, cherry.ErrInvalidLogin()
		default:
			return nil, cherry.ErrLoginFailed()
		}
	}

	return ret, err
}

func (c *httpWebAPIClient) GetVolumes(ctx context.Context, token string, userID string) (ret []*auth.AccessObject, err error) {
	c.log.Infoln("Getting volumes")

	var volumes []umtypes.WebAPIResource

	resp, err := c.client.R().SetContext(ctx).SetResult(&volumes).SetHeader("Authorization", token).Get("/api/volumes")
	if err != nil {
		c.log.WithError(err).Errorln("Unable to get volumes from WebAPI")
		return nil, err
	}
	if resp.Error() != nil {
		return nil, errors.New("Unable to get volumes")
	}

	for _, v := range volumes {
		ret = append(ret, &auth.AccessObject{Id: fmt.Sprintf("%x", (sha256.Sum256([]byte(userID + v.Name)))), Label: v.Name, Access: "owner"})
	}

	return ret, nil
}

func (c *httpWebAPIClient) GetNamespaces(ctx context.Context, token string) (ret []*auth.AccessObject, err error) {
	c.log.Infoln("Getting namespaces")

	var namespaces []umtypes.WebAPIResource

	resp, err := c.client.R().SetContext(ctx).SetResult(&namespaces).SetHeader("Authorization", token).Get("/api/namespaces")
	if err != nil {
		c.log.WithError(err).Errorln("Unable to get namespaces from WebAPI")
		return nil, err
	}
	if resp.Error() != nil {
		return nil, errors.New("Unable to get namespaces")
	}

	for _, v := range namespaces {
		ret = append(ret, &auth.AccessObject{Id: v.ID, Label: v.Name, Access: "owner"})
	}

	return ret, nil
}
