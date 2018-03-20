package clients

import (
	"context"
	"os"
	"time"

	"fmt"

	"crypto/sha256"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"

	"net/http"
	"net/rpc/jsonrpc"

	kube_types "git.containerum.net/ch/kube-client/pkg/model"

	"errors"

	"git.containerum.net/ch/auth/proto"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/user-manager"
)

type WebAPIError struct {
	Error string `json:"message"`
}

// WebAPIClient is an interface for web-api service from old architecture.
type WebAPIClient interface {
	Login(ctx context.Context, request *umtypes.LoginRequest) (ret *umtypes.WebAPILoginResponse, err error)
	GetVolumes(ctx context.Context, token string, userID string) (ret []*authProto.AccessObject, err error)
	GetNamespaces(ctx context.Context, token string, userID string) (ret []*authProto.AccessObject, err error)
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

	resp, err := c.client.R().SetContext(ctx).SetBody(kube_types.Login{Login: request.Login, Password: request.Password}).SetResult(&ret).Post("/api/login")
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

func (c *httpWebAPIClient) GetVolumes(ctx context.Context, token string, userID string) (ret []*authProto.AccessObject, err error) {
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
		ret = append(ret, &authProto.AccessObject{Id: fmt.Sprintf("%x", (sha256.Sum256([]byte(userID + v.Name)))), Label: v.Name, Access: "owner"})
	}

	return ret, nil
}

func (c *httpWebAPIClient) GetNamespaces(ctx context.Context, token string, userID string) (ret []*authProto.AccessObject, err error) {
	c.log.Infoln("Getting namespaces")

	serverAddr := os.Getenv("CH_STORE_ADDRESS")

	type Namespace struct {
		Id             *string
		Label          *string
		UserId         *string
		Created        *time.Time
		Active         *bool
		Removed        *bool
		KubeExist      *bool
		MemoryLimit    *int
		CpuLimit       *int
		LimitUpdated   *time.Time
		MaxServices    *int
		TrafficLimit   *int
		MaxInternalSvc *int
	}

	// client, err := rpc.Dial("tcp", serverAddr)
	client, err := jsonrpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	namespaces := new([]Namespace)
	if err := client.Call("Namespace.GetAll", userID, namespaces); err != nil {
		c.log.WithError(err).Warn("Unable to get namespace from store")
		return nil, errors.New("Unable to get namespaces from store")
	}
	if namespaces == nil {
		return nil, errors.New("Unable to get namespaces from store. Namespaces in nil")
	}

	for _, ns := range *namespaces {
		ret = append(ret, &authProto.AccessObject{
			Id:     *ns.Id,
			Label:  *ns.Label,
			Access: "owner",
		})
		c.log.WithFields(logrus.Fields{
			"ID":    *ns.Id,
			"Label": *ns.Label,
		}).Debug("Append NS Info")
	}

	return ret, nil
}
