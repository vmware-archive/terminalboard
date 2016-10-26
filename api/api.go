package api

import (
	"net/http"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf/terminalboard/api/middleware"
)

type PipelineStatusGetter interface {
	GetPipelineStatuses() ([]PipelineStatus, error)
}

func NewRouter(c PipelineStatusGetter) (http.Handler, error) {
	pr := middleware.NewPanicRecovery()

	routa := mux.NewRouter()
	routa.Handle("/api/pipeline_statuses", pr.Wrap(middleware.AllowCORS(MakePipelineStatusHandler(c.GetPipelineStatuses))))

	return routa, nil
}

func MakePipelineStatusHandler(run func() ([]PipelineStatus, error)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		statuses, err := run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		json.NewEncoder(w).Encode(statuses)
	})
}
