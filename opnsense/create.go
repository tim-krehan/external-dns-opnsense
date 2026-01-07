package opnsense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (override *OpnSenseHostOverride) Create(api *OpnSenseApi) error {
	// Check, if override already exists
	foundOverrides, err := SearchHostOverrides(api, fmt.Sprintf("%s %s", override.HostName, override.Domain))
	if err != nil {
		return err
	}
	if len(foundOverrides) > 0 {
		if len(foundOverrides) > 1 {
			matchingOverrides := 0
			for _, o := range foundOverrides {
				if o.HostName == override.HostName && o.Domain == override.Domain {
					matchingOverrides++
				}
			}
			if matchingOverrides == 1 {
				override.Uuid = foundOverrides[0].Uuid
			} else {
				return fmt.Errorf("Read: Multiple host overrides found for %s.%s", override.HostName, override.Domain)
			}
		}
		log.Printf("Create: Host override %s.%s already exists, trying to update\n", override.HostName, override.Domain)
		return override.Update(api)
	} else {
		reqBody := struct {
			Host *OpnSenseHostOverride `json:"host"`
		}{
			Host: override,
		}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		log.Printf("Create: Creating DNS entry [%s] %s => %s (TTL %s)\n", override.Type, override.HostName+"."+override.Domain, (override.Mx + override.Server + override.TxtData), override.TTL)

		resp, err := api.ApiRequest(http.MethodPost, "/unbound/settings/add_host_override/", bytes.NewReader(jsonBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to create DNS entry, status code: %d", resp.StatusCode)
			return ErrFailedToCreate
		}

		// check if the response contains an error message
		var apiResp struct {
			Result string `json:"result"`
			Uuid   string `json:"uuid"`
		}
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		if err != nil {
			return err
		}
		if apiResp.Result != "saved" {
			log.Printf("API returned error: %s", apiResp.Result)
			return ErrApiReturnedError
		}
		return nil
	}
}
