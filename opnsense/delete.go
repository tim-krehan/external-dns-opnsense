package opnsense

import (
	"fmt"
	"log"
	"net/http"
)

func (override *OpnSenseHostOverride) Delete(api *OpnSenseApi) error {

	// Construct the API endpoint
	endpoint := fmt.Sprintf("/unbound/settings/del_host_override/%s", override.Uuid)

	// Make the DELETE request
	resp, err := api.ApiRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to delete DNS entry with UUID %s, status code: %d\n", override.Uuid, resp.StatusCode)
		return ErrFailedToDelete
	}

	// Log success
	log.Printf("Successfully deleted DNS entry with UUID %s\n", override.Uuid)
	return nil
}
