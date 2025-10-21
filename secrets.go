package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type OpnSenseSecrets struct {
    APIKey    string
    APISecret string
}

func LoadOpnSenseSecrets() OpnSenseSecrets {
    // Load .env file if present
    _ = godotenv.Load()

    apiKey := os.Getenv("OPNSENSE_API_KEY")
    apiSecret := os.Getenv("OPNSENSE_API_SECRET")

    if apiKey == "" || apiSecret == "" {
        log.Println("Warning: OPNSENSE_API_KEY or OPNSENSE_API_SECRET not set")
    }

    return OpnSenseSecrets{
        APIKey:    apiKey,
        APISecret: apiSecret,
    }
}
