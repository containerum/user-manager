package clients

import (
	"encoding/json"

	"git.containerum.net/ch/json-types/errors"
	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/json-iterator/go"
	"gopkg.in/resty.v1"
)

// OAuthUserInfo describes information about user needed to login via 3rd-party resource
type OAuthUserInfo struct {
	UserID string
	Email  string
}

// OAuthClient is an interface to 3rd-party resource for fetching information needed for login
type OAuthClient interface {
	GetUserInfo(ctx context.Context, authCode string) (*OAuthUserInfo, *errors.Error)
	GetResource() umtypes.OAuthResource
}

var oAuthClients = make(map[umtypes.OAuthResource]OAuthClient)

// OAuthClientByResource returns oauth client for service by it`s name.
// Client for resource must be registered using RegisterOAuthClient
func OAuthClientByResource(resource umtypes.OAuthResource) (client OAuthClient, exists bool) {
	client, exists = oAuthClients[resource]
	return
}

// RegisterOAuthClient registers an oauth client for resource.
func RegisterOAuthClient(client OAuthClient) {
	oAuthClients[client.GetResource()] = client
}

type oAuthClientConfig struct {
	log  *logrus.Entry
	rest *resty.Client
}

type githubOAuthClient struct {
	oAuthClientConfig
}

// NewGithubOAuthClient returns resty client for http://github.com
func NewGithubOAuthClient() OAuthClient {
	log := logrus.WithField("component", "github_client")
	client := resty.New().
		SetHostURL("https://api.github.com").
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetHeader("Content-Type", "application/json")

	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal

	return &githubOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log:  log,
			rest: client,
		},
	}
}

func (gh *githubOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.GitHubOAuth
}

type githubError struct {
	Message string `json:"message"`
}

type githubResponce struct {
	ID    json.Number `json:"id"`
	Email string      `json:"email,omitempty"`
}

func (gh *githubOAuthClient) GetUserInfo(ctx context.Context, authCode string) (*OAuthUserInfo, *errors.Error) {
	gh.log.Info("Getting user info from github")

	resp, err := gh.rest.R().SetContext(ctx).
		SetQueryParam("access_token", authCode).
		SetError(githubError{}).
		SetResult(githubResponce{}).
		Get("/user")

	if err != nil {
		gh.log.WithError(err)
		return nil, errors.New(err.Error())
	}

	if resp.Error().(*githubError).Message != "" {
		gh.log.Errorln(resp.Error().(*githubError).Message)
		return nil, errors.NewWithCode(resp.Error().(*githubError).Message, resp.StatusCode())
	}

	return &OAuthUserInfo{
		UserID: resp.Result().(*githubResponce).ID.String(),
		Email:  resp.Result().(*githubResponce).Email,
	}, nil
}

type googleOAuthClient struct {
	oAuthClientConfig
}

// NewGoogleOAuthClient returns resty client for http://google.com
func NewGoogleOAuthClient() OAuthClient {
	log := logrus.WithField("component", "google_client")
	client := resty.New().
		SetHostURL("https://www.googleapis.com/oauth2/v2").
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetHeader("Content-Type", "application/json")

	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal

	return &googleOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log:  log,
			rest: client,
		},
	}
}

func (gc *googleOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.GoogleOAuth
}

type googleError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type googleResponse struct {
	Email         string `json:"email"`
	ID            string `json:"id"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (gc *googleOAuthClient) GetUserInfo(ctx context.Context, authCode string) (*OAuthUserInfo, *errors.Error) {
	gc.log.Info("Getting user info from Google")

	resp, err := gc.rest.R().SetContext(ctx).
		SetQueryParam("access_token", authCode).
		SetResult(googleResponse{}).
		SetError(googleError{}).
		Get("/userinfo")

	if err != nil {
		gc.log.WithError(err)
		return nil, errors.New(err.Error())
	}

	if resp.Error().(*googleError).Error.Code != 0 {
		gc.log.Errorln(resp.Error().(*googleError).Error)
		return nil, errors.NewWithCode(resp.Error().(*googleError).Error.Message, resp.StatusCode())
	}

	if !resp.Result().(*googleResponse).VerifiedEmail {
		return &OAuthUserInfo{
			UserID: resp.Result().(*googleResponse).ID,
		}, nil
	}

	return &OAuthUserInfo{
		UserID: resp.Result().(*googleResponse).ID,
		Email:  resp.Result().(*googleResponse).Email,
	}, nil
}

type facebookOAuthClient struct {
	oAuthClientConfig
}

// NewFacebookOAuthClient returns resty client for http://facebook.com
func NewFacebookOAuthClient() OAuthClient {
	log := logrus.WithField("component", "facebook_client")
	client := resty.New().
		SetHostURL("https://graph.facebook.com/v2.11").
		SetLogger(log.WriterLevel(logrus.DebugLevel)).
		SetDebug(true).
		SetHeader("Content-Type", "application/json")

	client.JSONMarshal = jsoniter.Marshal
	client.JSONUnmarshal = jsoniter.Unmarshal

	return &facebookOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log:  log,
			rest: client,
		},
	}
}

func (fb *facebookOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.FacebookOAuth
}

type facebookError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type facebookResponse struct {
	Email string `json:"email"`
	ID    string `json:"id"`
}

func (fb *facebookOAuthClient) GetUserInfo(ctx context.Context, authCode string) (*OAuthUserInfo, *errors.Error) {
	fb.log.Info("Getting user info from facebook")

	resp, err := fb.rest.R().SetContext(ctx).
		SetQueryParam("fields", "id,email").
		SetQueryParam("access_token", authCode).
		SetResult(facebookResponse{}).
		SetError(facebookError{}).
		Get("/me")

	if err != nil {
		fb.log.WithError(err)
		return nil, errors.New(err.Error())
	}

	if resp.Error().(*facebookError).Error.Code != 0 {
		fb.log.Errorln(resp.Error().(*facebookError).Error)
		if resp.Error().(*facebookError).Error.Code == 190 {
			return nil, errors.NewWithCode(resp.Error().(*facebookError).Error.Message, 403)
		}

		return nil, errors.NewWithCode(resp.Error().(*facebookError).Error.Message, resp.StatusCode())
	}

	return &OAuthUserInfo{
		UserID: resp.Result().(*facebookResponse).ID,
		Email:  resp.Result().(*facebookResponse).Email,
	}, nil
}
