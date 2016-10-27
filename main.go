package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/pivotal-cf/terminalboard/api"
	capi "github.com/pivotal-cf/terminalboard/concourse/api"
	"golang.org/x/oauth2"
)

const (
	concourseHostEnvKey     = "CONCOURSE_HOST"
	concourseUsernameEnvKey = "CONCOURSE_USERNAME"
	concoursePasswordEnvKey = "CONCOURSE_PASSWORD"
	portEnvKey              = "PORT"
	defaultTeam             = "main"
)

func main() {
	concourseHost := os.Getenv(concourseHostEnvKey)
	concourseUsername := os.Getenv(concourseUsernameEnvKey)
	concoursePassword := os.Getenv(concoursePasswordEnvKey)
	port := os.Getenv(portEnvKey)

	if concourseHost == "" {
		panic(fmt.Sprintf("concourseHost must be provided via %s", concourseHostEnvKey))
	}

	if concourseUsername == "" {
		panic(fmt.Sprintf("concourseUsername must be provided via %s", concourseUsernameEnvKey))
	}

	if concoursePassword == "" {
		panic(fmt.Sprintf("concoursePassword must be provided via %s", concoursePasswordEnvKey))
	}

	if port == "" {
		panic(fmt.Sprintf("port must be provided via %s", portEnvKey))
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cts := concourseTokenSource{
		concourseHost:     concourseHost,
		concourseUsername: concourseUsername,
		concoursePassword: concoursePassword,
		defaultTeam:       defaultTeam,
	}

	httpClient := capi.OAuthHTTPClient(oauth2.ReuseTokenSource(nil, cts), true)

	checker := api.NewChecker(concourseHost, defaultTeam, httpClient)

	router, err := api.NewRouter(checker)
	if err != nil {
		panic(err)
	}

	address := fmt.Sprintf(":%s", port)
	fmt.Println(fmt.Sprintf("Listening on %s", address))
	err = http.ListenAndServe(address, router)
	if err != nil {
		panic(err)
	}
}

type concourseTokenSource struct {
	concourseHost     string
	defaultTeam       string
	concourseUsername string
	concoursePassword string
}

func (c concourseTokenSource) Token() (*oauth2.Token, error) {
	token, err := capi.LoginWithBasicAuth(
		c.concourseHost,
		c.defaultTeam,
		c.concourseUsername,
		c.concoursePassword,
		true,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(
			"Error getting token: '%s'", err.Error(),
		))
		return nil, err
	}

	oAuthToken := &oauth2.Token{
		TokenType:   token.Type,
		AccessToken: token.Value,
		Expiry:      time.Now().Add(24 * time.Hour),
	}

	return oAuthToken, nil
}
