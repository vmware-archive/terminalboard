package api

const (
	SUCCESS = "success"
	FAILURE = "failure"
	STOPPED = "stopped"
)

type PipelineStatuses []PipelineStatus

type PipelineStatus struct {
	Name             string `json:"pipelineName"`
	Status           string `json:"pipelineStatus"`
	CurrentlyRunning bool   `json:"currentlyRunning"`
	URL              string `json:"url"`
}

func (ps PipelineStatuses) Len() int {
	return len(ps)
}

func (ps PipelineStatuses) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func (ps PipelineStatuses) Less(i, j int) bool {
	if ps[i].Status == FAILURE && ps[j].Status != FAILURE {
		return true
	} else if ps[i].Status == STOPPED && ps[j].Status == FAILURE {
		return false
	} else if ps[i].Status == STOPPED && ps[j].Status != STOPPED {
		return true
	} else {
		return false
	}
}
