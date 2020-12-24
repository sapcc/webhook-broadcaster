package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/concourse/concourse/go-concourse/concourse"
	"golang.org/x/oauth2"
)

type client struct {
	concourseURL string
	username     string
	password     string
	oauth2Config *oauth2.Config
	token        *oauth2.Token
	ctx          context.Context
}

func NewConcourseClient(concourseURL string, username string, password string) (*client, error) {
	c := client{
		concourseURL: concourseURL,
		username:     username,
		password:     password,
	}

	tokenEndPoint, err := url.Parse("sky/issuer/token")
	if err != nil {
		return &client{}, err
	}

	base, err := url.Parse(concourseURL)
	if err != nil {
		return &client{}, err
	}

	tokenURL := base.ResolveReference(tokenEndPoint)

	/* We leverage the fact that `fly` is considered a "public client" to fetch our oauth token */
	c.oauth2Config = &oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL.String()},
		Scopes:       []string{"openid", "profile", "email", "federated:id", "groups"},
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}
	c.ctx = context.WithValue(context.Background(), oauth2.HTTPClient, &httpClient)

	return &c, nil
}

func (c *client) RefreshClientWithToken() (concourse.Client, error) {
	if !c.token.Valid() {
		var err error
		c.token, err = c.oauth2Config.PasswordCredentialsToken(c.ctx, c.username, c.password)
		if err != nil {
			return nil, err
		}
	}

	httpClient := c.oauth2Config.Client(c.ctx, c.token)
	concourseClient := concourse.NewClient(c.concourseURL, httpClient, false)

	return concourseClient, nil
}
