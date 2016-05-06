package api

import (
	"net/http"

	"github.com/tedsuo/rata"
)

type RunFunc func() ([]byte, error)

type router struct {
	checker Checker
}

func NewRouter(c Checker) (http.Handler, error) {

	r := router{
		checker: c,
	}

	routes := rata.Routes{
		{Name: "pipeline_statuses", Method: "GET", Path: "/api/pipeline_statuses"},
		{Name: "fake_statuses", Method: "GET", Path: "/api/fakes"},
		// {Name: "root", Method: "GET", Path: "/"},
	}

	handlers := rata.Handlers{
		"pipeline_statuses": r.getHandler(c.GetPipelineStatuses),
		"fake_statuses":     r.getHandler(c.FakeStatuses),
		// "root":              r.getHandler(loadRoot),
	}

	handler, err := rata.NewRouter(routes, handlers)
	if err != nil {
		return nil, err
	}

	return handler, nil
}

func (r *router) getHandler(run RunFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(body)
	})
}
