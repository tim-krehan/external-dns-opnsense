package main

import (
	"encoding/json"
	"net/http"
)

func recordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	records := cfg.listEntries()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func applyChangesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ApplyChangesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	for _, rec := range req.Create {
		cfg.createEntry(rec)
	}

	for _, rec := range req.Delete {
		cfg.deleteEntry(rec)
	}

	w.WriteHeader(http.StatusOK)
}
