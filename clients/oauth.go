package clients

import (
	"fmt"
	"strconv"

	"github.com/google/go-github/github"
	"github.com/huandu/facebook"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	google "google.golang.org/api/oauth2/v2"
)

type OAuthUserInfo struct {
	UserID string
	Email  string
}

type OAuthClient interface {
	GetUserInfo(accessToken string) (info *OAuthUserInfo, err error)
}

type OAuthResource string

const (
	GitHubOAuth   OAuthResource = "github"
	GoogleOAuth   OAuthResource = "google"
	FacebookOAuth OAuthResource = "facebook"
)

var oAuthClients = make(map[OAuthResource]OAuthClient)

func OAuthClientByResource(resource OAuthResource) (client OAuthClient, exists bool) {
	client, exists = oAuthClients[resource]
	return
}

func init() {
	oAuthClients[GitHubOAuth] = &githubOAuthClient{log: logrus.WithField("component", "github_oauth").Logger}
	oAuthClients[GoogleOAuth] = &googleOAuthClient{log: logrus.WithField("component", "google_oauth").Logger}
	oAuthClients[FacebookOAuth] = &facebookOAuthClient{log: logrus.WithField("component", "facebook_oauth").Logger}
}

type githubOAuthClient struct {
	log *logrus.Logger
}

func (gh *githubOAuthClient) GetUserInfo(accessToken string) (info *OAuthUserInfo, err error) {
	gh.log.WithField("token", accessToken).Info("Get GitHub user info")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		gh.log.WithError(err).Error("Request error")
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
	log *logrus.Logger
}

func (gc *googleOAuthClient) GetUserInfo(accessToken string) (info *OAuthUserInfo, err error) {
	gc.log.WithField("token", accessToken).Info("Get Google user info")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client, err := google.New(tc)
	if err != nil {
		gc.log.WithError(err).Error("Client create failed")
		return nil, err
	}

	googleInfo, err := google.NewUserinfoV2MeService(client).Get().Do()
	if err != nil {
		gc.log.WithError(err).Error("Fetch user info failed")
		return nil, err
	}

	return &OAuthUserInfo{
		UserID: googleInfo.Id,
		Email:  googleInfo.Email,
	}, nil
}

type facebookOAuthClient struct {
	log *logrus.Logger
}

func (fb *facebookOAuthClient) GetUserInfo(accessToken string) (info *OAuthUserInfo, err error) {
	fb.log.WithField("token", accessToken).Info("Get Facebook user info")

	resp, err := facebook.Get("/me", facebook.Params{
		"access_token": accessToken,
		"fields":       "id,email",
	})
	if err != nil {
		fb.log.WithError(err).Error("Fetch user info failed")
		return nil, err
	}

	return &OAuthUserInfo{
		UserID: resp.Get("id").(string),
		Email:  resp.Get("email").(string),
	}, nil
}
