package clients

import (
	"context"
	"time"

	"git.containerum.net/ch/user-manager/pkg/db"
	"github.com/containerum/cherry"
	headers "github.com/containerum/utils/httputil"
	utils "github.com/containerum/utils/httputil"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// PermissionsClient is an interface to permissions service.
type PermissionsClient interface {
	DeleteUserNamespaces(ctx context.Context, user *db.User) error
}

type httpPermissionsClient struct {
	rest *resty.Client
	log  *logrus.Entry
}

// NewHTTPPermissionsClient returns client for permissions service working via restful api
func NewHTTPPermissionsClient(serverURL string) PermissionsClient {
	log := logrus.WithField("component", "permissions_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetTimeout(3 * time.Second).
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpPermissionsClient{
		rest: client,
		log:  log,
	}
}

func (c *httpPermissionsClient) DeleteUserNamespaces(ctx context.Context, user *db.User) error {
	c.log.WithField("user_id", user.ID).Info("Deleting user namespaces")
	headersMap := utils.RequestHeadersMap(ctx)
	headersMap[headers.UserIDXHeader] = user.ID
	headersMap[headers.UserRoleXHeader] = user.Role
	resp, err := c.rest.R().SetContext(ctx).
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
