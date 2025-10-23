package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	opnsense "external-dns-opnsense/opnsense"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

// recordsHandler handles HTTP requests to retrieve DNS records.
// Only POST requests are allowed. It responds with a JSON-encoded list of records.
func recordsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var changes plan.Changes
		if err := json.NewDecoder(r.Body).Decode(&changes); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		// Create a new context for this request
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		// Apply the changes using the new context
		if err := ApplyChanges(api.WithContext(ctx), changes); err != nil {
			http.Error(w, "Error applying changes", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		// Create a new context for this request
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		// Retrieve the list of DNS records using the new context
		records := ReadEntries(api.WithContext(ctx), "")

		// Set the response content type to JSON and encode the records into the response
		w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
		json.NewEncoder(w).Encode(records)
	default:
		http.Error(w, "Only GET and POST allowed", http.StatusMethodNotAllowed)
	}
}

func ApplyChanges(api *opnsense.OpnSenseApi, changes plan.Changes) []error {
	var errors []error
	for _, create := range changes.Create {
		if err := CreateEntry(api, create); err != nil {
			log.Printf("Error creating entry %v: %v", create, err)
			errors = append(errors, err)
		}
	}
	for _, delete := range changes.Delete {
		if err := DeleteEntry(api, delete); err != nil {
			log.Printf("Error deleting entry %v: %v", delete, err)
			errors = append(errors, err)
		}
	}
	for _, update := range changes.UpdateNew {
		if err := UpdateEntry(api, update); err != nil {
			log.Printf("Error updating entry %v: %v", update, err)
			errors = append(errors, err)
		}
	}
	return errors
}

func ReadEntries(api *opnsense.OpnSenseApi, searchString string) []*endpoint.Endpoint {
	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	overrides, err := opnsense.SearchHostOverrides(api.WithContext(ctx), searchString)
	if err != nil {
		log.Printf("List: Error retrieving host overrides: %v\n", err)
		return []*endpoint.Endpoint{}
	}
	endpoints := []*endpoint.Endpoint{}
	for _, r := range overrides {
		ttl := int64(0)
		targets := []string{}
		if r.TTL != "" {
			ttl, err = strconv.ParseInt(r.TTL, 10, 64)
			if err != nil {
				log.Printf("Error converting TTL to int: %v", err)
				continue
			}
		}
		switch r.Type {
		case "A":
			targets = append(targets, r.Server)
		case "AAAA":
			targets = append(targets, r.Server)
		case "MX":
			targets = append(targets, r.Mx)
		case "TXT":
			targets = append(targets, r.TxtData)
		default:
		}
		endpoint := endpoint.Endpoint{
			DNSName:    r.HostName + "." + r.Domain,
			RecordType: r.Type,
			Targets:    targets,
			RecordTTL:  endpoint.TTL(ttl),
			ProviderSpecific: endpoint.ProviderSpecific{endpoint.ProviderSpecificProperty{
				Name:  "uuid",
				Value: r.Uuid,
			}},
		}
		endpoints = append(endpoints, &endpoint)
	}
	log.Printf("List: Retrieved %d records\n", len(endpoints))
	endpoints = MergeRecordsWithSameFQDN(endpoints)
	if err != nil {
		log.Printf("List: Error merging records: %v\n", err)
		return []*endpoint.Endpoint{}
	}
	return endpoints
}

func CreateEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	// Implementation to create a DNS entry in OpnSense based on the endpoint details.
	// This is a placeholder and should be replaced with actual API calls to create the record.
	log.Printf("Creating entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	hostname := strings.Split(ep.DNSName, ".")[0]
	domain := strings.Join(strings.Split(ep.DNSName, ".")[1:], ".")
	override := opnsense.OpnSenseHostOverride{
		HostName:    hostname,
		Domain:      domain,
		Type:        ep.RecordType,
		TTL:         strconv.FormatInt(int64(ep.RecordTTL), 10),
		Enabled:     "1",
		Description: "Created by external-dns",
	}

	// assign target here
	return nil

	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	err := override.Create(api.WithContext(ctx))
	if err != nil {
		log.Printf("CreateEntry: Error creating host override: %v\n", err)
		return err
	}
	return nil
}

func UpdateEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	// Implementation to update a DNS entry in OpnSense based on the endpoint details.
	// This is a placeholder and should be replaced with actual API calls to update the record.
	log.Printf("Updating entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	return nil
}

func DeleteEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	// Implementation to delete a DNS entry in OpnSense based on the endpoint details.
	// This is a placeholder and should be replaced with actual API calls to delete the record.
	log.Printf("Deleting entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	return nil
}

func MergeRecordsWithSameFQDN(records []*endpoint.Endpoint) []*endpoint.Endpoint {
	if len(records) == 0 {
		return records
	}

	recordMap := make(map[string]*endpoint.Endpoint)
	for _, record := range records {
		key := record.DNSName + "|" + record.RecordType
		if existingRecord, exists := recordMap[key]; exists {
			existingRecord.Targets = append(existingRecord.Targets, record.Targets...)
		} else {
			recordMap[key] = record
		}
	}

	// Rebuild the slice from the map
	mergedRecords := []*endpoint.Endpoint{}
	for _, record := range recordMap {
		mergedRecords = append(mergedRecords, record)
	}
	return mergedRecords
}
