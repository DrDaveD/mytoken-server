package oauth2x

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type Config struct {
	Issuer    string
	Ctx       context.Context
	endpoints *Endpoints
}

type Endpoints struct {
	Authorization string `json:"authorization_endpoint"`
	Token         string `json:"token_endpoint"`
	Userinfo      string `json:"userinfo_endpoint"`
	Registration  string `json:"registration_endpoint"`
	Revocation    string `json:"revocation_endpoint"`
	Introspection string `json:"introspection_endpoint"`
}

func (e *Endpoints) OAuth2() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  e.Authorization,
		TokenURL: e.Token,
	}
}

func (c *Config) Endpoints() (*Endpoints, error) {
	var err error
	if c.endpoints == nil {
		err = c.discovery()
	}
	return c.endpoints, err
}

func NewConfig(ctx context.Context, issuer string) *Config {
	c := &Config{
		Issuer: issuer,
		Ctx:    ctx,
	}
	c.discovery()
	return c
}

func (c *Config) discovery() error {
	wellKnown := strings.TrimSuffix(c.Issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequest("GET", wellKnown, nil)
	if err != nil {
		return err
	}
	resp, err := doRequest(c.Ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", resp.Status, body)
	}

	var endpoints Endpoints
	if err := json.Unmarshal(body, &endpoints); err != nil {
		return fmt.Errorf("oauth2x: failed to decode provider discovery object: %v", err)
	}
	c.endpoints = &endpoints
	return nil
}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := http.DefaultClient
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = c
	}
	return client.Do(req.WithContext(ctx))
}
