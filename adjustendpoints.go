package main

import (
	"encoding/json"
	"external-dns-opnsense/opnsense"
	"net/http"

	"sigs.k8s.io/external-dns/endpoint"
)

// adjustendpointsHandler handles HTTP requests to apply changes to DNS records.
// Only POST requests are allowed. It processes create and delete operations based on the request body.
func adjustendpointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var endpoints []*endpoint.Endpoint
	json.NewDecoder(r.Body).Decode(&endpoints)
	adjustedEndpoints, err := AdjustEndpoints(api, endpoints)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	json.NewEncoder(w).Encode(adjustedEndpoints)
}

func AdjustEndpoints(api *opnsense.OpnSenseApi, endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	var createdEndpoints []*endpoint.Endpoint
	var error error

	// filter out CNAME Records

	return createdEndpoints, error
}
