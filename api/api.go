package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type RunFunc func() ([]byte, error)

type router struct {
	checker Checker
}

func NewRouter(c Checker) (http.Handler, error) {

	r := router{
		checker: c,
	}

	routa := mux.NewRouter()
	routa.HandleFunc("/api/pipeline_statuses", r.getHandler(c.GetPipelineStatuses))
	routa.HandleFunc("/api/fakes", r.getHandler(c.FakeStatuses))

	return routa, nil
}

func (r *router) getHandler(run RunFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
		w.Header().Set("Access-Control-Allow-Method", "GET, OPTIONS")
		w.Write(body)
	})
}
