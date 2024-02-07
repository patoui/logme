package models

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/patoui/logme/internal/helpers"
)

const accountIdKey = "accountId"
const layout = "2006-01-02 15:04:05"

type Log struct {
	Uuid      *uuid.UUID         `ch:"uuid" json:"uuid"`
	Name      string             `ch:"name" json:"name"`
	AccountId *uint32            `ch:"account_id" json:"account_id"`
	Content   string             `ch:"content" json:"content"`
	DateTime  helpers.CustomTime `ch:"dt" json:"timestamp"`
}

type CreateLog struct {
	AccountId int                `json:"account_id"`
	Name      string             `json:"name"`
	Content   string             `json:"content"`
	Timestamp helpers.CustomTime `json:"timestamp"`
}

func (log *CreateLog) Create(dbLogs driver.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logsErr := dbLogs.AsyncInsert(
		ctx,
		fmt.Sprintf(
			`INSERT INTO logs (account_id, dt, name, content) VALUES (%d, '%s', '%s', '%s')`,
			log.AccountId,
			log.Timestamp.Time.Format("2006-01-02 15:04:05"),
			log.Name,
			log.Content,
		),
		os.Getenv("DB_LOGS_ASYNC_WAIT") == "true",
	)

	if logsErr != nil {
		return logsErr
	}

	return logsErr
}

func List(dbLogs driver.Conn, accountId int, query string) ([]Log, error) {
	var logs []Log
	rows, logsErr := dbLogs.Query(
		context.Background(),
		fmt.Sprintf("SELECT * FROM logs WHERE account_id = %d ORDER BY dt DESC", accountId),
	)

	if logsErr != nil {
		return nil, logsErr
	}

	for rows.Next() {
		var currentLog struct {
			Uuid      *uuid.UUID `ch:"uuid"`
			Name      string     `ch:"name"`
			AccountId *uint32    `ch:"account_id"`
			Content   string     `ch:"content"`
			DateTime  time.Time  `ch:"dt"`
		}
		if scanErr := rows.ScanStruct(&currentLog); scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, Log{
			Uuid:      currentLog.Uuid,
			Name:      currentLog.Name,
			AccountId: currentLog.AccountId,
			Content:   currentLog.Content,
			DateTime:  helpers.CustomTime{Time: currentLog.DateTime},
		})
	}

	return logs, nil
}
