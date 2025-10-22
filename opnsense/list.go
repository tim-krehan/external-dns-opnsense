package opnsense

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	externaldns "external-dns-opnsense/externaldns"
)

// OpnSenseHostOverride represents a DNS host override entry in OpnSense.
type OpnSenseHostOverride struct {
	Uuid        string `json:"uuid"`
	Enabled     string `json:"enabled"`
	HostName    string `json:"hostname"`
	Domain      string `json:"domain"`
	Type        string `json:"rr"`
	MxPrio      string `json:"mxprio"`
	Mx          string `json:"mx"`
	TTL         string `json:"ttl"`
	Server      string `json:"server"`
	TxtData     string `json:"txtdata"`
	Description string `json:"description"`
}

// listEntries retrieves all DNS host override entries from the OpnSense API.
// It returns a slice of Record objects representing the entries.
func (cfg OpnSenseApi) ListEntries() []externaldns.Record {
	body := map[string]interface{}{
		"current":  1,
		"rowCount": -1,
		"sort":     map[string]interface{}{},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}
	resp, err := cfg.ApiRequest(http.MethodPost, "/unbound/search_host_override/", bytes.NewReader(jsonBody))
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return []externaldns.Record{}
	}
	defer resp.Body.Close()
	var overrides struct {
		Rows     []OpnSenseHostOverride `json:"rows"`
		RowCount int                    `json:"rowCount"`
		Total    int                    `json:"total"`
		Current  int                    `json:"current"`
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response data: %v", err)
		return []externaldns.Record{}
	}
	err = json.Unmarshal(responseData, &overrides)
	if err != nil {
		log.Printf("Error unmarshaling response data: %v", err)
		return []externaldns.Record{}
	}
	records := []externaldns.Record{}
	for _, r := range overrides.Rows {
		ttl, err := strconv.ParseInt(r.TTL, 10, 64)
		if err != nil {
			log.Printf("Error converting TTL to int: %v", err)
			continue
		}
		records = append(records, externaldns.Record{
			DNSName:    r.HostName + "." + r.Domain,
			RecordType: r.Type,
			TTL:        ttl,
			Targets:    []string{},
		})
	}
	return records
}
