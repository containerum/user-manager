package clients

import (
	"context"

	"git.containerum.net/ch/grpc-proto-files/auth"
	"git.containerum.net/ch/json-types/errors"
	"git.containerum.net/ch/user-manager/models"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type ResourceServiceClient interface {
	GetUserAccess(ctx context.Context, user *models.User) (*auth.ResourcesAccess, error)
}

type httpResourceServiceClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

func NewHTTPResourceServiceClient(serverURL string) ResourceServiceClient {
	log := logrus.WithField("component", "resource_service_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetError(errors.Error{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpResourceServiceClient{
		rest: client,
		log:  log,
	}
}

func (c *httpResourceServiceClient) GetUserAccess(ctx context.Context, user *models.User) (*auth.ResourcesAccess, error) {
	c.log.WithField("user_id", user.ID).Info("Getting user access from resource service")
	resp, err := c.rest.R().SetContext(ctx).
		SetResult(auth.ResourcesAccess{}).
		SetHeader("x-user-id", user.ID).
		Get("/access")
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error().(*errors.Error)
	}
	return resp.Result().(*auth.ResourcesAccess), nil
}
