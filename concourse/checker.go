package concourse

import (
	gc "github.com/concourse/go-concourse/concourse"
	"github.com/mfine30/terminalboard/api"
)

type Checker struct {
	gc.Client
}

func (c *Checker) GetPipelineStatuses() ([]api.PipelineStatus, error) {
	return nil, nil
}
