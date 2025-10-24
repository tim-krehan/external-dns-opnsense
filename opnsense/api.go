package opnsense

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadConfigFromEnv loads the API configuration from environment variables.
// It ensures that all required configuration parameters are present.
func LoadConfigFromEnv() *OpnSenseApi {
	_ = godotenv.Load()

	apiKey := os.Getenv("OPNSENSE_API_KEY")
	apiSecret := os.Getenv("OPNSENSE_API_SECRET")
	apiHost := os.Getenv("OPNSENSE_API_HOST")
	envApiTimeout := os.Getenv("OPNSENSE_API_TIMEOUT")
	domainFilter := strings.Split(os.Getenv("DOMAIN_FILTER"), ",")
	ownerId := os.Getenv("EXTERNAL_DNS_OWNER")

	missingConfig := false
	missingConfigParams := []string{}
	if apiKey == "" {
		missingConfig = true
		missingConfigParams = append(missingConfigParams, "OPNSENSE_API_KEY")
	}
	if apiSecret == "" {
		missingConfig = true
		missingConfigParams = append(missingConfigParams, "OPNSENSE_API_SECRET")
	}
	if apiHost == "" {
		missingConfig = true
		missingConfigParams = append(missingConfigParams, "OPNSENSE_API_HOST")
	}
	if missingConfig {
		log.Fatalf("Missing required configuration parameters: %v", missingConfigParams)
	}

	timeout := 30 * time.Second
	if timeoutStr := envApiTimeout; timeoutStr != "" {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		} else {
			log.Printf("Invalid timeout value '%s', using default of 30s", timeoutStr)
		}
	}
	if ownerId == "" {
		log.Printf("EXTERNAL_DNS_OWNER not set, using default value 'default'")
		ownerId = "default"
	}

	log.Printf("Using OpnSense API Host: %s", apiHost)
	log.Printf("With Timeout: %s", timeout.String())

	api := OpnSenseApi{
		Ctx:             context.Background(),
		APIKey:          apiKey,
		APISecret:       apiSecret,
		APIHost:         apiHost,
		ApiTimeout:      timeout,
		DNSDomainFilter: domainFilter,
		OwnerID:         ownerId,
	}
	return &api
}

// ApiRequest performs an HTTP request to the OpnSense API with the specified method, endpoint, and body.
// It handles context management and adds the required authentication headers.
func (api *OpnSenseApi) ApiRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	ctx := api.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, api.ApiTimeout)
		defer cancel()
	}

	u, err := url.Parse(api.APIHost)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api", endpoint)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Add Basic Authentication header
	auth := base64.StdEncoding.EncodeToString([]byte(api.APIKey + ":" + api.APISecret))
	req.Header.Set("Authorization", "Basic "+auth)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: api.ApiTimeout, // Match the client timeout to the context timeout
	}

	// log.Printf("Making %s request to %s", method, u.String())
	resp, err := client.Do(req)
	// log.Printf("Received response with status code: %d and length %d", resp.StatusCode, resp.ContentLength)

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("Request to %s failed due to context error: %v", u.String(), ctx.Err())
		}
		return nil, err
	}

	return resp, nil
}

// WithContext creates a copy of the OpnSenseApi with the specified context.
func (api *OpnSenseApi) WithContext(ctx context.Context) *OpnSenseApi {
	api.Ctx = ctx
	return api
}

func (api *OpnSenseApi) ApplyChanges() error {
	var applyResponse struct {
		Status string `json:"status"`
	}
	ctx, cancel := context.WithTimeout(context.Background(), api.ApiTimeout)
	defer cancel()
	resp, err := api.WithContext(ctx).ApiRequest(http.MethodPost, "/unbound/service/reconfigure", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(&applyResponse)
	if err != nil {
		return err
	}
	if applyResponse.Status != "ok" {
		log.Printf("API returned error during apply changes: %s", applyResponse.Status)
		// return with error
		return ErrFailedToApply
	}

	log.Printf("ApplyChanges: Successfull\n")
	return nil
}
