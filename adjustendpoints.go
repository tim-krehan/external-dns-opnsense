package main

import (
	"encoding/json"
	"external-dns-opnsense/opnsense"
	"log"
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
	if err := json.NewDecoder(r.Body).Decode(&endpoints); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	adjustedEndpoints, err := AdjustEndpoints(api, endpoints)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	json.NewEncoder(w).Encode(adjustedEndpoints)
}

func AdjustEndpoints(api *opnsense.OpnSenseApi, endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	// Pass through only supported record types; do not drop everything.
	out := make([]*endpoint.Endpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		switch ep.RecordType {
		case endpoint.RecordTypeA, endpoint.RecordTypeAAAA, endpoint.RecordTypeTXT:
			out = append(out, ep)
		default:
			log.Printf("AdjustEndpoints: skipping unsupported record type %s for %s", ep.RecordType, ep.DNSName)
		}
	}
	log.Printf("AdjustEndpoints: accepted %d of %d endpoints", len(out), len(endpoints))
	return out, nil
}
