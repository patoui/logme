package db

import (
    "errors"
    "os"

    "github.com/meilisearch/meilisearch-go"
)

func Connection() (*meilisearch.Client, error) {
    host := os.Getenv("MEILISEARCH_HOST")
    if host == "" {
        return &meilisearch.Client{}, errors.New("environment variable MEILISEARCH_HOST required for connection")
    }

    apiKey := os.Getenv("MEILISEARCH_API_KEY")
    if apiKey == "" {
        return &meilisearch.Client{}, errors.New("environment variable MEILISEARCH_API_KEY required for connection")
    }

    client := meilisearch.NewClient(meilisearch.ClientConfig{
        Host: host,
        APIKey: apiKey,
    })

    return client, nil
}
