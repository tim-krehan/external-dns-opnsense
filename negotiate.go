package main

import (
	"context"
	"encoding/json"
	"net/http"
)

// negotiateHandler handles HTTP requests to negotiate and retrieve the current configuration state.
// Only GET requests are allowed. It responds with a JSON-encoded configuration state.
func negotiateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	records := ReadEntries(api.WithContext(ctx), "")

	// Set the response content type to JSON and encode the state into the response.
	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	json.NewEncoder(w).Encode(records)
}
