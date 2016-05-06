package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mfine30/terminalboard/api"
)

func main() {
	concourseAPIPrefix := fmt.Sprintf("%s/api/v1/", os.Getenv("CONCOURSE_HOST"))
	concourseUsername := os.Getenv("CONCOURSE_USERNAME")
	concoursePassword := os.Getenv("CONCOURSE_PASSWORD")

	port := os.Getenv("PORT")

	address := "localhost:" + port
	fmt.Print(address)

	checker := api.NewChecker(concourseAPIPrefix, concourseUsername, concoursePassword)
	router, err := api.NewRouter(checker)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(address, router)
	// err = http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		panic(err)
	}
}
