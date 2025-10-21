package main

import (
	"log"
	"net/http"
)

var cfg OpnSenseConfig

func main() {
	http.HandleFunc("/records", recordsHandler)
	http.HandleFunc("/applychanges", applyChangesHandler)

	cfg = LoadConfigFromEnv()

	log.Println("Webhook server listening on :30000")
	log.Fatal(http.ListenAndServe(":30000", nil))
}
