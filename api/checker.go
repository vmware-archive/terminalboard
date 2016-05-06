package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/concourse/atc"
)

const (
	SUCCESS = "success"
	FAILURE = "failure"
	STOPPED = "stopped"
)

type Checker interface {
	GetPipelineStatuses() ([]byte, error)
	FakeStatuses() ([]byte, error)
}

type checker struct {
	pipelinePrefix string
	apiPrefix      string
	username       string
	password       string
}

type PipelineStatus struct {
	Name             string `json:"pipelineName"`
	Status           string `json:"pipelineStatus"`
	CurrentlyRunning bool   `json:"currentlyRunning"`
	URL              string `json:"url"`
}

func NewChecker(host, username, password string) Checker {
	return &checker{
		pipelinePrefix: fmt.Sprintf("%s/pipelines", host),
		apiPrefix:      fmt.Sprintf("%s/api/v1/", host),
		username:       username,
		password:       password,
	}
}

func (c *checker) FakeStatuses() ([]byte, error) {
	statuses := []PipelineStatus{
		PipelineStatus{
			Name:             "first",
			Status:           SUCCESS,
			CurrentlyRunning: false,
		},
		PipelineStatus{
			Name:             "second",
			Status:           FAILURE,
			CurrentlyRunning: true,
		},
		PipelineStatus{
			Name:             "another-pipeline-long-name",
			Status:           STOPPED,
			CurrentlyRunning: false,
		},
		PipelineStatus{
			Name:             "yet-another-pipeline-long-name",
			Status:           SUCCESS,
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
	statuses := c.getPipelineStatuses()
	data, err := json.Marshal(statuses)
	if err != nil {
		panic(err)
	}
	return data, nil
}

func (c *checker) getPipelineStatuses() []PipelineStatus {
	pipelines := c.getPipelines()

	var statuses []PipelineStatus

	for _, pipeline := range pipelines {
		jobs := c.getPipelineJobs(pipeline)
		if len(jobs) > 0 {
			status := c.getPipelineStatusFromJobs(pipeline, jobs)
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

func (c *checker) getPipelineStatusFromJobs(pipeline string, jobs []atc.Job) PipelineStatus {
	pipelineStatus := PipelineStatus{
		Name:             pipeline,
		Status:           SUCCESS,
		CurrentlyRunning: false,
		URL:              fmt.Sprintf("%s/%s", c.pipelinePrefix, pipeline),
	}

	for _, job := range jobs {
		if pipelineStatus.Status == FAILURE && pipelineStatus.CurrentlyRunning == true {
			return pipelineStatus
		}
		if nextBuild := job.NextBuild; nextBuild != nil && nextBuild.Status == "started" {
			pipelineStatus.CurrentlyRunning = true
		}

		if job.FinishedBuild != nil {
			jobStatus := job.FinishedBuild.Status

			if pipelineStatus.Status != FAILURE {
				switch jobStatus {
				case "failed":
					pipelineStatus.Status = FAILURE
				case "errored":
					pipelineStatus.Status = FAILURE
				case "paused":
					pipelineStatus.Status = STOPPED
				case "aborted":
					pipelineStatus.Status = STOPPED
				}
			}
		}
	}

	return pipelineStatus
}
