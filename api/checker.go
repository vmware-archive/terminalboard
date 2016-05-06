package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/concourse/atc"
)

const (
	GREEN  = "GREEN"
	RED    = "RED"
	YELLOW = "YELLOW"
)

type Checker interface {
	GetPipelineStatuses() ([]byte, error)
	FakeStatuses() ([]byte, error)
}

type checker struct {
	apiPrefix string
	username  string
	password  string
}

type PipelineStatus struct {
	Name             string `json:"pipelineName"`
	Status           string `json:"pipelineStatus"`
	CurrentlyRunning bool   `json:"currentlyRunning"`
}

func NewChecker(apiPrefix, username, password string) Checker {
	return &checker{
		apiPrefix: apiPrefix,
		username:  username,
		password:  password,
	}
}

func (c *checker) FakeStatuses() ([]byte, error) {
	statuses := []PipelineStatus{
		PipelineStatus{
			Name:             "first",
			Status:           GREEN,
			CurrentlyRunning: false,
		},
		PipelineStatus{
			Name:             "second",
			Status:           RED,
			CurrentlyRunning: true,
		},
		PipelineStatus{
			Name:             "another-pipeline-long-name",
			Status:           YELLOW,
			CurrentlyRunning: false,
		},
		PipelineStatus{
			Name:             "yet-another-pipeline-long-name",
			Status:           GREEN,
			CurrentlyRunning: true,
		},
	}

	data, err := json.Marshal(statuses)
	if err != nil {
		panic(err)
	}
	return data, nil
}

func (c *checker) GetPipelineStatuses() ([]byte, error) {
	return []byte{}, nil
}

func (c *checker) getPipelineStatuses() []PipelineStatus {
	pipelines := c.getPipelines()

	var statuses []PipelineStatus

	for _, pipeline := range pipelines {
		jobs := c.getPipelineJobs(pipeline)
		if len(jobs) > 0 {
			status := getPipelineStatusFromJobs(pipeline, jobs)
			statuses = append(statuses, status)
		}

	}

	return statuses
}

func (c *checker) getPipelines() []string {
	pipelinesEndpoint := c.apiPrefix + "pipelines"

	body := c.getFromConcourse(pipelinesEndpoint)

	var pipelines []atc.Pipeline
	json.Unmarshal(body, &pipelines)
	fmt.Printf("Results: %v\n", pipelines)

	var pipelineNames []string

	for _, pipeline := range pipelines {
		pipelineNames = append(pipelineNames, pipeline.Name)
	}

	return pipelineNames
}

func (c *checker) getPipelineJobs(pipeline string) []atc.Job {
	pipelineJobsEndpoint := c.apiPrefix + "pipelines/" + pipeline + "/jobs"
	body := c.getFromConcourse(pipelineJobsEndpoint)

	var jobs []atc.Job
	json.Unmarshal(body, &jobs)
	fmt.Printf("Results: %v\n", jobs)

	return jobs
}

func (c *checker) getFromConcourse(endpoint string) []byte {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(c.username, c.password)
	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func getPipelineStatusFromJobs(pipeline string, jobs []atc.Job) PipelineStatus {
	pipelineStatus := PipelineStatus{
		Name:             pipeline,
		Status:           GREEN,
		CurrentlyRunning: false,
	}

	for _, job := range jobs {
		if nextBuild := job.NextBuild; nextBuild != nil && nextBuild.Status == "started" {
			pipelineStatus.CurrentlyRunning = true
		}

		jobStatus := job.FinishedBuild.Status
		if jobStatus != RED {
			switch jobStatus {
			case "failed":
				pipelineStatus.Status = RED
			case "errored":
				pipelineStatus.Status = RED
			case "paused":
				pipelineStatus.Status = YELLOW
			case "aborted":
				pipelineStatus.Status = YELLOW
			}
		}
	}

	return pipelineStatus
}
