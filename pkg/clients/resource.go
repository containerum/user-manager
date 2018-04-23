package clients

import (
	"context"
	"time"

	"git.containerum.net/ch/auth/proto"
	"git.containerum.net/ch/kube-client/pkg/cherry"
	"git.containerum.net/ch/user-manager/pkg/db"
	umtypes "git.containerum.net/ch/user-manager/pkg/models"
	utils "git.containerum.net/ch/utils/httputil"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// ResourceServiceClient is an interface to resource-service.
type ResourceServiceClient interface {
	// GetUserAccess returns information about user access to resources (namespace, volumes) needed for token creation.
	GetUserAccess(ctx context.Context, user *db.User) (*authProto.ResourcesAccess, error)
	DeleteUserNamespaces(ctx context.Context, user *db.User) error
	DeleteUserVolumes(ctx context.Context, user *db.User) error
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
		SetTimeout(3 * time.Second).
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpResourceServiceClient{
		rest: client,
		log:  log,
	}
}

func (c *httpResourceServiceClient) GetUserAccess(ctx context.Context, user *db.User) (*authProto.ResourcesAccess, error) {
	c.log.WithField("user_id", user.ID).Info("Getting user access from resource service")
	headersMap := utils.RequestHeadersMap(ctx)
	headersMap[umtypes.UserIDHeader] = user.ID
	headersMap[umtypes.UserRoleHeader] = user.Role
	resp, err := c.rest.R().SetContext(ctx).
		SetResult(authProto.ResourcesAccess{}).
		SetHeaders(headersMap). // forward request headers to other our service
		Get("/access")
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error().(*cherry.Err)
	}
	return resp.Result().(*authProto.ResourcesAccess), nil
}

func (c *httpResourceServiceClient) DeleteUserNamespaces(ctx context.Context, user *db.User) error {
	c.log.WithField("user_id", user.ID).Info("Deleting user namespaces")
	headersMap := utils.RequestHeadersMap(ctx)
	headersMap[umtypes.UserIDHeader] = user.ID
	headersMap[umtypes.UserRoleHeader] = user.Role
	resp, err := c.rest.R().SetContext(ctx).
		SetResult(authProto.ResourcesAccess{}).
		SetHeaders(headersMap). // forward request headers to other our service
		Delete("/namespaces")
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}

func (c *httpResourceServiceClient) DeleteUserVolumes(ctx context.Context, user *db.User) error {
	c.log.WithField("user_id", user.ID).Info("Deleting user volumes")
	headersMap := utils.RequestHeadersMap(ctx)
	headersMap[umtypes.UserIDHeader] = user.ID
	headersMap[umtypes.UserRoleHeader] = user.Role
	resp, err := c.rest.R().SetContext(ctx).
		SetResult(authProto.ResourcesAccess{}).
		SetHeaders(headersMap). // forward request headers to other our service
		Delete("/volumes")
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error().(*cherry.Err)
	}
	return nil
}
