package main

import (
	"context"
	"encoding/json"
	externaldns "external-dns-opnsense/externaldns"
	"net/http"
)

// negotiateHandler handles HTTP requests to negotiate and retrieve the current configuration state.
// Only GET requests are allowed. It responds with a JSON-encoded configuration state.
func negotiateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the current configuration state.
	var state string // Replace with actual state retrieval logic if needed.

	// Set the response content type to JSON and encode the state into the response.
	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	json.NewEncoder(w).Encode(state)
}

// recordsHandler handles HTTP requests to retrieve DNS records.
// Only POST requests are allowed. It responds with a JSON-encoded list of records.
func recordsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		http.Error(w, "POST not implemented", http.StatusNotImplemented)
	case http.MethodGet:
		// Create a new context for this request
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()

		// Retrieve the list of DNS records using the new context
		records := api.WithContext(ctx).ListEntries()

		// Set the response content type to JSON and encode the records into the response
		w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
		json.NewEncoder(w).Encode(records)
	}
}

// adjustendpointsHandler handles HTTP requests to apply changes to DNS records.
// Only POST requests are allowed. It processes create and delete operations based on the request body.
func adjustendpointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a new context for this request
	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()

	// Decode the JSON request body into an ApplyChangesRequest struct.
	var req externaldns.ApplyChangesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process the create operations.
	for _, rec := range req.Create {
		api.WithContext(ctx).CreateEntry(rec)
	}

	// Process the delete operations.
	for _, rec := range req.Delete {
		api.WithContext(ctx).DeleteEntry(rec)
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	w.WriteHeader(http.StatusOK)
}
