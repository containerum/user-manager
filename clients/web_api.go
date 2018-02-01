package clients

import (
	"net/http"

	"context"

	"fmt"

	"crypto/sha256"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"

	"git.containerum.net/ch/grpc-proto-files/auth"
)

// WebAPIClient is an interface for web-api service from old architecture.
type WebAPIClient interface {
	Login(ctx context.Context, request *umtypes.WebAPILoginRequest) (ret *umtypes.WebAPILoginResponse, statusCode int, err error)
	GetVolumes(ctx context.Context, token string, userID string) (ret []*auth.AccessObject, statusCode int, err error)
	GetNamespaces(ctx context.Context, token string) (ret []*auth.AccessObject, statusCode int, err error)
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
		SetError(umtypes.WebAPIError{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpWebAPIClient{
		log:    log,
		client: client,
	}
}

// returns raw answer from web-api
func (c *httpWebAPIClient) Login(ctx context.Context, request *umtypes.WebAPILoginRequest) (ret *umtypes.WebAPILoginResponse, statusCode int, err error) {
	c.log.WithField("login", request.Username).Infoln("Signing in through web-api")

	resp, err := c.client.R().SetContext(ctx).SetBody(request).SetResult(&ret).Post("/api/login")
	if err != nil {
		c.log.WithError(err).Errorln("Sign in through web-api request failed")
		return nil, http.StatusInternalServerError, err
	}
	if resp.Error() != nil {
		err = resp.Error().(*umtypes.WebAPIError)
	}

	return ret, resp.StatusCode(), err
}

func (c *httpWebAPIClient) GetVolumes(ctx context.Context, token string, userID string) (ret []*auth.AccessObject, statusCode int, err error) {
	c.log.Infoln("Getting volumes")

	var volumes []umtypes.WebAPIVolumesResponse

	resp, err := c.client.R().SetContext(ctx).SetResult(&volumes).SetHeader("Authorization", token).Get("/api/volumes")
	if err != nil {
		c.log.WithError(err).Errorln("Unable to get volumes from WebAPI")
		return nil, http.StatusInternalServerError, err
	}
	if resp.Error() != nil {
		return nil, http.StatusInternalServerError, resp.Error().(*umtypes.WebAPIError)
	}

	for _, v := range volumes {
		ret = append(ret, &auth.AccessObject{Id: fmt.Sprintf("%x", (sha256.Sum256([]byte(userID + v.Name)))), Label: v.Name, Access: "owner"})
	}

	return ret, resp.StatusCode(), nil
}

func (c *httpWebAPIClient) GetNamespaces(ctx context.Context, token string) (ret []*auth.AccessObject, statusCode int, err error) {
	c.log.Infoln("Getting namespaces")

	var namespaces []umtypes.WebAPINamespaceResponse

	resp, err := c.client.R().SetContext(ctx).SetResult(&namespaces).SetHeader("Authorization", token).Get("/api/namespaces")
	if err != nil {
		c.log.WithError(err).Errorln("Unable to get namespaces from WebAPI")
		return nil, http.StatusInternalServerError, err
	}
	if resp.Error() != nil {
		return nil, http.StatusInternalServerError, resp.Error().(*umtypes.WebAPIError)
	}

	for _, v := range namespaces {
		ret = append(ret, &auth.AccessObject{Id: v.ID, Label: v.Name, Access: "owner"})
	}

	return ret, resp.StatusCode(), nil
}
