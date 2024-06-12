package db

import (
	"errors"
	"os"

	"github.com/rueian/valkey-go"
)

func Cache() (client valkey.Client, err error) {
	addr := os.Getenv("CACHE_ADDR")
	if addr == "" {
		return nil, errors.New("environment variable CACHE_ADDR required")
	}

	client, err = valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})

	if err != nil {
		return nil, err
	}

	return client, nil
}
