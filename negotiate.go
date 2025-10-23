package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// negotiateHandler handles HTTP requests to negotiate and retrieve the current configuration state.
// Only GET requests are allowed. It responds with a JSON-encoded configuration state.
func negotiateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Negotiating configuration state")

	config := struct {
		DomainFilter struct {
			Domains []string `json:"domains"`
		} `json:"DomainFilter"`
	}{
		DomainFilter: struct {
			Domains []string `json:"domains"`
		}{
			Domains: api.DNSDomainFilter,
		},
	}

	// Set the response content type to JSON and encode the state into the response.
	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	json.NewEncoder(w).Encode(config)
}
