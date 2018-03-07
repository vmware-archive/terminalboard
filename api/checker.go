package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/pivotal-cf/terminalboard/concourse/models"
)

type Checker struct {
	Host      string
	apiPrefix string
	Client    *http.Client
}

func NewChecker(host, team string, client *http.Client) *Checker {
	return &Checker{
		Host:      host,
		apiPrefix: fmt.Sprintf("%s/api/v1/teams/%s/", host, team),
		Client:    client,
	}
}
func (c *Checker) GetPipelineStatuses() ([]PipelineStatus, error) {
	statuses, err := c.getPipelineStatuses()
	if err != nil {
		return nil, err
	}
	sort.Sort(PipelineStatuses(statuses))

	return statuses, nil
}

func (c *Checker) getPipelineStatuses() ([]PipelineStatus, error) {
	fmt.Println("Getting all pipelines")
	pipelines, err := c.getPipelines()
	if err != nil {
		return nil, err
	}

	if len(pipelines) == 0 {
		return nil, fmt.Errorf("No pipelines found")
	}

	fmt.Println(fmt.Sprintf(
		"Getting all pipelines complete, total count: %d", len(pipelines)))

	startTime := time.Now()
	fmt.Println("Getting all jobs")

	statusChan := make(chan *PipelineStatus, len(pipelines))
	errChan := make(chan error, len(pipelines))

	for _, pipeline := range pipelines {
		go func(pipeline models.Pipeline) {
			status, err := c.getPipelineJobsStatus(pipeline)
			statusChan <- status
			errChan <- err
		}(pipeline)
	}

	var statuses []PipelineStatus
	for i := 0; i < len(pipelines); i++ {
		status := <-statusChan
		if status != nil {
			statuses = append(statuses, *status)
		}
	}

	var errors []error
	for i := 0; i < len(pipelines); i++ {
		err := <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return nil, errors[0]
	}

	endTime := time.Now()

	elapsedTime := endTime.Sub(startTime)
	fmt.Println(fmt.Sprintf(
		"Getting all jobs complete, took %f seconds", elapsedTime.Seconds()))
	return statuses, nil
}

func (c *Checker) getPipelines() ([]models.Pipeline, error) {
	pipelinesEndpoint := c.apiPrefix + "pipelines"

	body, err := c.getFromConcourse(pipelinesEndpoint)
	if err != nil {
		return nil, err
	}

	var pipelines []models.Pipeline
	err = json.Unmarshal(body, &pipelines)
	if err != nil {
		return nil, err
	}

	return pipelines, nil
}

func (c *Checker) getPipelineJobsStatus(pipeline models.Pipeline) (*PipelineStatus, error) {
	jobs, err := c.getPipelineJobs(pipeline.Name)
	if err != nil {
		return nil, err
	}

	if len(jobs) > 0 {
		status := c.getPipelineStatusFromJobs(pipeline, jobs)
		return &status, nil
	}
	return nil, nil
}

func (c *Checker) getPipelineJobs(pipeline string) ([]models.Job, error) {
	pipelineJobsEndpoint := c.apiPrefix + "pipelines/" + pipeline + "/jobs"
	body, err := c.getFromConcourse(pipelineJobsEndpoint)
	if err != nil {
		return nil, err
	}

	var jobs []models.Job
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (c *Checker) getFromConcourse(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Sending request to %s", endpoint)
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Received non-200 response: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Received response: %s")
	return body, nil
}

func (c *Checker) getPipelineStatusFromJobs(pipeline models.Pipeline, jobs []models.Job) PipelineStatus {
	pipelineStatus := PipelineStatus{
		Name:             pipeline.Name,
		Status:           SUCCESS,
		CurrentlyRunning: false,
		URL:              c.Host + "/teams/" + pipeline.TeamName + "/pipelines/" + pipeline.Name,
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
