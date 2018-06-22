package clients

import (
	"context"

	"time"

	"git.containerum.net/ch/auth/proto"
	"github.com/containerum/cherry"
	"github.com/containerum/utils/httputil"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

// AuthClient is an extension of AuthClient interface with Closer interface to close connection
type AuthClient interface {
	CreateToken(ctx context.Context, in *authProto.CreateTokenRequest) (*authProto.CreateTokenResponse, error)
	DeleteToken(ctx context.Context, in *authProto.DeleteTokenRequest) (*empty.Empty, error)
	DeleteUserTokens(ctx context.Context, in *authProto.DeleteUserTokensRequest) (*empty.Empty, error)
}

type httpAuthClient struct {
	log    *logrus.Entry
	client *resty.Client
}

// NewHTTPAuthClient returns client for auth service works using http protocol
func NewHTTPAuthClient(serverURL string) (AuthClient, error) {
	log := logrus.WithField("component", "auth_client")
	client := resty.New().
		SetHostURL(serverURL).
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetTimeout(3 * time.Second).
		SetError(cherry.Err{})
	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal
	return &httpAuthClient{
		log:    log,
		client: client,
	}, nil
}

func (c *httpAuthClient) CreateToken(ctx context.Context, in *authProto.CreateTokenRequest) (*authProto.CreateTokenResponse, error) {
	c.log.Debugf("create token %+v", in)

	headersMap := httputil.RequestXHeadersMap(ctx)
	headersMap[httputil.UserAgentXHeader] = in.GetUserAgent()
	headersMap[httputil.UserClientXHeader] = in.GetFingerprint()
	headersMap[httputil.UserIDXHeader] = in.GetUserId()
	headersMap[httputil.UserIPXHeader] = in.GetUserIp()
	headersMap[httputil.UserRoleXHeader] = in.GetUserRole()
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeaders(headersMap).
		SetResult(authProto.CreateTokenResponse{}).
		Post("/token")
	if err != nil {
		return nil, err
	}
	if resp.Error() != nil {
		return nil, resp.Error().(*cherry.Err)
	}
	return resp.Result().(*authProto.CreateTokenResponse), nil
}

func (c *httpAuthClient) DeleteToken(ctx context.Context, in *authProto.DeleteTokenRequest) (*empty.Empty, error) {
	c.log.WithFields(logrus.Fields{
		"token_id": in.GetTokenId(),
		"user_id":  in.GetUserId(),
	}).Debugf("delete token")

	zero := &empty.Empty{}
	headersMap := httputil.RequestXHeadersMap(ctx)
	headersMap[httputil.UserIDXHeader] = in.GetUserId()
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeaders(headersMap).
		SetPathParams(map[string]string{"token_id": in.GetTokenId()}).
		Delete("/token/{token_id}")
	if err != nil {
		return zero, err
	}
	if resp.Error() != nil {
		return zero, resp.Error().(*cherry.Err)
	}
	return zero, nil
}

func (c *httpAuthClient) DeleteUserTokens(ctx context.Context, in *authProto.DeleteUserTokensRequest) (*empty.Empty, error) {
	c.log.WithField("user_id", in.GetUserId()).Debugf("delete user tokens")

	zero := &empty.Empty{}
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeaders(httputil.RequestXHeadersMap(ctx)).
		SetPathParams(map[string]string{"user_id": in.GetUserId()}).
		Delete("/user/{user_id}/tokens")
	if err != nil {
		return zero, err
	}
	if resp.Error() != nil {
		return zero, resp.Error().(*cherry.Err)
	}
	return zero, nil
}
