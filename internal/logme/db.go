package logme

import (
	"errors"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connection() (driver.Conn, error) {
	addr := os.Getenv("DB_ADDR")
	if addr == "" {
		return nil, errors.New("environment variable DB_ADDR required for migrations")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "logme"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: dbName,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Debug: true,
	})

	// Failed to connect
	if err != nil {
		return nil, err
	}

	return conn, nil
}
