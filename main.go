package main

import (
	opnsense "external-dns-opnsense/opnsense"
	"log"
	"net/http"
)

// Global variable to hold the OpnSense configuration
var api opnsense.OpnSenseApi

func main() {
	// Register HTTP handlers for the webhook server
	http.HandleFunc("/", negotiateHandler)                      // Handles negotiation requests
	http.HandleFunc("/records", recordsHandler)                 // Handles requests to retrieve or edit DNS records
	http.HandleFunc("/adjustendpoints", adjustendpointsHandler) // Handles requests to adjust DNS endpoints

	// Load the OpnSense configuration from environment variables
	api = opnsense.LoadConfigFromEnv()

	// Start the webhook server on port 8888
	log.Println("Webhook server listening on :8888")
	log.Fatal(http.ListenAndServe("localhost:8888", nil)) // Log fatal errors if the server fails to start
}
