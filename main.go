package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mfine30/terminalboard/api"
	capi "github.com/mfine30/terminalboard/concourse/api"
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

	token, err := capi.LoginWithBasicAuth(
		concourseHost,
		defaultTeam,
		concourseUsername,
		concoursePassword,
		true,
	)

	if err != nil {
		panic(err)
	}

	httpClient := capi.OAuthHTTPClient(token, true)

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
