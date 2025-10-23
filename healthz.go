package main

import "net/http"

// healthzHandler handles HTTP requests for health checks.
// It responds with a 200 OK status to indicate the server is healthy.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
