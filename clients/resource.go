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

// ResourceServiceClient is an interface to resource-service.
type ResourceServiceClient interface {
	// GetUserAccess returns information about user access to resources (namespace, volumes) needed for token creation.
	GetUserAccess(ctx context.Context, user *models.User) (*auth.ResourcesAccess, error)
}

type httpResourceServiceClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

// NewHTTPResourceServiceClient returns client for resource-service working via restful api
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
	headersMap := make(map[string]string)
	headersMap["X-User-ID"] = user.ID
	headersMap["X-User-Role"] = user.Role

	resp, err := c.rest.R().SetContext(ctx).
		SetResult(auth.ResourcesAccess{}).
		SetHeaders(headersMap).
		Get("/access")
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error().(*errors.Error)
	}
	return resp.Result().(*auth.ResourcesAccess), nil
}
