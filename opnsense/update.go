package opnsense

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (override *OpnSenseHostOverride) Update(api *OpnSenseApi) error {
	endpoint := fmt.Sprintf("/unbound/settings/set_host_override/%s", override.Uuid)
	reqBody := struct {
		Host *OpnSenseHostOverride `json:"host"`
	}{
		Host: override,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := api.ApiRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to update DNS entry with UUID %s, status code: %d\n", override.Uuid, resp.StatusCode)
		return ErrFailedToUpdate
	}

	// check if the response contains an error message
	var apiResp struct {
		Result string `json:"result"`
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
