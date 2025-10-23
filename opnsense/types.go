package opnsense

import (
	"context"
	"time"
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

// OpnSenseApi represents the API configuration for interacting with the OpnSense API.
type OpnSenseApi struct {
	Ctx             context.Context
	APIKey          string
	APISecret       string
	APIHost         string
	ApiTimeout      time.Duration
	DNSDomainFilter []string
}
