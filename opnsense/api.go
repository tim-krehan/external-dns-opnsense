package opnsense

import (
	"context"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
)

// OpnSenseApi represents the API configuration for interacting with the OpnSense API.
type OpnSenseApi struct {
	Ctx        context.Context
	APIKey     string
	APISecret  string
	APIHost    string
	ApiTimeout time.Duration
}

// LoadConfigFromEnv loads the API configuration from environment variables.
// It ensures that all required configuration parameters are present.
func LoadConfigFromEnv() OpnSenseApi {
	_ = godotenv.Load()

	apiKey := os.Getenv("OPNSENSE_API_KEY")
	apiSecret := os.Getenv("OPNSENSE_API_SECRET")
	apiHost := os.Getenv("OPNSENSE_API_HOST")
	envApiTimeout := os.Getenv("OPNSENSE_API_TIMEOUT")

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

	log.Printf("Using OpnSense API Host: %s", apiHost)
	log.Printf("With Timeout: %s", timeout.String())

	return OpnSenseApi{
		Ctx:        context.Background(),
		APIKey:     apiKey,
		APISecret:  apiSecret,
		APIHost:    apiHost,
		ApiTimeout: timeout,
	}
}

// ApiRequest performs an HTTP request to the OpnSense API with the specified method, endpoint, and body.
// It handles context management and adds the required authentication headers.
func (cfg OpnSenseApi) ApiRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	ctx := cfg.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ApiTimeout)
		defer cancel()
	}

	u, err := url.Parse(cfg.APIHost)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "api", endpoint)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// Add Basic Authentication header
	auth := base64.StdEncoding.EncodeToString([]byte(cfg.APIKey + ":" + cfg.APISecret))
	req.Header.Set("Authorization", "Basic "+auth)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: cfg.ApiTimeout, // Match the client timeout to the context timeout
	}

	resp, err := client.Do(req)

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("Request to %s failed due to context error: %v", u.String(), ctx.Err())
		}
		return nil, err
	}

	return resp, nil
}

// WithContext creates a copy of the OpnSenseApi with the specified context.
func (cfg OpnSenseApi) WithContext(ctx context.Context) OpnSenseApi {
	cfg.Ctx = ctx
	return cfg
}
