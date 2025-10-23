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

const RecordDescriptionPrefix = "Created by external-dns - "

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

		ownerId := strings.Replace(r.Description, RecordDescriptionPrefix, "", 1)
		endpoint := endpoint.Endpoint{
			DNSName:    r.HostName + "." + r.Domain,
			RecordType: r.Type,
			Targets:    targets,
			RecordTTL:  endpoint.TTL(ttl),
			ProviderSpecific: endpoint.ProviderSpecific{endpoint.ProviderSpecificProperty{
				Name:  "uuid",
				Value: r.HostName + "." + r.Domain + "," + r.Uuid,
			}},
			Labels: map[string]string{
				"owner": ownerId,
			},
		}
		endpoints = append(endpoints, &endpoint)
	}
	log.Printf("List: Retrieved %d records\n", len(endpoints))
	endpoints = MergeRecordsWithSameFQDN(endpoints)
	return endpoints
}

func CreateEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	log.Printf("Creating entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	hostname := strings.Split(ep.DNSName, ".")[0]
	domain := strings.Join(strings.Split(ep.DNSName, ".")[1:], ".")
	override := opnsense.OpnSenseHostOverride{
		HostName:    hostname,
		Domain:      domain,
		Type:        ep.RecordType,
		TTL:         strconv.FormatInt(int64(ep.RecordTTL), 10),
		Enabled:     "1",
		Description: RecordDescriptionPrefix + ep.Labels["owner"],
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
	var uuids []string
	hostname := strings.Split(ep.DNSName, ".")[0]
	domain := strings.Join(strings.Split(ep.DNSName, ".")[1:], ".")
	for _, property := range ep.ProviderSpecific {
		if property.Name == "uuid" {
			uuids = strings.Split(property.Value, ",")
			break
		}
	}
	override := opnsense.OpnSenseHostOverride{
		HostName:    hostname,
		Domain:      domain,
		Type:        ep.RecordType,
		TTL:         strconv.FormatInt(int64(ep.RecordTTL), 10),
		Enabled:     "1",
		Description: RecordDescriptionPrefix + ep.Labels["owner"],
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
		uuid := ""
		for _, id := range uuids {
			parts := strings.Split(id, ",")
			fqdn := parts[0]
			uuid = parts[1]
			if fqdn == ep.DNSName {
				override.Uuid = uuid
				break
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		log.Printf("UpdateEntry: Updating host override: %+v\n", override)
		err := override.Update(api.WithContext(ctx))
		if err != nil {
			log.Printf("UpdateEntry: Error updating host override: %v\n", err)
			return err
		}
	}

	return nil
}

func DeleteEntry(api *opnsense.OpnSenseApi, ep *endpoint.Endpoint) error {
	log.Printf("Deleting entry: %s %s %v\n", ep.DNSName, ep.RecordType, ep.Targets)
	var uuids []string
	hostname := strings.Split(ep.DNSName, ".")[0]
	domain := strings.Join(strings.Split(ep.DNSName, ".")[1:], ".")
	for _, property := range ep.ProviderSpecific {
		if property.Name == "uuid" {
			uuids = strings.Split(property.Value, ",")
			break
		}
	}
	override := opnsense.OpnSenseHostOverride{
		HostName:    hostname,
		Domain:      domain,
		Type:        ep.RecordType,
		TTL:         strconv.FormatInt(int64(ep.RecordTTL), 10),
		Enabled:     "1",
		Description: RecordDescriptionPrefix + ep.Labels["owner"],
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

		// find matching uuid for this fqdn
		uuid := ""
		for _, id := range uuids {
			parts := strings.Split(id, ",")
			if len(parts) >= 2 {
				fqdn := parts[0]
				uuid = parts[1]
				if fqdn == ep.DNSName {
					override.Uuid = uuid
					break
				}
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
		defer cancel()
		log.Printf("DeleteEntry: Deleting host override: %+v\n", override)
		err := override.Delete(api.WithContext(ctx))
		if err != nil {
			log.Printf("DeleteEntry: Error deleting host override: %v\n", err)
			return err
		}
	}

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
			// merge targets with deduplication
			targetSet := make(map[string]struct{})
			for _, t := range existingRecord.Targets {
				targetSet[t] = struct{}{}
			}
			for _, t := range record.Targets {
				if _, ok := targetSet[t]; !ok {
					existingRecord.Targets = append(existingRecord.Targets, t)
					targetSet[t] = struct{}{}
				}
			}

			// ensure ProviderSpecific exists
			if len(existingRecord.ProviderSpecific) == 0 && len(record.ProviderSpecific) > 0 {
				existingRecord.ProviderSpecific = append(existingRecord.ProviderSpecific, record.ProviderSpecific[0])
			} else if len(existingRecord.ProviderSpecific) > 0 && len(record.ProviderSpecific) > 0 {
				// append unique provider-specific value fragments (avoid duplicating identical fragments)
				existingVals := strings.Split(existingRecord.ProviderSpecific[0].Value, ";")
				newVals := strings.Split(record.ProviderSpecific[0].Value, ";")
				valSet := make(map[string]struct{})
				for _, v := range existingVals {
					valSet[v] = struct{}{}
				}
				for _, v := range newVals {
					if _, ok := valSet[v]; !ok {
						existingVals = append(existingVals, v)
						valSet[v] = struct{}{}
					}
				}
				existingRecord.ProviderSpecific[0].Value = strings.Join(existingVals, ";")
			}
		} else {
			// store a copy to avoid accidental shared-slice mutations
			copied := *record
			recordMap[key] = &copied
		}
	}

	// Rebuild the slice from the map
	mergedRecords := []*endpoint.Endpoint{}
	for _, record := range recordMap {
		mergedRecords = append(mergedRecords, record)
	}
	return mergedRecords
}
