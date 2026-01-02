package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	opnsense "external-dns-opnsense/opnsense"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

// recordsHandler handles HTTP requests to retrieve or apply DNS records.
// Supports GET (list records) and POST (apply plan.Changes).
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
		errs := ApplyChanges(api.WithContext(ctx), changes)
		if len(errs) > 0 {
			http.Error(w, "Error applying changes", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		// Create a new context for this request
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		// Retrieve the list of DNS records using the new context
		records := ReadEntries(api.WithContext(ctx), api.OwnerID)

		// Set the response content type to JSON and encode the records into the response
		w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
		json.NewEncoder(w).Encode(records)
	default:
		http.Error(w, "Only GET and POST allowed", http.StatusMethodNotAllowed)
	}
}

func ApplyChanges(api *opnsense.OpnSenseApi, changes plan.Changes) []error {
	var errors []error
	for _, delete := range changes.Delete {
		if err := DeleteEntry(api, delete); err != nil {
			log.Printf("Error deleting entry %v: %v", delete, err)
			errors = append(errors, err)
		}
	}
	for _, create := range changes.Create {
		if err := CreateEntry(api, create); err != nil {
			log.Printf("Error creating entry %v: %v", create, err)
			errors = append(errors, err)
		}
	}
	for _, update := range changes.UpdateNew {
		if err := UpdateEntry(api, update); err != nil {
			log.Printf("Error updating entry %v: %v", update, err)
			errors = append(errors, err)
		}
	}
	if err := api.ApplyChanges(); err != nil {
		log.Printf("Error applying changes to OPNsense: %v", err)
		errors = append(errors, err)
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
				// Do not drop the record on TTL parse error; log and use 0 as TTL
				log.Printf("Error converting TTL to int for %s.%s: %v. Using TTL=0", r.HostName, r.Domain, err)
				ttl = 0
			}
		}
		switch r.Type {
		case "A":
			targets = append(targets, r.Server)
		case "AAAA":
			targets = append(targets, r.Server)
		case "TXT":
			targets = append(targets, r.TxtData)
		default:
		}

		endpoint := endpoint.Endpoint{
			DNSName:    r.HostName + "." + r.Domain,
			RecordType: r.Type,
			Targets:    targets,
			RecordTTL:  endpoint.TTL(ttl),
			Labels: map[string]string{
				"owner": r.Description,
				"uuid":  r.Uuid,
			},
		}
		endpoints = append(endpoints, &endpoint)
	}
	log.Printf("List: Retrieved %d records\n", len(endpoints))
	return endpoints
}

func CreateEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	log.Printf("Creating entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	parts := strings.Split(ep.DNSName, ".")
	if len(parts) < 2 {
		log.Printf("Invalid DNSName: %s", ep.DNSName)
		return fmt.Errorf("invalid DNSName: %s", ep.DNSName)
	}
	hostname := parts[0]
	domain := strings.Join(parts[1:], ".")
	override := opnsense.OpnSenseHostOverride{
		HostName:    hostname,
		Domain:      domain,
		Type:        ep.RecordType,
		TTL:         strconv.FormatInt(int64(ep.RecordTTL), 10),
		Enabled:     "1",
		Description: api.OwnerID,
	}

	for _, target := range ep.Targets {
		switch ep.RecordType {
		case endpoint.RecordTypeA:
			override.Server = target
		case endpoint.RecordTypeAAAA:
			override.Server = target
		case endpoint.RecordTypePTR:
			override.Server = target
		case endpoint.RecordTypeTXT:
			override.TxtData = target
		default:
			log.Printf("Record %s is not supported", ep.RecordType)
		}
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		log.Printf("CreateEntry: Creating host override: %+v\n", override)
		err := override.Create(api.WithContext(ctx))
		if err != nil {
			log.Printf("CreateEntry: Error creating host override: %v\n", err)
			return err
		}
	}

	return nil
}

func UpdateEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	log.Printf("Updating entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	parts := strings.Split(ep.DNSName, ".")
	if len(parts) < 2 {
		log.Printf("Invalid DNSName: %s", ep.DNSName)
		return fmt.Errorf("invalid DNSName: %s", ep.DNSName)
	}
	hostname := parts[0]
	domain := strings.Join(parts[1:], ".")
	override := opnsense.OpnSenseHostOverride{
		Uuid: ep.Labels["uuid"],
		HostName: hostname,
		Domain: domain,
	}
	err := override.Read(api)
	if err != nil {
		log.Printf("UpdateEntry: Error finding existing override with id %s: %v\n", ep.Labels["uuid"], err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	log.Printf("UpdateEntry: Updating host override: %+v\n", override)
	err = override.Update(api.WithContext(ctx))
	if err != nil {
		log.Printf("UpdateEntry: Error updating host override: %v\n", err)
		return err
	}

	return nil
}

func DeleteEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	log.Printf("Deleting entry: %s %s %v %v\n", ep.DNSName, ep.RecordType, ep.Targets, ep.Labels)
	parts := strings.Split(ep.DNSName, ".")
	if len(parts) < 2 {
		log.Printf("Invalid DNSName: %s", ep.DNSName)
		return fmt.Errorf("invalid DNSName: %s", ep.DNSName)
	}
	hostname := parts[0]
	domain := strings.Join(parts[1:], ".")
	override := opnsense.OpnSenseHostOverride{
		Uuid: ep.Labels["uuid"],
		HostName: hostname,
		Domain: domain,
	}
	err := override.Read(api)
	if err != nil {
		log.Printf("DeleteEntry: Error finding existing override with id %s: %v\n", ep.Labels["uuid"], err)
		return err
	}

	if hostname != override.HostName {
		return fmt.Errorf("DeleteEntry: Hostname does not match with expected Value. Hostname: %s, Expected: %s", override.HostName, hostname)
	}
	if domain != override.Domain {
		return fmt.Errorf("DeleteEntry: Domain does not match with expected Value. Domain: %s, Expected: %s", override.Domain, domain)
	}

	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	log.Printf("DeleteEntry: Deleting host override: %+v\n", override)
	err = override.Delete(api.WithContext(ctx))
	if err != nil {
		log.Printf("DeleteEntry: Error deleting host override: %v\n", err)
		return err
	}

	return nil
}

// func FindOverrides(api *opnsense.OpnSenseApi, DNSName string, RecordType string) ([]*opnsense.OpnSenseHostOverride, error) {
// 	searchString := strings.Join(strings.Split(DNSName, "."), " ") + " " + api.OwnerID + " " + string(RecordType)
// 	log.Printf("FindOverrideUUID: searching for overrides matching endpoint %s with search string '%s'", DNSName, searchString)
// 	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
// 	defer cancel()
// 	overrides, err := opnsense.SearchHostOverrides(api.WithContext(ctx), searchString)
// 	if err != nil {
// 		return nil, fmt.Errorf("FindOverrideUUID: error searching host overrides for %s: %v", DNSName, err)
// 	}
// 	if len(overrides) == 0 {
// 		log.Printf("FindOverrideUUID: no overrides found matching endpoint %s", DNSName)
// 		return nil, fmt.Errorf("FindOverrideUUID: no overrides found matching endpoint %s", DNSName)
// 	}
// 	var overridesToReturn []*opnsense.OpnSenseHostOverride
// 	for _, o := range overrides {
// 		if o.HostName+"."+o.Domain != DNSName {
// 			continue
// 		}
// 		if o.Type != string(RecordType) {
// 			continue
// 		}
// 		// Match found
// 		log.Printf("FindOverrideUUID: found matching override with UUID %s for [%s] %s", o.Uuid, RecordType, DNSName)
// 		overridesToReturn = append(overridesToReturn, o)
// 	}
// 	return overridesToReturn, nil
// }
