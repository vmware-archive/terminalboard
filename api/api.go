package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"encoding/json"
)

type PipelineStatusGetter interface {
	GetPipelineStatuses() ([]PipelineStatus, error)
}

func NewRouter(c PipelineStatusGetter, c2 PipelineStatusGetter) (http.Handler, error) {
	routa := mux.NewRouter()
	routa.HandleFunc("/api/pipeline_statuses", AllowCORS(MakePipelineStatusHandler(c.GetPipelineStatuses)))
	routa.HandleFunc("/api/v1/pipeline_statuses", AllowCORS(MakePipelineStatusHandler(c2.GetPipelineStatuses)))

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

func AllowCORS(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
		w.Header().Set("Access-Control-Allow-Method", "GET, OPTIONS")

		h(w, req)
	})
}

