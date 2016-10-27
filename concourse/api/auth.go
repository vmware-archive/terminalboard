package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
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
	fmt.Println(fmt.Sprintf("Logging in to: '%s'", url))
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

	if res.StatusCode != http.StatusOK {
		return TargetToken{}, fmt.Errorf("Unexpected response code: '%d'", res.StatusCode)
	}

	var token TargetToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		dump, _ := httputil.DumpResponse(res, true)
		fmt.Fprintln(os.Stderr, fmt.Sprintf(
			"Error decoding request: '%s', response dump: '%q'",
			err.Error(),
			dump,
		))
	}

	return token, err
}

func OAuthHTTPClient(tokenSource oauth2.TokenSource, insecure bool) *http.Client {
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	transport := &oauth2.Transport{
		Source: tokenSource,
		Base:   baseTransport,
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
