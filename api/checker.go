package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/concourse/atc"
)

type Checker struct {
	pipelinePrefix string
	apiPrefix      string
	username       string
	password       string
}

func NewChecker(host, username, password string) *Checker {
	return &Checker{
		pipelinePrefix: fmt.Sprintf("%s/pipelines", host),
		apiPrefix:      fmt.Sprintf("%s/api/v1/", host),
		username:       username,
		password:       password,
	}
}
func (c *Checker) GetPipelineStatuses() ([]PipelineStatus, error) {
	statuses := c.getPipelineStatuses()
	sort.Sort(PipelineStatuses(statuses))

	return statuses, nil
}

func (c *Checker) getPipelineStatuses() []PipelineStatus {
	fmt.Println("Getting all pipelines")
	pipelines := c.getPipelines()
	fmt.Println(fmt.Sprintf(
		"Getting all pipelines complete, total count: %d", len(pipelines)))

	var statuses []PipelineStatus

	startTime := time.Now()
	fmt.Println("Getting all jobs")

	statusChan := make(chan *PipelineStatus, len(pipelines))

	for _, pipeline := range pipelines {
		go func(pipelineName string) {
			statusChan <- c.getPipelineJobsStatus(pipelineName)
		}(pipeline)
	}

	for i := 0; i < len(pipelines); i++ {
		status := <-statusChan
		if status != nil {
			statuses = append(statuses, *status)
		}
	}

	endTime := time.Now()

	elapsedTime := endTime.Sub(startTime)
	fmt.Println(fmt.Sprintf(
		"Getting all jobs complete, took %f seconds", elapsedTime.Seconds()))
	return statuses
}

func (c *Checker) getPipelines() []string {
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

func (c *Checker) getPipelineJobsStatus(pipeline string) *PipelineStatus {
	jobs := c.getPipelineJobs(pipeline)
	if len(jobs) > 0 {
		status := c.getPipelineStatusFromJobs(pipeline, jobs)
		return &status
	}
	return nil
}

func (c *Checker) getPipelineJobs(pipeline string) []atc.Job {
	pipelineJobsEndpoint := c.apiPrefix + "pipelines/" + pipeline + "/jobs"
	body := c.getFromConcourse(pipelineJobsEndpoint)

	var jobs []atc.Job
	json.Unmarshal(body, &jobs)

	return jobs
}

func (c *Checker) getFromConcourse(endpoint string) []byte {
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

func (c *Checker) getPipelineStatusFromJobs(pipeline string, jobs []atc.Job) PipelineStatus {
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
