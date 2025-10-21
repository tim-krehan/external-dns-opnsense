package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
)

type OpnSenseConfig struct {
	Ctx       context.Context
	APIKey    string
	APISecret string
	APIHost   string
}

func LoadConfigFromEnv() OpnSenseConfig {
	_ = godotenv.Load()

	apiKey := os.Getenv("OPNSENSE_API_KEY")
	apiSecret := os.Getenv("OPNSENSE_API_SECRET")
	apiHost := os.Getenv("OPNSENSE_API_HOST")

	if apiKey == "" || apiSecret == "" {
		log.Println("Warning: OPNSENSE_API_KEY or OPNSENSE_API_SECRET not set")
	}

	return OpnSenseConfig{
		Ctx:       context.Background(),
		APIKey:    apiKey,
		APISecret: apiSecret,
		APIHost:   apiHost,
	}
}

func (cfg OpnSenseConfig) ApiRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	ctx := cfg.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}

	u, err := url.Parse(cfg.APIHost)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", cfg.APIKey)
	req.Header.Set("X-API-Secret", cfg.APISecret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (cfg OpnSenseConfig) listEntries() []Record {
	// resp, err := cfg.ApiRequest(http.MethodGet, "/zone/record", nil)
	// if err != nil {
	// 	// handle error
	// }
	// defer resp.Body.Close()
	return []Record{
		{
			DNSName:    "example.com",
			Targets:    []string{"1.2.3.4"},
			RecordType: "A",
			TTL:        300,
		},
	}
}

func (cfg OpnSenseConfig) createEntry(rec Record) {
	log.Printf("Create: %+v\n", rec)
}

func (cfg OpnSenseConfig) deleteEntry(rec Record) {
	log.Printf("Delete: %+v\n", rec)
}
