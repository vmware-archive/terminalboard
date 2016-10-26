package middleware

import "net/http"

func AllowCORS(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
		w.Header().Set("Access-Control-Allow-Method", "GET, OPTIONS")

		h(w, req)
	})
}
