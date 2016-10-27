package middleware

import (
	"fmt"
	"net/http"
	"os"
)

type PanicRecovery struct {
}

func NewPanicRecovery() *PanicRecovery {
	return &PanicRecovery{}
}

func (p PanicRecovery) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if panicInfo := recover(); panicInfo != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(os.Stderr, "Panic while serving request: %+v", panicInfo)
			}
		}()
		next.ServeHTTP(rw, req)
	})
}
