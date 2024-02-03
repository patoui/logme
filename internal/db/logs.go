package db

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func LogsConnect() (driver.Conn, error) {
	addr := os.Getenv("DB_LOGS_ADDR")
	if addr == "" {
		return nil, errors.New("environment variable DB_LOGS_ADDR required")
	}

	dbName := os.Getenv("DB_LOGS_NAME")
	if dbName == "" {
		dbName = "logs"
	}

	dbDebug := os.Getenv("DB_LOGS_DEBUG")
	isDebug := false
	if dbDebug == "true" {
		isDebug = true
	}

    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{addr},
        Auth: clickhouse.Auth{
            Database: dbName,
            // Username: "default",
            // Password: "<DEFAULT_USER_PASSWORD>",
        },
        Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
        },
  		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
        ClientInfo: clickhouse.ClientInfo{
            Products: []struct {
                Name    string
                Version string
            }{
                {Name: "an-example-go-client", Version: "0.1"},
            },
        },

        Debugf: func(format string, v ...interface{}) {
            fmt.Printf(format, v)
        },
        Debug: isDebug,
        TLS: &tls.Config{
            InsecureSkipVerify: true,
        },
    })

    if err != nil {
        return nil, err
    }

    return conn, nil
}