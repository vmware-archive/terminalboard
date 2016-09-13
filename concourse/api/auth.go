package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
)

type TargetToken struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func LoginWithBasicAuth(
	url string,
	teamName string,
	username string,
	password string,
	insecure bool,
) (TargetToken, error) {
	c := basicAuthHttpClient(username, password, insecure)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/teams/%s/auth/token", url, teamName), nil)
	if err != nil {
		return TargetToken{}, err
	}

	res, err := c.Do(req)
	if err != nil {
		return TargetToken{}, err
	}
	defer res.Body.Close()

	var token TargetToken
	err = json.NewDecoder(res.Body).Decode(&token)

	return token, err
}

func OAuthHTTPClient(token TargetToken, insecure bool) *http.Client {
	var transport http.RoundTripper

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	oAuthToken := &oauth2.Token{
		TokenType:   token.Type,
		AccessToken: token.Value,
	}

	transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(oAuthToken),
		Base:   transport,
	}

	return &http.Client{Transport: transport}
}

func basicAuthHttpClient(
	username string,
	password string,
	insecure bool,
) *http.Client {
	return &http.Client{
		Transport: basicAuthTransport{
			username: username,
			password: password,
			base: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
				Dial: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).Dial,
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

type basicAuthTransport struct {
	username string
	password string

	base http.RoundTripper
}

func (t basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.username, t.password)
	return t.base.RoundTrip(r)
}
