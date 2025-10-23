package opnsense

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func SearchHostOverrides(api *OpnSenseApi, searchPhrase string) ([]OpnSenseHostOverride, error) {
	body := map[string]interface{}{
		"current":      1,
		"rowCount":     -1,
		"sort":         map[string]interface{}{},
		"searchPhrase": searchPhrase,
	}
	endpoint := "/unbound/settings/search_host_override/"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := api.ApiRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResponse struct {
		Rows     []OpnSenseHostOverride `json:"rows"`
		RowCount int                    `json:"rowCount"`
		Total    int                    `json:"total"`
		Current  int                    `json:"current"`
	}
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}
	return searchResponse.Rows, nil
}
