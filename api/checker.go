package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/mfine30/terminalboard/concourse/models"
)

type Checker struct {
	Host string
	apiPrefix      string
	Client *http.Client
}

func NewChecker(host, team string, client *http.Client) *Checker {
	return &Checker{
		Host: host,
		apiPrefix:      fmt.Sprintf("%s/api/v1/teams/%s/", host, team),
		Client:         client,
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
	if len(pipelines) == 0 {
		panic("No pipelines found")
	}

	fmt.Println(fmt.Sprintf(
		"Getting all pipelines complete, total count: %d", len(pipelines)))

	var statuses []PipelineStatus

	startTime := time.Now()
	fmt.Println("Getting all jobs")

	statusChan := make(chan *PipelineStatus, len(pipelines))

	for _, pipeline := range pipelines {
		go func(pipeline models.Pipeline) {
			statusChan <- c.getPipelineJobsStatus(pipeline)
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

func (c *Checker) getPipelines() []models.Pipeline {
	pipelinesEndpoint := c.apiPrefix + "pipelines"

	body := c.getFromConcourse(pipelinesEndpoint)

	var pipelines []models.Pipeline
	json.Unmarshal(body, &pipelines)

	return pipelines
}

func (c *Checker) getPipelineJobsStatus(pipeline models.Pipeline) *PipelineStatus {
	jobs := c.getPipelineJobs(pipeline.Name)
	if len(jobs) > 0 {
		status := c.getPipelineStatusFromJobs(pipeline, jobs)
		return &status
	}
	return nil
}

func (c *Checker) getPipelineJobs(pipeline string) []models.Job {
	pipelineJobsEndpoint := c.apiPrefix + "pipelines/" + pipeline + "/jobs"
	body := c.getFromConcourse(pipelineJobsEndpoint)

	var jobs []models.Job
	json.Unmarshal(body, &jobs)

	return jobs
}

func (c *Checker) getFromConcourse(endpoint string) []byte {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		panic(err)
	}

	res, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(fmt.Sprintf("Received non-200 response, %d", res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func (c *Checker) getPipelineStatusFromJobs(pipeline models.Pipeline, jobs []models.Job) PipelineStatus {
	pipelineStatus := PipelineStatus{
		Name:             pipeline.Name,
		Status:           SUCCESS,
		CurrentlyRunning: false,
		URL:              c.Host + pipeline.URL,
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
