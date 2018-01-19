package clients

import (
	"fmt"
	"strconv"

	"net/http"

	umtypes "git.containerum.net/ch/json-types/user-manager"
	"github.com/google/go-github/github"
	"github.com/huandu/facebook"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	facebookOAuth "golang.org/x/oauth2/facebook"
	githubOAuth "golang.org/x/oauth2/github"
	googleOAuth "golang.org/x/oauth2/google"
	google "google.golang.org/api/oauth2/v2"
)

// OAuthUserInfo describes information about user needed to login via 3rd-party resource
type OAuthUserInfo struct {
	UserID string
	Email  string
}

// OAuthClient is an interface to 3rd-party resource for fetching information needed for login
type OAuthClient interface {
	GetUserInfo(ctx context.Context, authCode string) (info *OAuthUserInfo, err error)
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
	log *logrus.Entry
	*oauth2.Config
}

func (c *oAuthClientConfig) exchange(ctx context.Context, authCode string) (*http.Client, error) {
	c.log.WithField("auth_code", authCode).Debugln("exchanging auth code")
	token, err := c.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}
	ts := c.TokenSource(ctx, token)
	tc := oauth2.NewClient(ctx, ts)
	return tc, nil
}

type githubOAuthClient struct {
	oAuthClientConfig
}

// NewGithubOAuthClient returns oauth client for http://github.com
func NewGithubOAuthClient(appID, appSecret string) OAuthClient {
	return &githubOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log: logrus.WithField("component", "github_oauth"),
			Config: &oauth2.Config{
				ClientID:     appID,
				ClientSecret: appSecret,
				Endpoint:     githubOAuth.Endpoint,
				Scopes:       []string{string(github.ScopeUser), string(github.ScopeUserEmail)},
			},
		},
	}
}

func (gh *githubOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.GitHubOAuth
}

func (gh *githubOAuthClient) GetUserInfo(ctx context.Context, authCode string) (info *OAuthUserInfo, err error) {
	gh.log.Infoln("Get GitHub user info")
	tc, err := gh.exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}

	client := github.NewClient(tc)

	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		gh.log.WithError(err).Errorln("Request error")
		return nil, err
	}
	if resp.StatusCode >= 400 {
		gh.log.WithField("error", resp.Status).Errorf("GitHub API error")
		return nil, fmt.Errorf("github API error")
	}

	return &OAuthUserInfo{
		UserID: strconv.Itoa(user.GetID()),
		Email:  user.GetEmail(),
	}, nil
}

type googleOAuthClient struct {
	oAuthClientConfig
}

// NewGoogleOAuthClient returns oauth client for http://google.com
func NewGoogleOAuthClient(appID, appSecret string) OAuthClient {
	return &googleOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log: logrus.WithField("component", "google_oauth"),
			Config: &oauth2.Config{
				ClientID:     appID,
				ClientSecret: appSecret,
				Endpoint:     googleOAuth.Endpoint,
				Scopes:       []string{google.UserinfoProfileScope, google.UserinfoEmailScope},
			},
		},
	}
}

func (gc *googleOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.GoogleOAuth
}

func (gc *googleOAuthClient) GetUserInfo(ctx context.Context, authCode string) (info *OAuthUserInfo, err error) {
	gc.log.Infoln("Get Google user info")
	tc, err := gc.exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}

	client, err := google.New(tc)
	if err != nil {
		gc.log.WithError(err).Errorln("Client create failed")
		return nil, err
	}

	googleInfo, err := google.NewUserinfoV2MeService(client).Get().Do()
	if err != nil {
		gc.log.WithError(err).Errorln("Fetch user info failed")
		return nil, err
	}

	return &OAuthUserInfo{
		UserID: googleInfo.Id,
		Email:  googleInfo.Email,
	}, nil
}

type facebookOAuthClient struct {
	oAuthClientConfig
}

// NewFacebookOAuthClient returns oauth client for http://facebook.com
func NewFacebookOAuthClient(appID, appSecret string) OAuthClient {
	return &facebookOAuthClient{
		oAuthClientConfig: oAuthClientConfig{
			log: logrus.WithField("component", "facebook_oauth"),
			Config: &oauth2.Config{
				ClientID:     appID,
				ClientSecret: appSecret,
				Endpoint:     facebookOAuth.Endpoint,
				Scopes:       []string{"email", "public_profile"},
			},
		},
	}
}

func (fb *facebookOAuthClient) GetResource() umtypes.OAuthResource {
	return umtypes.FacebookOAuth
}

func (fb *facebookOAuthClient) GetUserInfo(ctx context.Context, authCode string) (info *OAuthUserInfo, err error) {
	fb.log.Infoln("Get Facebook user info")

	tc, err := fb.exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}
	session := facebook.Session{
		HttpClient: tc,
		Version:    "v2.4",
	}

	resp, err := session.Get("/me", facebook.Params{
		"fields": "id,email",
	})
	if err != nil {
		fb.log.WithError(err).Errorln("Fetch user info failed")
		return nil, err
	}

	return &OAuthUserInfo{
		UserID: resp.Get("id").(string),
		Email:  resp.Get("email").(string),
	}, nil
}
